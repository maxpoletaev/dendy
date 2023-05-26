package main

import (
	"flag"
	"fmt"
	"os"
	"time"

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
}

func (o *opts) parse() *opts {
	flag.BoolVar(&o.stepMode, "step", false, "enable step mode (press space to step cpu)")
	flag.BoolVar(&o.disasm, "disasm", false, "enable cpu disassembler")
	flag.BoolVar(&o.showFPS, "fps", false, "show fps")
	flag.Parse()
	return o
}

func main() {
	o := new(opts).parse()

	if flag.NArg() != 1 {
		fmt.Println("usage: dendy [-fps] [-disasm] <rom_file.nes>")
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
		window = display.Show(&ppu.Frame, joy, zap, 2)
	)

	cpu.EnableDisasm = o.disasm || o.stepMode
	window.ShowFPS = o.showFPS
	cpu.AllowIllegal = true

	bus := &Bus{
		cart:   cart,
		screen: window,
		cpu:    cpu,
		ppu:    ppu,
		joy1:   joy,
		zap:    zap,
	}

	bus.Reset()
	window.NoSignal()

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

			// Each F key press will execute one frame.
			if window.KeyPressed(display.KeyF) {
				for {
					_, frameComplete := bus.Tick()
					if frameComplete {
						fmt.Println("") // Separate frames with a newline in the log.
						break
					}
				}
			}

			time.Sleep(100 * time.Millisecond)
			window.Noop()
			continue
		}

		bus.Tick()
	}
}
