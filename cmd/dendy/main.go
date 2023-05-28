package main

import (
	"flag"
	"fmt"
	"os"

	cpupkg "github.com/maxpoletaev/dendy/cpu"
	"github.com/maxpoletaev/dendy/display"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	ppupkg "github.com/maxpoletaev/dendy/ppu"
)

type opts struct {
	disasm   bool
	showFPS  bool
	stepMode bool
	scale    int
}

func (o *opts) parse() *opts {
	flag.BoolVar(&o.stepMode, "step", false, "enable step mode (press space to step cpu)")
	flag.BoolVar(&o.disasm, "disasm", false, "enable cpu disassembler")
	flag.BoolVar(&o.showFPS, "showfps", false, "show fps counter")
	flag.IntVar(&o.scale, "scale", 2, "scale factor (default: 2)")
	flag.Parse()
	return o
}

func (o *opts) sanitize() {
	if o.scale < 1 {
		o.scale = 1
	}
}

func main() {
	o := new(opts).parse()
	o.sanitize()

	if flag.NArg() != 1 {
		fmt.Println("usage: dendy [-scale=2] [-showfps] [-disasm] <rom_file.nes>")
		os.Exit(1)
	}

	cart, err := ines.Load(flag.Arg(0))
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to open rom file: %s", err))
		os.Exit(1)
	}

	var (
		cpu = cpupkg.New()
		ppu = ppupkg.New(cart)
		joy = input.NewJoystick()
		zap = input.NewZapper()
	)

	var (
		window = display.Show(&ppu.Frame, o.scale)
	)

	cpu.EnableDisasm = o.disasm || o.stepMode
	window.ShowFPS = o.showFPS
	cpu.AllowIllegal = true

	bus := &Bus{
		screen: window,
		cart:   cart,
		cpu:    cpu,
		ppu:    ppu,
		joy1:   joy,
		zap:    zap,
	}

	bus.Reset()

	for !window.ShouldClose() {
		if o.stepMode {
			// Each space key press will execute one cpu instruction.
			if window.KeyPressed(display.KeySpace) {
				for {
					tick := bus.Tick()
					if tick.InstrComplete {
						break
					}
				}
			}

			// Each Enter key press will execute one frame (~30k instructions).
			if window.KeyPressed(display.KeyEnter) {
				for {
					tick := bus.Tick()
					if tick.FrameComplete {
						fmt.Println("") // Separate frames with a newline in the log.
						break
					}
				}
			}

			window.Refresh()

			continue
		}

		tick := bus.Tick()

		// Whether zapper matches the Y coordinate of the target is determined by the
		// current scanline. So we need to handle zapper input after each scanline.
		if tick.ScanlineComplete {
			window.UpdateZapper(zap)
		}

		if tick.FrameComplete {
			window.UpdateJoystick(joy)
			window.HandleHotKeys()
			window.Refresh()

			// There is no light during vblank (the electron beam is off), but the emulated
			// screen is not black, so we need to reset the sensor value here explicitly.
			zap.SetBrightness(0)

			if window.IsResetPressed() {
				bus.Reset()
			}
		}
	}
}
