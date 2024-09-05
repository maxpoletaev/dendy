package system

import (
	"errors"
	"fmt"
	"image/color"
	"io"

	apupkg "github.com/maxpoletaev/dendy/apu"
	cpupkg "github.com/maxpoletaev/dendy/cpu"
	"github.com/maxpoletaev/dendy/disasm"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/binario"
	ppupkg "github.com/maxpoletaev/dendy/ppu"
)

// System the emulated system. It owns all the components and is responsible for
// coordinating their interactions. It also provides the main interface for
// running the emulation.
type System struct {
	bus   *Bus
	ram   []byte
	cpu   *cpupkg.CPU
	ppu   *ppupkg.PPU
	apu   *apupkg.APU
	cart  ines.Cartridge
	port1 input.Device
	port2 input.Device

	scanlineReady bool
	frameReady    bool
	cycles        uint64
	debugWriter   io.StringWriter
}

// New creates a new System instance with the given Cartridge and input devices.
func New(cart ines.Cartridge, port1, port2 input.Device) *System {
	ram := make([]byte, 2048)
	ppu := ppupkg.New(cart)
	cpu := cpupkg.New()
	apu := apupkg.New()

	s := &System{
		ram:   ram,
		cpu:   cpu,
		ppu:   ppu,
		apu:   apu,
		cart:  cart,
		port1: port1,
		port2: port2,
		bus:   newBus(ram, ppu, apu, cart, port1, port2),
	}

	s.initDMACallbacks()

	s.Reset()

	return s
}

func (s *System) initDMACallbacks() {
	// PPU DMA transfers 256 bytes of data from CPU memory to PPU OAM memory.
	// It is triggered by writing to $4014 and takes 513 CPU cycles to complete.
	s.ppu.SetDMACallback(func(addr uint16, data []byte) {
		for i := uint16(0); i < uint16(len(data)); i++ {
			data[i] = s.bus.Read(addr + i)
		}

		s.cpu.Halt += 513
		if s.cpu.Halt%2 == 1 {
			s.cpu.Halt++
		}
	})

	// APU DMA transfers an audio sample (1 byte) from CPU memory to APU memory.
	// It happens automatically when DMC requests a sample and takes 4 CPU cycles.
	s.apu.SetDMACallback(func(addr uint16) byte {
		data := s.bus.Read(addr)
		s.cpu.Halt += 4
		return data
	})
}

func (s *System) Reset() {
	// NOTE: Order matters
	s.cart.Reset()
	s.ppu.Reset()
	s.apu.Reset()
	s.port1.Reset()
	s.port2.Reset()
	s.cpu.Reset(s.bus)

	s.cycles = 0
	s.frameReady = false
	s.scanlineReady = false
}

func (s *System) disassemble() {
	_, err1 := s.debugWriter.WriteString(disasm.DebugStep(s.bus, s.cpu))
	_, err2 := s.debugWriter.WriteString("\n")

	if err := errors.Join(err1, err2); err != nil {
		panic(fmt.Sprintf("error writing disassembly: %v", err))
	}
}

// Tick advances the emulation by one internal clock cycle (PPU cycle).
func (s *System) Tick() {
	s.cycles++

	if s.cycles%3 == 0 {
		instructionComplete := s.cpu.Tick(s.bus)

		if instructionComplete && s.debugWriter != nil {
			s.disassemble()
		}

		s.apu.Tick()

		if s.apu.PendingIRQ {
			s.apu.PendingIRQ = false
			s.cpu.TriggerIRQ()
		}
	}

	s.ppu.Tick()

	if s.ppu.PendingNMI {
		s.ppu.PendingNMI = false
		s.cpu.TriggerNMI()
	}

	if s.ppu.ScanlineComplete {
		s.ppu.ScanlineComplete = false
		s.scanlineReady = true

		s.cart.ScanlineTick()

		if s.cart.PendingIRQ() {
			s.cpu.TriggerIRQ()
		}
	}

	if s.ppu.FrameComplete {
		s.ppu.FrameComplete = false
		s.frameReady = true
	}
}

// SetFastForward sets the fast-forward mode. In this mode, the emulator will
// skip rendering frames and audio samples, and will only run the CPU and PPU.
func (s *System) SetFastForward(v bool) {
	s.ppu.FastForward = v
}

// ScanlineReady returns true if a scanline has just completed.
func (s *System) ScanlineReady() (v bool) {
	if s.scanlineReady {
		s.scanlineReady = false
		return true
	}
	return false
}

// FrameReady returns true if a frame has just completed.
func (s *System) FrameReady() (v bool) {
	if s.frameReady {
		s.frameReady = false
		return true
	}
	return false
}

// Frame returns a pointer to the current frame buffer.
// The returned buffer is only valid until the next call to Tick.
func (s *System) Frame() []color.RGBA {
	return s.ppu.Frame
}

// AudioSample returns the next audio sample from the APU.
func (s *System) AudioSample() float32 {
	return s.apu.Output()
}

// SetDebugOutput sets the writer for debug (disassembly) output.
func (s *System) SetDebugOutput(w io.StringWriter) {
	s.debugWriter = w
}

// SetNoSpriteLimit enables or disables scanline sprite limit on the PPU.
func (s *System) SetNoSpriteLimit(v bool) {
	s.ppu.NoSpriteLimit = v
}

// SaveState saves the current state of the system to the given writer.
func (s *System) SaveState(w *binario.Writer) error {
	err := errors.Join(
		w.WriteByteSlice(s.ram[:]),
		w.WriteUint64(s.cycles),
		s.cpu.SaveState(w),
		s.ppu.SaveState(w),
		s.apu.SaveState(w),
		s.cart.SaveState(w),
		s.port1.SaveState(w),
		s.port2.SaveState(w),
	)

	return err
}

// LoadState loads the state of the system from the given reader.
func (s *System) LoadState(r *binario.Reader) error {
	err := errors.Join(
		r.ReadByteSliceTo(s.ram[:]),
		r.ReadUint64To(&s.cycles),
		s.cpu.LoadState(r),
		s.ppu.LoadState(r),
		s.apu.LoadState(r),
		s.cart.LoadState(r),
		s.port1.LoadState(r),
		s.port2.LoadState(r),
	)

	return err
}
