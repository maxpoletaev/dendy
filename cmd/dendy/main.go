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
		window = display.Show(&ppu.Frame, joy, zap, o.scale)
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
					instrComplete, _ := bus.Tick()
					if instrComplete {
						break
					}
				}
			}

			// Each Enter key press will execute one frame (~30k instructions).
			if window.KeyPressed(display.KeyEnter) {
				for {
					_, frameComplete := bus.Tick()
					if frameComplete {
						fmt.Println("") // Separate frames with a newline in the log.
						break
					}
				}
			}

			window.Refresh()

			continue
		}

		_, frameComplete := bus.Tick()
		if frameComplete {
			window.Refresh()
			window.HandleInput()

			if window.IsResetPressed() {
				bus.Reset()
			}
		}
	}
}
