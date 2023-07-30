package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	cpupkg "github.com/maxpoletaev/dendy/cpu"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/loglevel"
	"github.com/maxpoletaev/dendy/nes"
	"github.com/maxpoletaev/dendy/netplay"
	ppupkg "github.com/maxpoletaev/dendy/ppu"
	"github.com/maxpoletaev/dendy/screen"
)

type opts struct {
	scale         int
	noSpriteLimit bool
	connectAddr   string
	listenAddr    string
	bufsize       int
	noSave        bool
	showFPS       bool
	verbose       bool
	disasm        string
	cpuprof       string
	fps           int
}

func (o *opts) parse() *opts {
	flag.IntVar(&o.scale, "scale", 2, "scale factor (default: 2)")
	flag.BoolVar(&o.noSpriteLimit, "nospritelimit", false, "disable sprite limit")
	flag.StringVar(&o.connectAddr, "connect", "", "netplay connect address (default: none)")
	flag.StringVar(&o.listenAddr, "listen", "", "netplay listen address (default: none)")
	flag.IntVar(&o.bufsize, "bufsize", 0, "netplay input buffer size (default: 0)")
	flag.BoolVar(&o.noSave, "nosave", false, "disable save states")
	flag.BoolVar(&o.showFPS, "showfps", false, "show fps counter")

	// Debugging flags.
	flag.StringVar(&o.cpuprof, "cpuprof", "", "write cpu profile to file")
	flag.StringVar(&o.disasm, "disasm", "", "write cpu disassembly to file")
	flag.BoolVar(&o.verbose, "verbose", false, "enable verbose logging")
	flag.IntVar(&o.fps, "fps", 0, "set emulator speed (default: 60)")

	flag.Parse()
	return o
}

func (o *opts) sanitize() {
	if o.scale < 1 {
		o.scale = 1
	}
}

func (o *opts) logLevel() loglevel.Level {
	if o.verbose {
		return loglevel.LevelDebug
	}

	return loglevel.LevelInfo
}

func loadState(bus *nes.Bus, saveFile string) (bool, error) {
	f, err := os.OpenFile(saveFile, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("[ERROR] failed to close save file: %s", err)
		}
	}()

	decoder := gob.NewDecoder(f)
	if err := bus.Load(decoder); err != nil {
		return false, err
	}

	return true, nil
}

func saveState(bus *nes.Bus, saveFile string) error {
	tmpFile := saveFile + ".tmp"

	f, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("[ERROR] failed to close save file: %s", err)
		}
	}()

	encoder := gob.NewEncoder(f)
	if err := bus.Save(encoder); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, saveFile); err != nil {
		return err
	}

	return nil
}

func runOffline(bus *nes.Bus, o *opts, saveFile string) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()
	bus.Zap = input.NewZapper()
	bus.Reset()

	if o.disasm != "" {
		file, err := os.Create(o.disasm)
		if err != nil {
			log.Printf("[ERROR] failed to create disassembly file: %s", err)
			os.Exit(1)
		}

		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("[ERROR] failed to close disassembly file: %s", err)
			}
		}()

		bus.DisasmWriter = bufio.NewWriterSize(file, 1024*1024)
		bus.DisasmEnabled = true
	}

	if !o.noSave {
		if ok, err := loadState(bus, saveFile); err != nil {
			log.Printf("[ERROR] failed to load save file: %s", err)
			os.Exit(1)
		} else if ok {
			log.Printf("[INFO] state loaded: %s", saveFile)
		}
	}

	w := screen.Show(&bus.PPU.Frame, o.scale)
	w.InputDelegate = bus.Joy1.SetButtons
	w.ZapperDelegate = bus.Zap.Update
	w.ResetDelegate = bus.Reset
	w.ShowFPS = o.showFPS

	if o.fps > 0 {
		w.SetFrameRate(o.fps)
	}

	for {
		tick := bus.Tick()

		if tick.ScanlineComplete {
			w.UpdateZapper()
		}

		if tick.FrameComplete {
			if w.ShouldClose() {
				break
			}

			bus.Zap.VBlank()
			w.UpdateJoystick()
			w.HandleHotKeys()
			w.Refresh()

			for !w.InFocus() {
				if w.ShouldClose() {
					break
				}

				w.Refresh()
			}
		}
	}

	if !o.noSave {
		if err := saveState(bus, saveFile); err != nil {
			log.Printf("[ERROR] failed to save state: %s", err)
			os.Exit(1)
		}

		log.Printf("[INFO] state saved: %s", saveFile)
	}
}

func runAsServer(bus *nes.Bus, o *opts) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()

	game := netplay.NewGame(bus)
	game.BufferSize = o.bufsize
	game.RemoteJoy = bus.Joy2
	game.LocalJoy = bus.Joy1
	game.Reset(nil)

	if o.disasm != "" {
		file, err := os.Create(o.disasm)
		if err != nil {
			log.Printf("[ERROR] failed to create disassembly file: %s", err)
			os.Exit(1)
		}

		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("[ERROR] failed to close disassembly file: %s", err)
			}
		}()

		bus.DisasmWriter = bufio.NewWriterSize(file, 1024*1024)
		bus.DisasmEnabled = false // will be controlled by the game
		game.DisasmEnabled = true
	}

	log.Printf("[INFO] waiting for client...")
	server, addr, err := netplay.Listen(game, o.listenAddr)

	if err != nil {
		log.Printf("[ERROR] failed to listen: %v", err)
		os.Exit(1)
	}

	log.Printf("[INFO] client connected: %s", addr)
	log.Printf("[INFO] starting game...")

	w := screen.Show(&bus.PPU.Frame, o.scale)
	w.SetTitle(fmt.Sprintf("%s (P1)", screen.Title))
	w.InputDelegate = server.SendButtons
	w.ResetDelegate = server.SendReset
	w.ShowFPS = o.showFPS
	w.ShowPing = true

	if o.fps > 0 {
		w.SetFrameRate(o.fps)
	}

	server.SendReset()
	server.Start()

	for {
		if w.ShouldClose() {
			break
		}

		w.SetLatency(server.Latency())
		w.HandleHotKeys()
		w.UpdateJoystick()
		server.RunFrame()
		w.Refresh()
	}
}

func runAsClient(bus *nes.Bus, o *opts) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()

	game := netplay.NewGame(bus)
	game.BufferSize = o.bufsize
	game.RemoteJoy = bus.Joy1
	game.LocalJoy = bus.Joy2
	game.Reset(nil)

	if o.disasm != "" {
		file, err := os.Create(o.disasm)
		if err != nil {
			log.Printf("[ERROR] failed to create disassembly file: %s", err)
			os.Exit(1)
		}

		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("[ERROR] failed to close disassembly file: %s", err)
			}
		}()

		bus.DisasmWriter = bufio.NewWriterSize(file, 1024*1024)
		bus.DisasmEnabled = false // will be controlled by the game
		game.DisasmEnabled = true
	}

	log.Printf("[INFO] connecting to server...")
	client, addr, err := netplay.Connect(game, o.connectAddr)

	if err != nil {
		log.Printf("[ERROR] failed to connect: %v", err)
		os.Exit(1)
	}

	log.Printf("[INFO] connected to server: %s", addr)
	log.Printf("[INFO] starting game...")

	w := screen.Show(&bus.PPU.Frame, o.scale)
	w.SetTitle(fmt.Sprintf("%s (P2)", screen.Title))
	w.InputDelegate = client.SendButtons
	w.ShowFPS = o.showFPS
	w.ShowPing = true

	if o.fps > 0 {
		w.SetFrameRate(o.fps)
	}

	client.Start()

	for {
		if w.ShouldClose() {
			break
		}

		w.SetLatency(client.Latency())
		w.HandleHotKeys()
		w.UpdateJoystick()
		client.RunFrame()
		w.Refresh()
	}
}

func main() {
	o := new(opts).parse()
	o.sanitize()

	log.Default().SetFlags(0)
	log.Default().SetOutput(&loglevel.LevelFilter{
		Level:  o.logLevel(),
		Output: os.Stderr,
	})

	if flag.NArg() != 1 {
		fmt.Println("usage: dendy [-scale=2] [-nosave] [-nospritelimit] [-listen=<addr>:<port>] [-connect=<addr>:<port>] romfile")
		os.Exit(1)
	}

	if o.cpuprof != "" {
		log.Printf("[INFO] writing cpu profile to %s", o.cpuprof)

		f, err := os.Create(o.cpuprof)
		if err != nil {
			log.Printf("[ERROR] failed to create cpu profile: %v", err)
			os.Exit(1)
		}

		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Printf("[ERROR] failed to start cpu profile: %v", err)
			os.Exit(1)
		}

		defer pprof.StopCPUProfile()
	}

	romFile := flag.Arg(0)
	log.Printf("[INFO] loading rom file: %s", romFile)

	cart, err := ines.Load(romFile)
	if err != nil {
		log.Printf("[ERROR] failed to open rom file: %s", err)
		os.Exit(1)
	}

	cpu := cpupkg.New()
	cpu.AllowIllegal = true

	ppu := ppupkg.New(cart)
	ppu.NoSpriteLimit = o.noSpriteLimit

	bus := &nes.Bus{
		Cart: cart,
		CPU:  cpu,
		PPU:  ppu,
	}

	switch {
	case o.listenAddr != "":
		log.Printf("[INFO] starting server mode")
		runAsServer(bus, o)
	case o.connectAddr != "":
		log.Printf("[INFO] starting client mode")
		runAsClient(bus, o)
	default:
		saveFile := strings.TrimSuffix(romFile, filepath.Ext(romFile)) + ".save"
		log.Printf("[INFO] starting offline mode")
		runOffline(bus, o, saveFile)
	}
}
