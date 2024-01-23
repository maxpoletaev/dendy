package netplay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/maxpoletaev/dendy/console"
	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/binario"
	"github.com/maxpoletaev/dendy/internal/ringbuf"
	"github.com/maxpoletaev/dendy/ui"
)

type Checkpoint struct {
	Frame       uint32
	State       []byte
	Crc32       uint32
	LocalInput  uint8
	RemoteInput uint8
}

// Game is a network play state manager. It keeps track of the inputs from both
// players and makes sure their state is synchronized.
type Game struct {
	frame uint32 // emulated frame counter, can go back on checkpoint rollback
	bus   *console.Bus
	gen   uint32
	tick  uint64

	checkpoint  *Checkpoint
	catching    *Checkpoint
	current     *Checkpoint
	catchingPos int

	localInput      *ringbuf.Buffer[uint8]
	remoteInput     *ringbuf.Buffer[uint8]
	speculatedInput *ringbuf.Buffer[uint8]
	lastRemoteInput uint8

	rtt         time.Duration // round trip time
	remoteFrame uint32
	frameDrift  int
	sleepFrames int

	playDurationAvg time.Duration // how long it takes to emulate a frame
	playDurBuffer   *ringbuf.Buffer[time.Duration]

	audio          *ui.AudioOut
	audioBuffer    []float32
	audioBufferPos int

	LocalJoy      *input.Joystick
	RemoteJoy     *input.Joystick
	DisasmEnabled bool
}

func NewGame(bus *console.Bus, audio *ui.AudioOut) *Game {
	return &Game{
		bus:         bus,
		current:     &Checkpoint{},
		checkpoint:  &Checkpoint{},
		catching:    &Checkpoint{},
		audio:       audio,
		audioBuffer: make([]float32, consts.AudioBufferSize),
	}
}

func (g *Game) Init(cp *Checkpoint) {
	g.lastRemoteInput = 0
	g.sleepFrames = 0
	g.remoteFrame = 0
	g.catchingPos = 0
	g.frameDrift = 0
	g.frame = 0

	g.localInput = ringbuf.New[uint8](300)
	g.remoteInput = ringbuf.New[uint8](300)
	g.speculatedInput = ringbuf.New[uint8](300)
	g.playDurBuffer = ringbuf.New[time.Duration](10)

	if cp != nil {
		g.checkpoint = cp
		g.rollback(g.checkpoint)
	} else {
		g.save(g.checkpoint)
	}

	g.gen++ // messages in-flight are no longer valid
}

func (g *Game) Reset() {
	g.bus.Reset()
}

func (g *Game) Sleep(n int) {
	g.sleepFrames = n
}

func (g *Game) SetRTT(t time.Duration) {
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
	if g.frame%100 != 0 {
		return // no need to report every frame
	}

	g.playDurBuffer.PushBackEvict(delta)

	var sum time.Duration
	for i := 0; i < g.playDurBuffer.Len(); i++ {
		sum += g.playDurBuffer.At(i)
	}

	g.playDurationAvg = sum / time.Duration(g.playDurBuffer.Len())
	//log.Printf("[INFO] frame duration: %s", g.playDurationAvg)
}

func (g *Game) playFrame() {
	for {
		g.bus.Tick()
		g.tick++

		if g.tick%consts.TicksPerSample == 0 {
			g.audioBuffer[g.audioBufferPos] = g.bus.APU.Output()
			g.audioBufferPos++

			if g.audioBufferPos == len(g.audioBuffer) {
				g.audio.WaitStreamProcessed()
				g.audio.UpdateStream(g.audioBuffer)
				g.audioBufferPos = 0
			}
		}

		if g.bus.FrameComplete() {
			g.frame++
			break
		}
	}

	// Overflow will happen after ~2 years of continuous play :)
	// Don't think it's a problem though.
	if g.frame == 0 {
		panic("frame counter overflow")
	}
}

func (g *Game) playFrameFast() {
	start := time.Now()

	g.bus.PPU.EnableFastForward()
	defer g.bus.PPU.DisableFastForward()

	for {
		g.bus.Tick()

		if g.bus.FrameComplete() {
			g.frame++
			break
		}
	}

	// Overflow will happen after ~2 years of continuous play :)
	// Don't think it's a problem though.
	if g.frame == 0 {
		panic("frame counter overflow")
	}

	g.reportFrameDuration(time.Since(start))
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

	if g.frameDrift > frameDriftLimit {
		g.sleepFrames = g.frameDrift
		return
	}

	g.processDelayedInput(startTime)
	g.playFrame()
}

func (g *Game) save(cp *Checkpoint) {
	buf := bytes.NewBuffer(cp.State[:0])
	writer := binario.NewWriter(buf, binary.LittleEndian)

	if err := g.bus.SaveState(writer); err != nil {
		panic(fmt.Errorf("failed create checkpoint: %w", err))
	}

	cp.Frame = g.frame
	cp.State = buf.Bytes() // re-assign in case it was re-allocated

	cp.LocalInput = g.LocalJoy.Buttons()
	cp.RemoteInput = g.RemoteJoy.Buttons()
}

func (g *Game) rollback(cp *Checkpoint) {
	buf := bytes.NewBuffer(cp.State)
	reader := binario.NewReader(buf, binary.LittleEndian)

	if err := g.bus.LoadState(reader); err != nil {
		panic(fmt.Errorf("failed to restore checkpoint: %w", err))
	}

	g.frame = cp.Frame
	g.LocalJoy.SetButtons(cp.LocalInput)
	g.RemoteJoy.SetButtons(cp.RemoteInput)
}

// HandleLocalInput adds records and applies the input from the local player.
// Since the remote player is behind, it assumes that it just keeps pressing
// the same buttons until it catches up.
func (g *Game) HandleLocalInput(buttons uint8) {
	g.LocalJoy.SetButtons(buttons)
	g.RemoteJoy.SetButtons(g.lastRemoteInput)

	g.localInput.PushBack(buttons)
	g.speculatedInput.PushBack(g.lastRemoteInput)
}

// HandleRemoteInput adds the input from the remote player.
func (g *Game) HandleRemoteInput(buttons uint8, frame uint32) {
	g.remoteInput.PushBack(buttons)
	g.lastRemoteInput = buttons

	if g.rtt > 0 {
		// Try to guess the frame of the remote player is on adjusted for latency.
		g.remoteFrame = frame + uint32(g.rtt/2/frameDuration) + 1
		g.frameDrift = int(g.frame) - int(g.remoteFrame)
	}
}

func (g *Game) replayLocalInput(startTime time.Time, endFrame uint32, inputPos int) {
	for f := g.frame; f < endFrame; f++ {
		timeLeft := frameDuration - time.Since(startTime)

		if timeLeft < g.playDurationAvg*3 { // TODO: why 3?
			g.save(g.catching)
			g.rollback(g.current)
			g.catchingPos = inputPos

			return
		}

		g.RemoteJoy.SetButtons(g.speculatedInput.At(inputPos))
		g.LocalJoy.SetButtons(g.localInput.At(inputPos))

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
	endFrame := g.frame
	g.rollback(g.checkpoint)

	// Ensure we are always back to where we started.
	defer func() {
		if g.frame != endFrame {
			panic(fmt.Errorf("frame advanced from %d to %d", endFrame, g.frame))
		}
	}()

	// Enable CPU disassembly if requested. We do it only for frames where we have
	// both local and remote inputs, so that we can compare.
	if g.DisasmEnabled {
		g.bus.DisasmEnabled = true
	}

	// Replay the inputs until the local and remote emulators are in sync.
	for i := 0; i < inputSize; i++ {
		timeLeft := frameDuration - time.Since(startTime)

		if timeLeft < g.playDurationAvg*3 {
			g.save(g.checkpoint)
			g.rollback(g.current)
			g.dropInputs(i)

			return
		}

		g.LocalJoy.SetButtons(g.localInput.At(i))
		g.RemoteJoy.SetButtons(g.remoteInput.At(i))

		g.playFrameFast()
	}

	// Disable CPU disassembly, since from now on we have only the predicted input
	// from the remote player, so this part will be rolling back eventually.
	if g.DisasmEnabled {
		g.bus.DisasmEnabled = false
	}

	// Rebuild the speculated input from this point as the last remote input could have changed.
	for i := inputSize; i < g.localInput.Len(); i++ {
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
