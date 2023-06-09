package netplay

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/generic"
	"github.com/maxpoletaev/dendy/nes"
)

type Checkpoint struct {
	Frame uint64
	State []byte
}

type InputBatch struct {
	StartFrame uint64
	Input      []byte
}

func (b *InputBatch) Add(input uint8) {
	b.Input = append(b.Input, input)
}

// Game is a network play state manager. It keeps track of the inputs from both
// players and makes sure their state is synchronized.
type Game struct {
	bus            *nes.Bus
	frame          uint64
	localInput     []uint8
	checkpoint     *Checkpoint
	remoteInput    *generic.Queue[InputBatch]
	simulatedInput uint8

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
func (g *Game) Frame() uint64 {
	return g.frame
}

func (g *Game) playFrame() {
	for {
		tr := g.bus.Tick()

		if tr.FrameComplete {
			g.frame++
			break
		}
	}
}

// RunFrame runs a single frame of the game.
func (g *Game) RunFrame() {
	if in, ok := g.remoteInput.Peek(); ok {
		endFrame := in.StartFrame + uint64(len(in.Input))

		if g.frame >= endFrame {
			g.applyRemoteInput(in)
			g.remoteInput.Pop()
		}
	}

	g.playFrame()
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

// AddRemoteInput adds the input from the remote player. This is where all the
// magic happens. The remote input is usually a few frames behind the local
// emulator state. The emulator is reset to the last checkpoint and then both
// local and remote inputs are replayed until it catches up to the current frame.
func (g *Game) AddRemoteInput(batch InputBatch) {
	g.remoteInput.Push(batch)
}

func (g *Game) applyRemoteInput(batch InputBatch) {
	if len(batch.Input) > 0 {
		g.simulatedInput = batch.Input[len(batch.Input)-1]
	} else {
		g.simulatedInput = 0
	}

	endFrame := g.frame
	g.restoreCheckpoint()

	minLen := len(g.localInput)
	if len(batch.Input) < minLen {
		minLen = len(batch.Input)
	}

	// Replay the inputs until the local and remote emulators are in sync.
	for i := 0; i < minLen; i++ {
		g.LocalJoy.SetButtons(g.localInput[i])
		g.RemoteJoy.SetButtons(batch.Input[i])
		g.playFrame()
	}

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

	newInput := make([]uint8, 0, len(g.localInput))
	g.localInput = append(newInput, g.localInput[minLen:]...)
}
