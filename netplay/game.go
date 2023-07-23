package netplay

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/generic"
	"github.com/maxpoletaev/dendy/internal/rolling"
	"github.com/maxpoletaev/dendy/nes"
)

const (
	frameDuration     = time.Second / 60
	defaultInputBatch = 5
)

type Checkpoint struct {
	Frame int32
	State []byte
}

type InputBatch struct {
	StartFrame int32
	Input      []byte
}

func (b *InputBatch) Add(input uint8) {
	b.Input = append(b.Input, input)
}

// Game is a network play state manager. It keeps track of the inputs from both
// players and makes sure their state is synchronized.
type Game struct {
	frame          int32
	bus            *nes.Bus
	localInput     []uint8
	checkpoint     *Checkpoint
	remoteInput    *generic.Queue[InputBatch]
	simulatedInput uint8
	incarnation    uint32

	LocalJoy  *input.Joystick
	RemoteJoy *input.Joystick
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
	g.incarnation++

	g.localInput = nil
	g.simulatedInput = 0
	g.remoteInput = generic.NewQueue[InputBatch]()

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
func (g *Game) Frame() int32 {
	return g.frame
}

// Incarnation returns the current incarnation number. It is incremented every
// time the game is reset.
func (g *Game) Incarnation() uint32 {
	return g.incarnation
}

func (g *Game) playFrame() {
	for {
		tick := g.bus.Tick()

		if tick.FrameComplete {
			g.frame++
			break
		}
	}
}

// RunFrame runs a single frame of the game.
func (g *Game) RunFrame() {
	if in, ok := g.remoteInput.Peek(); ok {
		endFrame := in.StartFrame + int32(len(in.Input))

		switch rolling.Compare(g.frame, endFrame) {
		case rolling.Greater, rolling.Equal:
			g.applyRemoteInput(in)
			g.remoteInput.Pop()
		}
	}

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

// AddLocalInput adds records and applies the input from the local player.
// Since the remote player is behind, it assumes that it just keeps pressing
// the same buttons until it catches up. This is not always true, but it's
// good approximation for most games.
func (g *Game) AddLocalInput(buttons uint8) {
	g.localInput = append(g.localInput, buttons)
	g.RemoteJoy.SetButtons(g.simulatedInput)
	g.LocalJoy.SetButtons(buttons)
}

// AddRemoteInput adds the input from the remote player.
func (g *Game) AddRemoteInput(batch InputBatch) {
	g.remoteInput.Push(batch)
}

// applyRemoteInput applies the input from the remote player to the local
// emulator when it is available. This is where all the magic happens. The remote
// player is usually a few frames behind the local emulator state. The emulator
// is reset to the last checkpoint and then both local and remote inputs are
// replayed until they catch up to the current frame.
func (g *Game) applyRemoteInput(batch InputBatch) {
	g.simulatedInput = 0
	if len(batch.Input) > 0 {
		g.simulatedInput = batch.Input[len(batch.Input)-1]
	}

	// Need to ensure that the input is not behind the checkpoint, otherwise the
	// states will be out of sync. This should never happen, but in case it fires,
	// something is very broken.
	if rolling.LessThan(batch.StartFrame, g.checkpoint.Frame) {
		panic(fmt.Errorf("input is behind the checkpoint: %d < %d", batch.StartFrame, g.checkpoint.Frame))
	}

	start := time.Now()
	endFrame := g.frame
	g.restoreCheckpoint()
	numFrames := endFrame - g.frame

	minLen := len(g.localInput)
	if len(batch.Input) < minLen {
		minLen = len(batch.Input)
	}

	// Enable PPU fast-forwarding to speed up the replay, since we don't need to
	// render the intermediate frames.
	g.bus.PPU.BeginFastForward()
	defer g.bus.PPU.EndFastForward()

	// Replay the inputs until the local and remote emulators are in sync.
	for i := 0; i < minLen; i++ {
		g.LocalJoy.SetButtons(g.localInput[i])
		g.RemoteJoy.SetButtons(batch.Input[i])
		g.playFrame()
	}

	// This is the last state where both emulators are in sync.
	// Create a new checkpoint, so we can restore to this state later.
	g.createCheckpoint()

	// In case the local state is ahead (which is almost always the case), we
	// need to replay the local inputs and simulate the remote inputs.
	for i := minLen; i < len(g.localInput); i++ {
		g.LocalJoy.SetButtons(g.localInput[i])
		g.RemoteJoy.SetButtons(g.simulatedInput)

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
		log.Printf("[WARN] replay lag: %s (%d frames)", time.Since(start), numFrames)
	}

	// There might still be some local inputs left, so we need to keep them.
	newInput := make([]uint8, 0, len(g.localInput))
	g.localInput = append(newInput, g.localInput[minLen:]...)
}
