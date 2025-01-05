package netplay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"time"

	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/binario"
	"github.com/maxpoletaev/dendy/internal/ringbuf"
	"github.com/maxpoletaev/dendy/system"
	"github.com/maxpoletaev/dendy/ui"
)

type checkpoint struct {
	state       *bytes.Buffer
	reader      *binario.Reader
	writer      *binario.Writer
	frame       uint32
	crc32       uint32
	localInput  uint8
	remoteInput uint8
	rolledBack  bool
}

func newCheckpoint() *checkpoint {
	buf := bytes.NewBuffer(nil)

	// NOTE: checkpoint can only be rolled back once (because bytes.Buffer is not
	// seekable, and can only be read once). This works for now but may require a
	// seekable bytes.Buffer implementation in the future. We need to keep the buffer
	// global to avoid heap allocations on every frame.
	return &checkpoint{
		state:  buf,
		reader: binario.NewReader(buf, binary.LittleEndian),
		writer: binario.NewWriter(buf, binary.LittleEndian),
	}
}

// Game is a network play state manager. It keeps track of the inputs from both
// players and makes sure their state is synchronized.
type Game struct {
	nes   *system.System
	frame uint32
	gen   uint32
	tick  uint64

	syncState       *checkpoint // last known synchronized state
	headState       *checkpoint // latest local state (before rollback)
	catchupState    *checkpoint // state in-between sync and head states while catching up
	catchupInputPos int         // position in the local input buffer while catching up

	localInput           *ringbuf.Buffer[uint8]
	remoteInput          *ringbuf.Buffer[uint8]
	predictedRemoteInput *ringbuf.Buffer[uint8]
	lastRemoteInput      uint8
	localJoy             *input.Joystick
	remoteJoy            *input.Joystick

	frameEmulationTime time.Duration
	roundTripTime      time.Duration
	driftFrames        int
	sleepFrames        uint32
	audioOut           *ui.AudioOut
	audioBuffer        []float32
	audioBufferPos     int
	debugWriter        io.StringWriter
}

func NewGame(nes *system.System, audio *ui.AudioOut, localJoy, remoteJoy *input.Joystick) *Game {
	return &Game{
		nes:          nes,
		headState:    newCheckpoint(),
		syncState:    newCheckpoint(),
		catchupState: newCheckpoint(),
		audioOut:     audio,
		audioBuffer:  make([]float32, consts.AudioBufferSize),
		localJoy:     localJoy,
		remoteJoy:    remoteJoy,
	}
}

func (g *Game) Init(cp *checkpoint) {
	g.lastRemoteInput = 0
	g.catchupInputPos = 0
	g.sleepFrames = 0
	g.frame = 0

	g.localInput = ringbuf.New[uint8](512)
	g.remoteInput = ringbuf.New[uint8](512)
	g.predictedRemoteInput = ringbuf.New[uint8](512)

	if cp != nil {
		g.syncState = cp
	} else {
		g.save(g.syncState)
	}

	g.gen++ // messages in-flight are no longer valid
}

func (g *Game) Reset() {
	g.nes.Reset()
}

func (g *Game) SleepFrames(n uint32) {
	g.sleepFrames = n
}

func (g *Game) SetRoundTripTime(t time.Duration) {
	g.roundTripTime = t
}

// Frame returns the current frame number.
func (g *Game) Frame() uint32 {
	return g.frame
}

// Gen returns the current generation number. It is incremented every
// time the game is reset.
func (g *Game) Gen() uint32 {
	return g.gen
}

// Sleeping returns true if the game is currently sleeping to let the remote player catch up.
func (g *Game) Sleeping() bool {
	return g.sleepFrames > 0
}

func (g *Game) playFrame() {
	start := time.Now()

	for {
		g.nes.Tick()
		g.tick++

		if g.tick%consts.TicksPerAudioSample == 0 {
			if g.audioBufferPos < len(g.audioBuffer) {
				g.audioBuffer[g.audioBufferPos] = g.nes.AudioSample()
				g.audioBufferPos++
			}

			if g.audioBufferPos == len(g.audioBuffer) && g.audioOut.IsStreamProcessed() {
				g.audioOut.UpdateStream(g.audioBuffer)
				g.audioBufferPos = 0
			}
		}

		if g.nes.FrameReady() {
			g.frame++
			break
		}
	}

	g.frameEmulationTime = time.Since(start)

	// Overflow will happen after ~2 years of continuous play :)
	// Don't think it's a problem though.
	if g.frame == 0 {
		panic("frame counter overflow")
	}
}

func (g *Game) playFrameFast() {
	g.nes.SetFastForward(true)
	defer g.nes.SetFastForward(false)

	for {
		g.nes.Tick()
		if g.nes.FrameReady() {
			g.frame++
			break
		}
	}

	if g.frame == 0 {
		panic("frame counter overflow")
	}
}

func (g *Game) dropInputs(n int) {
	g.localInput.TruncFront(n)
	g.remoteInput.TruncFront(n)
	g.predictedRemoteInput.TruncFront(n)
}

// RunFrame runs a single frame of the game.
func (g *Game) RunFrame(startTime time.Time) {
	if g.sleepFrames > 0 {
		g.sleepFrames--
		return
	}

	g.processDelayedInput(startTime)
	g.playFrame()
}

func (g *Game) DriftFrames() int {
	return g.driftFrames
}

func (g *Game) save(cp *checkpoint) {
	cp.state.Reset()

	if err := g.nes.SaveState(cp.writer); err != nil {
		panic(fmt.Errorf("failed create checkpoint: %w", err))
	}

	cp.frame = g.frame
	cp.rolledBack = false
	cp.localInput = g.localJoy.Buttons()
	cp.remoteInput = g.remoteJoy.Buttons()
	cp.crc32 = crc32.ChecksumIEEE(cp.state.Bytes())
}

func (g *Game) rollback(cp *checkpoint) {
	if cp.rolledBack {
		panic("checkpoint already rolled back")
	}

	if err := g.nes.LoadState(cp.reader); err != nil {
		panic(fmt.Errorf("failed to restore checkpoint: %w", err))
	}

	g.frame = cp.frame
	g.localJoy.SetButtons(cp.localInput)
	g.remoteJoy.SetButtons(cp.remoteInput)
	cp.rolledBack = true
}

// HandleLocalInput adds records and applies the input from the local player.
// Since the remote player is behind, it assumes that it just keeps pressing
// the same buttons until it catches up.
func (g *Game) HandleLocalInput(buttons uint8) {
	g.localJoy.SetButtons(buttons)
	g.remoteJoy.SetButtons(g.lastRemoteInput)

	g.localInput.PushBack(buttons)
	g.predictedRemoteInput.PushBack(g.lastRemoteInput)
}

// HandleRemoteInput adds the input from the remote player.
func (g *Game) HandleRemoteInput(buttons uint8, frame uint32) {
	g.remoteInput.PushBack(buttons)
	g.lastRemoteInput = buttons

	if g.roundTripTime > 0 {
		localFrame := g.frame
		latencyFrames := uint32(g.roundTripTime / 2 / consts.FrameDuration)
		remoteFrame := frame + latencyFrames // just a good guess

		if localFrame < remoteFrame {
			g.driftFrames = -int(remoteFrame - localFrame)
		} else {
			g.driftFrames = int(localFrame - remoteFrame)
		}
	}
}

func (g *Game) replayLocalInput(startTime time.Time, endFrame uint32, inputPos int) {
	for f := g.frame; f < endFrame; f++ {
		remainingTime := consts.FrameDuration - time.Since(startTime)

		if remainingTime < g.frameEmulationTime {
			g.save(g.catchupState)
			g.rollback(g.headState)
			g.catchupInputPos = inputPos
			return
		}

		g.remoteJoy.SetButtons(g.predictedRemoteInput.At(inputPos))
		g.localJoy.SetButtons(g.localInput.At(inputPos))
		g.playFrameFast()

		inputPos++
	}

	g.catchupInputPos = 0
}

// processDelayedInput applies the input from the remote player to the local
// emulator when it is available. This is where all the magic happens. The remote
// player is usually a few frames behind the local emulator state. The emulator
// is reset to the last checkpoint and then both local and remote inputs are
// replayed until they catch up to the current frame.
func (g *Game) processDelayedInput(startTime time.Time) {
	endFrame := g.frame

	// Continue catching up to our current frame, as we didn't have enough time
	// during the last frame. As soon as the emulation is faster than the real
	// time, we will catch up eventually.
	if g.catchupInputPos != 0 {
		g.save(g.headState)
		g.rollback(g.catchupState)
		g.replayLocalInput(startTime, g.headState.frame, g.catchupInputPos)

		// If we are still behind, we will try again next frame.
		// TODO: detect when we are not making any progress and give up.
		if g.catchupInputPos != 0 {
			return
		}
	}

	numInputs := min(g.localInput.Len(), g.remoteInput.Len(), int(g.frame-g.syncState.frame))
	if numInputs == 0 {
		return
	}

	// Preserve the state before the rollback. We will restore it
	// in case we do not have enough time to catch up during this frame.
	g.save(g.headState)

	// Rollback to the last known synchronized state.
	g.rollback(g.syncState)

	// Ensure we are always back to where we started.
	defer func() {
		if g.frame != endFrame {
			panic(fmt.Errorf("frame advanced from %d to %d", endFrame, g.frame))
		}
	}()

	// Enable CPU disassembly if requested. We do it only for frames where we have
	// both local and remote inputs, so that we can compare.
	if g.debugWriter != nil {
		g.nes.SetDebugWriter(g.debugWriter)
	}

	// Replay the inputs until the local and remote emulators are in sync.
	for i := 0; i < numInputs; i++ {
		remainingTime := consts.FrameDuration - time.Since(startTime)

		// We only have 16ms to replay all frames. Going over this limit will create a
		// noticeable stutter and sound glitches. When we are close to the limit, save
		// the progress and jump back to the last known unsynchronized state.
		if remainingTime < g.frameEmulationTime {
			g.save(g.syncState)
			g.rollback(g.headState)
			g.dropInputs(i)
			return
		}

		g.localJoy.SetButtons(g.localInput.At(i))
		g.remoteJoy.SetButtons(g.remoteInput.At(i))

		g.playFrameFast()
	}

	// Disable CPU disassembly, since from now on we have only the predicted input
	// from the remote player, so this part will be rolling back eventually.
	if g.debugWriter != nil {
		g.nes.SetDebugWriter(nil)
	}

	// Rebuild the speculated input from this point as the last remote input could have changed.
	for i := numInputs; i < g.predictedRemoteInput.Len(); i++ {
		g.predictedRemoteInput.Set(i, g.remoteInput.At(numInputs-1))
	}

	// This is the last state where both emulators are in sync. Create a new checkpoint,
	// so we can rewind to this state later, and remove the inputs that we have
	// already processed.
	g.save(g.syncState)
	g.dropInputs(numInputs)

	// Replay the rest of the local inputs and use speculated values for the remote.
	g.replayLocalInput(startTime, endFrame, 0)
}

func (g *Game) SetDebugOutput(w io.StringWriter) {
	g.debugWriter = w
}
