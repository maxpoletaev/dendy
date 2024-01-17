package netplay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"time"

	"github.com/maxpoletaev/dendy/console"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/binario"
	"github.com/maxpoletaev/dendy/internal/ringbuf"
)

type Checkpoint struct {
	Frame uint32
	State []byte
	Crc32 uint32
}

type CheckFrame struct {
	Frame uint32
	Crc32 uint32
}

type PlayerInput struct {
	Frame   uint32
	Buttons uint8
}

// Game is a network play state manager. It keeps track of the inputs from both
// players and makes sure their state is synchronized.
type Game struct {
	frame uint32
	bus   *console.Bus
	cp    *Checkpoint
	gen   uint32

	localInput      *ringbuf.Buffer[uint8]
	remoteInput     *ringbuf.Buffer[uint8]
	speculatedInput *ringbuf.Buffer[uint8]
	lastRemoteInput uint8

	LocalJoy      *input.Joystick
	RemoteJoy     *input.Joystick
	DisasmEnabled bool
	BufferSize    int
}

func NewGame(bus *console.Bus) *Game {
	return &Game{
		bus: bus,
		cp:  &Checkpoint{},
	}
}

// Reset resets the emulator state to the given checkpoint. If cp is nil, the
// emulator is reset to the initial state.
func (g *Game) Reset(cp *Checkpoint) {
	g.frame = 0
	g.gen++

	g.localInput = ringbuf.New[uint8](300)
	g.remoteInput = ringbuf.New[uint8](300)
	g.speculatedInput = ringbuf.New[uint8](300)

	if cp != nil {
		g.cp = cp
		g.restoreCheckpoint()

		return
	}

	g.bus.Reset()
	g.createCheckpoint()
}

// Checkpoint returns the current checkpoint where both players are in sync. The
// returned value should not be modified and is only valid within the current frame.
func (g *Game) Checkpoint() *Checkpoint {
	return g.cp
}

// Frame returns the current frame number.
func (g *Game) Frame() uint32 {
	return g.frame
}

// Checksum returns the checksum of the current emulator state. It can be used
// to compare two game states.
func (g *Game) Checksum() uint32 {
	h := crc32.NewIEEE()
	writer := binario.NewWriter(h, binary.LittleEndian)

	if err := g.bus.SaveState(writer); err != nil {
		panic(fmt.Errorf("failed create checksum: %w", err))
	}

	return h.Sum32()
}

// Gen returns the current generation number. It is incremented every
// time the game is reset.
func (g *Game) Gen() uint32 {
	return g.gen
}

func (g *Game) playFrame() {
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
}

// RunFrame runs a single frame of the game.
func (g *Game) RunFrame() {
	g.applyRemoteInput()
	g.playFrame()
}

// Sleep pauses the execution for the given number of frames.
func (g *Game) Sleep(d int) {
	time.Sleep(time.Duration(d) * frameDuration)
}

func (g *Game) createCheckpoint() {
	buf := bytes.NewBuffer(g.cp.State[:0])
	writer := binario.NewWriter(buf, binary.LittleEndian)

	if err := g.bus.SaveState(writer); err != nil {
		panic(fmt.Errorf("failed create checkpoint: %w", err))
	}

	g.cp.Frame = g.frame
	g.cp.State = buf.Bytes() // re-assign in case it was re-allocated
}

func (g *Game) restoreCheckpoint() {
	buf := bytes.NewBuffer(g.cp.State)
	reader := binario.NewReader(buf, binary.LittleEndian)

	if err := g.bus.LoadState(reader); err != nil {
		panic(fmt.Errorf("failed to restore checkpoint: %w", err))
	}

	g.frame = g.cp.Frame
}

// HandleLocalInput adds records and applies the input from the local player.
// Since the remote player is behind, it assumes that it just keeps pressing
// the same buttons until it catches up. This is not always true, but it's
// good approximation for most games.
func (g *Game) HandleLocalInput(buttons uint8) {
	g.LocalJoy.SetButtons(buttons)
	g.RemoteJoy.SetButtons(g.lastRemoteInput)

	g.localInput.PushBack(buttons)
	g.speculatedInput.PushBack(g.lastRemoteInput)
}

// HandleRemoteInput adds the input from the remote player.
func (g *Game) HandleRemoteInput(buttons uint8) {
	g.remoteInput.PushBack(buttons)
	g.lastRemoteInput = buttons
}

// applyRemoteInput applies the input from the remote player to the local
// emulator when it is available. This is where all the magic happens. The remote
// player is usually a few frames behind the local emulator state. The emulator
// is reset to the last checkpoint and then both local and remote inputs are
// replayed until they catch up to the current frame.
func (g *Game) applyRemoteInput() {
	if g.remoteInput.Len() == 0 {
		return
	}

	inputSize := min(g.localInput.Len(), g.remoteInput.Len())
	start := time.Now()
	endFrame := g.frame
	g.restoreCheckpoint()

	// Enable PPU fast-forwarding to speed up the replay, since we don't need to
	// render the intermediate frames.
	g.bus.PPU.FastForward = true

	// Enable CPU disassembly if requested. We do it only for frames where we have
	// both local and remote inputs, so that we can compare.
	if g.DisasmEnabled {
		g.bus.DisasmEnabled = true
	}

	// Replay the inputs until the local and remote emulators are in sync.
	for i := 0; i < inputSize; i++ {
		g.LocalJoy.SetButtons(g.localInput.At(i))
		g.RemoteJoy.SetButtons(g.remoteInput.At(i))
		g.playFrame()
	}

	// Disable CPU disassembly, since from now on we have only the predicted input
	// from the remote player, so this part will be rolling back eventually.
	if g.DisasmEnabled {
		g.bus.DisasmEnabled = false
	}

	// This is the last state where both emulators are in sync.
	// Create a new checkpoint, so we can rewind to this state later.
	g.createCheckpoint()

	// In case the local state is ahead (which is almost always the case), we
	// need to replay the local inputs and simulate the remote inputs.
	for i := inputSize; i < g.localInput.Len(); i++ {
		g.RemoteJoy.SetButtons(g.speculatedInput.At(i))
		g.LocalJoy.SetButtons(g.localInput.At(i))

		if g.frame < endFrame {
			g.playFrame()
		}
	}

	if g.frame != endFrame {
		panic(fmt.Errorf("frame advanced from %d to %d", endFrame, g.frame))
	}

	// Replaying a large number of frames will inevitably create some lag
	// for the local player. There is not much we can do about it.
	if delta := time.Since(start); delta > frameDuration {
		log.Printf("[DEBUG] replay lag: %s (%d frames)", delta, delta/frameDuration)
	}

	// There might still be some local inputs left, so we need to keep them.
	g.localInput.TruncFront(inputSize)
	g.remoteInput.TruncFront(inputSize)
	g.speculatedInput.TruncFront(inputSize)

	// Disable fast-forwarding, since we are back to real-time.
	g.bus.PPU.FastForward = false
}
