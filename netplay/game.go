package netplay

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/generic"
	"github.com/maxpoletaev/dendy/nes"
)

const (
	frameDuration = time.Second / 60
)

type inputBatch struct {
	startFrame uint32
	inputs     []uint8
}

type Checkpoint struct {
	Frame uint32
	State []byte
	Crc32 uint32
}

type PlayerInput struct {
	Frame   uint32
	Buttons uint8
}

// Game is a network play state manager. It keeps track of the inputs from both
// players and makes sure their state is synchronized.
type Game struct {
	frame     uint32
	prevFrame uint32

	bus        *nes.Bus
	checkpoint *Checkpoint
	generation uint32

	localInput     []uint8
	remoteInput    *generic.Queue[PlayerInput]
	predictedInput uint8

	LocalJoy      *input.Joystick
	RemoteJoy     *input.Joystick
	DisasmEnabled bool
}

func NewGame(bus *nes.Bus) *Game {
	return &Game{
		bus:        bus,
		checkpoint: &Checkpoint{},
	}
}

// Reset resets the emulator state to the given checkpoint. If cp is nil, the
// emulator is reset to the initial state.
func (g *Game) Reset(cp *Checkpoint) {
	g.frame = 0
	g.prevFrame = 0
	g.generation++

	g.predictedInput = 0
	g.localInput = make([]uint8, 0, cap(g.localInput))
	g.remoteInput = generic.NewQueue[PlayerInput](300)

	if cp != nil {
		g.checkpoint = cp
		g.restoreCheckpoint()
		return
	}

	g.bus.Reset()
	g.createCheckpoint()
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

// Generation returns the current generation number. It is incremented every
// time the game is reset.
func (g *Game) Generation() uint32 {
	return g.generation
}

func (g *Game) playFrame() {
	for {
		tick := g.bus.Tick()

		if tick.FrameComplete {
			g.frame++
			break
		}
	}

	// Overflow will happen after ~2 years of continuous play at 60 FPS :)
	// Don't think it's a problem though.
	if g.frame == 0 {
		panic("frame counter overflow")
	}

	g.prevFrame = g.frame
}

func (g *Game) checkRemoteInput() {
	if !g.remoteInput.Empty() {
		first := g.remoteInput.Front()

		if first.Frame < g.frame {
			inputs := make([]byte, 0, g.remoteInput.Len())

			for !g.remoteInput.Empty() {
				in := g.remoteInput.Front()
				if in.Frame >= g.frame {
					break
				}

				inputs = append(inputs, in.Buttons)
				g.remoteInput.Dequeue()
			}

			g.applyRemoteInput(inputBatch{
				startFrame: first.Frame,
				inputs:     inputs,
			})
		}
	}
}

// RunFrame runs a single frame of the game.
func (g *Game) RunFrame() {
	g.checkRemoteInput()
	g.playFrame()
}

// Sleep pauses the execution for the given number of frames.
func (g *Game) Sleep(d int) {
	time.Sleep(time.Duration(d) * frameDuration)
}

func (g *Game) createCheckpoint() {
	buf := bytes.NewBuffer(g.checkpoint.State[:0])
	encoder := gob.NewEncoder(buf)

	if err := g.bus.Save(encoder); err != nil {
		panic(fmt.Errorf("failed create checkpoint: %w", err))
	}

	g.checkpoint.Frame = g.frame
	g.checkpoint.State = buf.Bytes()
}

func (g *Game) restoreCheckpoint() {
	buf := bytes.NewBuffer(g.checkpoint.State)
	decoder := gob.NewDecoder(buf)

	if err := g.bus.Load(decoder); err != nil {
		panic(fmt.Errorf("failed to restore checkpoint: %w", err))
	}

	g.frame = g.checkpoint.Frame
}

// HandleLocalInput adds records and applies the input from the local player.
// Since the remote player is behind, it assumes that it just keeps pressing
// the same buttons until it catches up. This is not always true, but it's
// good approximation for most games.
func (g *Game) HandleLocalInput(buttons uint8) {
	g.localInput = append(g.localInput, buttons)
	g.RemoteJoy.SetButtons(g.predictedInput)
	g.LocalJoy.SetButtons(buttons)
}

// HandleRemoteInput adds the input from the remote player.
func (g *Game) HandleRemoteInput(input PlayerInput) {
	g.remoteInput.Enqueue(input)

}

// applyRemoteInput applies the input from the remote player to the local
// emulator when it is available. This is where all the magic happens. The remote
// player is usually a few frames behind the local emulator state. The emulator
// is reset to the last checkpoint and then both local and remote inputs are
// replayed until they catch up to the current frame.
func (g *Game) applyRemoteInput(batch inputBatch) {
	g.predictedInput = 0
	if len(batch.inputs) > 0 {
		g.predictedInput = batch.inputs[len(batch.inputs)-1]
	}

	// Need to ensure that the input is not behind the checkpoint, otherwise the
	// states will be out of sync. This should never happen, but in case it fires,
	// something is very broken.
	if batch.startFrame != g.checkpoint.Frame+1 {
		panic(fmt.Sprintf("input is not aligned with the checkpoint: %d != %d", batch.startFrame, g.checkpoint.Frame+1))
	}

	start := time.Now()
	endFrame := g.frame
	g.restoreCheckpoint()
	startFrame := g.frame

	minLen := len(g.localInput)
	if len(batch.inputs) < minLen {
		minLen = len(batch.inputs)
	}

	// Enable PPU fast-forwarding to speed up the replay, since we don't need to
	// render the intermediate frames.
	g.bus.PPU.FastForward = true

	// Enable CPU disassembly if requested. We do it only for frames where we have
	// both local and remote inputs, so that we can compare.
	if g.DisasmEnabled {
		g.bus.DisasmEnabled = true
	}

	// Replay the inputs until the local and remote emulators are in sync.
	for i := 0; i < minLen; i++ {
		g.LocalJoy.SetButtons(g.localInput[i])
		g.RemoteJoy.SetButtons(batch.inputs[i])
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
	for i := minLen; i < len(g.localInput); i++ {
		g.LocalJoy.SetButtons(g.localInput[i])
		g.RemoteJoy.SetButtons(g.predictedInput)

		if g.frame < endFrame {
			g.playFrame()
		}
	}

	if g.frame != endFrame {
		panic(fmt.Errorf("frame advanced from %d to %d", endFrame, g.frame))
	}

	// Replaying a large number of frames will inevitably create some lag
	// for the local player. There is not much we can do about it.
	if time.Since(start) > frameDuration {
		log.Printf("[DEBUG] replay lag: %s (%d frames)", time.Since(start), endFrame-startFrame)
	}

	// There might still be some local inputs left, so we need to keep them.
	newInput := make([]uint8, 0, cap(g.localInput))
	g.localInput = append(newInput, g.localInput[minLen:]...)

	// Disable fast-forwarding, since we are back to real-time.
	g.bus.PPU.FastForward = false
}
