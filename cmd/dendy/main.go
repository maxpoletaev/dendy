package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	cpupkg "github.com/maxpoletaev/dendy/cpu"
	"github.com/maxpoletaev/dendy/display"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/nes"
	"github.com/maxpoletaev/dendy/netplay"
	ppupkg "github.com/maxpoletaev/dendy/ppu"
)

type opts struct {
	disasm   bool
	showFPS  bool
	slowMode bool
	scale    int

	cpuprof     string
	listenAddr  string
	connectAddr string
	batchSize   int
}

func (o *opts) parse() *opts {
	flag.BoolVar(&o.slowMode, "slow", false, "enable slow mode")
	flag.BoolVar(&o.disasm, "disasm", false, "enable cpu disassembler")
	flag.BoolVar(&o.showFPS, "showfps", false, "show fps counter")
	flag.IntVar(&o.scale, "scale", 2, "scale factor (default: 2)")
	flag.StringVar(&o.cpuprof, "cpuprof", "", "write cpu profile to file")

	flag.IntVar(&o.batchSize, "batchsize", 10, "input batch size for netplay (default: 10)")
	flag.StringVar(&o.connectAddr, "connect", "", "netplay connect address (default: none)")
	flag.StringVar(&o.listenAddr, "listen", "", "netplay listen address (default: none)")

	flag.Parse()
	return o
}

func (o *opts) sanitize() {
	if o.scale < 1 {
		o.scale = 1
	}
}

func runOffline(bus *nes.Bus, o *opts) {
	bus.Joy1 = input.NewJoystick()
	bus.Zap = input.NewZapper()
	bus.Reset()

	w := display.Show(&bus.PPU.Frame, o.scale)
	w.InputDelegate = bus.Joy1.SetButtons
	w.ZapperDelegate = bus.Zap.Update
	w.ShowFPS = o.showFPS

	if o.slowMode {
		w.ToggleSlowMode()
	}

	for {
		tick := bus.Tick()

		if tick.ScanlineComplete {
			w.UpdateZapper()
		}

		if tick.FrameComplete {
			if w.ShouldClose() {
				return
			}

			if w.IsResetPressed() {
				bus.Reset()
			}

			w.HandleHotKeys()
			w.UpdateJoystick()
			w.Refresh()
		}
	}
}

func runServer(bus *nes.Bus, o *opts) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()

	game := netplay.NewGame(bus)
	game.RemoteJoy = bus.Joy2
	game.LocalJoy = bus.Joy1
	game.Reset(nil)

	fmt.Printf("waiting for client...\n")

	server, err := netplay.Listen(game, o.listenAddr, netplay.Options{BatchSize: o.batchSize})
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		os.Exit(1)
	}

	w := display.Show(&bus.PPU.Frame, o.scale)
	w.SetTitle(fmt.Sprintf("%s (P1)", display.WindowTitle))
	w.InputDelegate = server.SendInput
	w.ShowFPS = o.showFPS

	if o.slowMode {
		w.ToggleSlowMode()
	}

	server.SendReset()
	server.Start()

	for {
		if w.ShouldClose() {
			return
		}

		if w.IsResetPressed() {
			server.SendReset()
		}

		w.HandleHotKeys()
		w.UpdateJoystick()
		server.RunFrame()
		w.Refresh()
	}
}

func runClient(bus *nes.Bus, o *opts) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()

	game := netplay.NewGame(bus)
	game.RemoteJoy = bus.Joy1
	game.LocalJoy = bus.Joy2
	game.Reset(nil)

	fmt.Println("connecting to server...")

	client, err := netplay.Connect(game, o.connectAddr, netplay.Options{BatchSize: o.batchSize})
	if err != nil {
		fmt.Printf("failed to connect: %v\n", err)
		os.Exit(1)
	}

	w := display.Show(&bus.PPU.Frame, o.scale)
	w.SetTitle(fmt.Sprintf("%s (P2)", display.WindowTitle))
	w.InputDelegate = client.SendInput
	w.ShowFPS = o.showFPS

	if o.slowMode {
		w.ToggleSlowMode()
	}

	client.Start()

	for {
		if w.ShouldClose() {
			return
		}

		if w.IsResetPressed() {
			fmt.Printf("reset not supported in client mode\n")
		}

		w.HandleHotKeys()
		w.UpdateJoystick()
		client.RunFrame()
		w.Refresh()
	}
}

func main() {
	o := new(opts).parse()
	o.sanitize()

	if flag.NArg() != 1 {
		fmt.Println("usage: dendy [-scale=2] [-showfps] [-disasm] <rom_file.nes>")
		os.Exit(1)
	}

	if o.cpuprof != "" {
		fmt.Printf("writing cpu profile to %s\n", o.cpuprof)

		f, err := os.Create(o.cpuprof)
		if err != nil {
			fmt.Printf("failed to create cpu profile: %v\n", err)
			os.Exit(1)
		}

		err = pprof.StartCPUProfile(f)
		if err != nil {
			fmt.Printf("failed to start cpu profile: %v\n", err)
			os.Exit(1)
		}

		defer pprof.StopCPUProfile()
	}

	cart, err := ines.Load(flag.Arg(0))
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to open rom file: %s", err))
		os.Exit(1)
	}

	var (
		cpu = cpupkg.New()
		ppu = ppupkg.New(cart)
	)

	cpu.EnableDisasm = o.disasm
	cpu.AllowIllegal = true

	// Initialize a basic console bus. The controllers will
	// depend on the mode, so we'll initialize them later.
	bus := &nes.Bus{
		Cart: cart,
		CPU:  cpu,
		PPU:  ppu,
	}

	switch {
	case o.listenAddr != "":
		runServer(bus, o)
	case o.connectAddr != "":
		runClient(bus, o)
	default:
		runOffline(bus, o)
	}
}
