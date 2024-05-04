package netplay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/binario"
	"github.com/maxpoletaev/dendy/internal/ringbuf"
	"github.com/maxpoletaev/dendy/system"
	"github.com/maxpoletaev/dendy/ui"
)

type Checkpoint struct {
	State       *bytes.Buffer
	Reader      *binario.Reader
	Writer      *binario.Writer
	Frame       uint32
	Crc32       uint32
	LocalInput  uint8
	RemoteInput uint8
	RolledBack  bool
}

func newCheckpoint() *Checkpoint {
	buf := bytes.NewBuffer(nil)

	// NOTE: checkpoint can only be rolled back once (because bytes.Buffer is not
	// seekable, and can only be read once). This works for now but may require a
	// seekable bytes.Buffer implementation in the future. We need to keep the buffer
	// global to avoid heap allocations on every frame.
	return &Checkpoint{
		State:  buf,
		Reader: binario.NewReader(buf, binary.LittleEndian),
		Writer: binario.NewWriter(buf, binary.LittleEndian),
	}
}

// Game is a network play state manager. It keeps track of the inputs from both
// players and makes sure their state is synchronized.
type Game struct {
	frame uint32 // emulated frame counter, can go back on checkpoint rollback
	nes   *system.System
	gen   uint32
	tick  uint64

	checkpoint  *Checkpoint
	catching    *Checkpoint
	current     *Checkpoint
	catchingPos int

	localInput      *ringbuf.Buffer[uint8]
	remoteInput     *ringbuf.Buffer[uint8]
	speculatedInput *ringbuf.Buffer[uint8]

	rtt                 time.Duration // round trip time
	frameDrift          int
	lastRemoteInput     uint8
	sleepFrames         uint32
	frameDuration       time.Duration // how long it takes to emulate a frame
	frameDurationWindow *ringbuf.Buffer[time.Duration]
	audio               *ui.AudioOut
	audioBuffer         []float32
	audioBufferPos      int
	debugWriter         io.StringWriter
	localJoy            *input.Joystick
	remoteJoy           *input.Joystick
}

func NewGame(nes *system.System, audio *ui.AudioOut, localJoy, remoteJoy *input.Joystick) *Game {
	return &Game{
		nes:         nes,
		current:     newCheckpoint(),
		checkpoint:  newCheckpoint(),
		catching:    newCheckpoint(),
		audio:       audio,
		audioBuffer: make([]float32, consts.AudioBufferSize),
		localJoy:    localJoy,
		remoteJoy:   remoteJoy,
	}
}

func (g *Game) Init(cp *Checkpoint) {
	g.lastRemoteInput = 0
	g.sleepFrames = 0
	g.catchingPos = 0
	g.frame = 0

	g.localInput = ringbuf.New[uint8](512)
	g.remoteInput = ringbuf.New[uint8](512)
	g.speculatedInput = ringbuf.New[uint8](512)
	g.frameDurationWindow = ringbuf.New[time.Duration](16)

	if cp != nil {
		g.checkpoint = cp
	} else {
		g.save(g.checkpoint)
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
	g.rtt = t
}

// Checkpoint returns the current checkpoint where both players are in sync. The
// returned value should not be modified and is only valid within the current frame.
func (g *Game) Checkpoint() *Checkpoint {
	return g.checkpoint
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

func (g *Game) reportFrameDuration(delta time.Duration) {
	g.frameDurationWindow.PushBackEvict(delta)

	var sum time.Duration
	for i := 0; i < g.frameDurationWindow.Len(); i++ {
		sum += g.frameDurationWindow.At(i)
	}

	g.frameDuration = sum / time.Duration(g.frameDurationWindow.Len())
	//log.Printf("[INFO] frame duration: %s", g.playDurationAvg)
}

func (g *Game) playFrame() {
	start := time.Now()

	for {
		g.nes.Tick()
		g.tick++

		if g.tick%consts.TicksPerSample == 0 {
			if g.audioBufferPos < len(g.audioBuffer) {
				g.audioBuffer[g.audioBufferPos] = g.nes.AudioSample()
				g.audioBufferPos++
			}

			if g.audioBufferPos == len(g.audioBuffer) && g.audio.IsStreamProcessed() {
				g.audio.UpdateStream(g.audioBuffer)
				g.audioBufferPos = 0
			}
		}

		if g.nes.FrameReady() {
			g.frame++
			break
		}
	}

	// Overflow will happen after ~2 years of continuous play :)
	// Don't think it's a problem though.
	if g.frame == 0 {
		panic("frame counter overflow")
	}

	if g.frame%10 == 0 {
		g.reportFrameDuration(time.Since(start))
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
	g.speculatedInput.TruncFront(n)
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

func (g *Game) FrameDrift() int {
	return g.frameDrift
}

func (g *Game) save(cp *Checkpoint) {
	cp.State.Reset()

	if err := g.nes.SaveState(cp.Writer); err != nil {
		panic(fmt.Errorf("failed create checkpoint: %w", err))
	}

	cp.Frame = g.frame
	cp.RolledBack = false
	cp.LocalInput = g.localJoy.Buttons()
	cp.RemoteInput = g.remoteJoy.Buttons()
}

func (g *Game) rollback(cp *Checkpoint) {
	if cp.RolledBack {
		panic("checkpoint already rolled back")
	}

	if err := g.nes.LoadState(cp.Reader); err != nil {
		panic(fmt.Errorf("failed to restore checkpoint: %w", err))
	}

	g.frame = cp.Frame
	cp.RolledBack = true
	g.localJoy.SetButtons(cp.LocalInput)
	g.remoteJoy.SetButtons(cp.RemoteInput)
}

// HandleLocalInput adds records and applies the input from the local player.
// Since the remote player is behind, it assumes that it just keeps pressing
// the same buttons until it catches up.
func (g *Game) HandleLocalInput(buttons uint8) {
	g.localJoy.SetButtons(buttons)
	g.remoteJoy.SetButtons(g.lastRemoteInput)

	g.localInput.PushBack(buttons)
	g.speculatedInput.PushBack(g.lastRemoteInput)
}

// HandleRemoteInput adds the input from the remote player.
func (g *Game) HandleRemoteInput(buttons uint8, frame uint32) {
	g.remoteInput.PushBack(buttons)
	g.lastRemoteInput = buttons

	if g.rtt > 0 {
		localFrame := g.frame
		latencyFrames := uint32(g.rtt / 2 / consts.FrameDuration)
		remoteFrame := frame + latencyFrames // just a good guess

		if localFrame < remoteFrame {
			g.frameDrift = -int(remoteFrame - localFrame)
		} else {
			g.frameDrift = int(localFrame - remoteFrame)
		}
	}
}

func (g *Game) replayLocalInput(startTime time.Time, endFrame uint32, inputPos int) {
	for f := g.frame; f < endFrame; f++ {
		timeLeft := consts.FrameDuration - time.Since(startTime)

		if timeLeft < g.frameDuration*2 {
			g.save(g.catching)
			g.rollback(g.current)
			g.catchingPos = inputPos

			return
		}

		g.remoteJoy.SetButtons(g.speculatedInput.At(inputPos))
		g.localJoy.SetButtons(g.localInput.At(inputPos))
		g.playFrameFast()

		inputPos++
	}

	g.catchingPos = 0
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
	if g.catchingPos != 0 {
		g.save(g.current)
		g.rollback(g.catching)
		g.replayLocalInput(startTime, g.current.Frame, g.catchingPos)

		// If we are still behind, we will try again next frame.
		// TODO: detect when we are not making any progress and give up.
		if g.catchingPos != 0 {
			return
		}
	}

	inputSize := min(g.localInput.Len(), g.remoteInput.Len(), int(g.frame-g.checkpoint.Frame))
	if inputSize == 0 {
		return
	}

	// Preserve the state before the rollback. We will restore it
	// in case we do not have enough time to catch up during this frame.
	g.save(g.current)

	// Rollback to the last known synchronized state.
	g.rollback(g.checkpoint)

	// Ensure we are always back to where we started.
	defer func() {
		if g.frame != endFrame {
			panic(fmt.Errorf("frame advanced from %d to %d", endFrame, g.frame))
		}
	}()

	// Enable CPU disassembly if requested. We do it only for frames where we have
	// both local and remote inputs, so that we can compare.
	if g.debugWriter != nil {
		g.nes.SetDebugOutput(g.debugWriter)
	}

	// Replay the inputs until the local and remote emulators are in sync.
	for i := 0; i < inputSize; i++ {
		timeLeft := consts.FrameDuration - time.Since(startTime)
		if timeLeft < g.frameDuration*2 {
			g.save(g.checkpoint)
			g.rollback(g.current)
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
		g.nes.SetDebugOutput(nil)
	}

	// Rebuild the speculated input from this point as the last remote input could have changed.
	for i := inputSize; i < g.speculatedInput.Len(); i++ {
		g.speculatedInput.Set(i, g.remoteInput.At(inputSize-1))
	}

	// This is the last state where both emulators are in sync. Create a new checkpoint,
	// so we can rewind to this state later, and remove the inputs that we have
	// already processed.
	g.save(g.checkpoint)
	g.dropInputs(inputSize)

	// Replay the rest of the local inputs and use speculated values for the remote.
	g.replayLocalInput(startTime, endFrame, 0)
}

func (g *Game) SetDebugOutput(w io.StringWriter) {
	g.debugWriter = w
}
