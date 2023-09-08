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

	apupkg "github.com/maxpoletaev/dendy/apu"
	cpupkg "github.com/maxpoletaev/dendy/cpu"
	ppupkg "github.com/maxpoletaev/dendy/ppu"

	"github.com/maxpoletaev/dendy/console"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/loglevel"
	"github.com/maxpoletaev/dendy/netplay"
	"github.com/maxpoletaev/dendy/ui"
)

const (
	sampleSize       = 32
	framesPerSecond  = 60
	samplesPerSecond = 44100
	ticksPerSecond   = 1789773 * 3
	ticksPerFrame    = ticksPerSecond / framesPerSecond
	samplesPerFrame  = samplesPerSecond / framesPerSecond
	ticksPerSample   = ticksPerSecond / samplesPerSecond
	windowTitle      = "Dendy Emulator"
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
	sound         bool
}

func (o *opts) parse() *opts {
	flag.IntVar(&o.scale, "scale", 2, "scale factor (default: 2)")
	flag.BoolVar(&o.noSpriteLimit, "nospritelimit", false, "disable sprite limit")
	flag.StringVar(&o.connectAddr, "connect", "", "netplay connect address (default: none)")
	flag.StringVar(&o.listenAddr, "listen", "", "netplay listen address (default: none)")
	flag.IntVar(&o.bufsize, "bufsize", 0, "netplay input buffer size (default: 0)")
	flag.BoolVar(&o.noSave, "nosave", false, "disable save states")
	flag.BoolVar(&o.showFPS, "showfps", false, "show fps counter")
	flag.BoolVar(&o.sound, "sound", false, "enable sound emulation")

	// Debugging flags.
	flag.StringVar(&o.cpuprof, "cpuprof", "", "write cpu profile to file")
	flag.StringVar(&o.disasm, "disasm", "", "write cpu disassembly to file")
	flag.BoolVar(&o.verbose, "verbose", false, "enable verbose logging")

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

func loadState(bus *console.Bus, saveFile string) (bool, error) {
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

func saveState(bus *console.Bus, saveFile string) error {
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

func runOffline(bus *console.Bus, o *opts, saveFile string) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()
	bus.Zapper = input.NewZapper()
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

	w := ui.CreateWindow(&bus.PPU.Frame, o.scale, o.verbose)
	w.SetFrameRate(framesPerSecond)
	w.SetTitle(windowTitle)
	defer w.Close()

	w.InputDelegate = bus.Joy1.SetButtons
	w.ZapperDelegate = bus.Zapper.Update
	w.ResetDelegate = bus.Reset
	w.ShowFPS = o.showFPS

	samples := make(chan float32, samplesPerSecond)
	audio := ui.CreateAudio(samplesPerSecond, sampleSize, 1, samplesPerFrame)
	audio.SetChannel(samples)
	defer audio.Close()

	var (
		sampleAcc      = float32(0.0)
		sampleDuration = float32(1.0 / ticksPerSample)
	)

	for {
		tick := bus.Tick()
		sampleAcc += sampleDuration

		if sampleAcc >= 1.0 {
			sample := bus.APU.Output()
			sampleAcc -= 1.0

			select {
			case samples <- sample:
				// noop
			default:
				log.Printf("[WARN] audio buffer overrun")
				samples <- sample
			}
		}

		if tick.ScanlineComplete {
			w.UpdateZapper()
		}

		if tick.FrameComplete {
			if w.ShouldClose() {
				break
			}

			audio.Update()

			bus.Zapper.VBlank()
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

func runAsServer(bus *console.Bus, o *opts) {
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
	sess, addr, err := netplay.Listen(game, o.listenAddr)

	if err != nil {
		log.Printf("[ERROR] failed to listen: %v", err)
		os.Exit(1)
	}

	log.Printf("[INFO] client connected: %s", addr)
	log.Printf("[INFO] starting game...")

	w := ui.CreateWindow(&bus.PPU.Frame, o.scale, o.verbose)
	w.SetTitle(fmt.Sprintf("%s (P1)", windowTitle))
	w.SetFrameRate(framesPerSecond)
	w.InputDelegate = sess.SendButtons
	w.ResetDelegate = sess.SendReset
	w.ShowFPS = o.showFPS
	w.ShowPing = true

	sess.SendReset()
	sess.Start()

	for {
		if w.ShouldClose() {
			break
		}

		w.SetLatencyInfo(sess.Latency())
		w.HandleHotKeys()
		w.UpdateJoystick()
		sess.RunFrame()
		w.Refresh()
	}
}

func runAsClient(bus *console.Bus, o *opts) {
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
	sess, addr, err := netplay.Connect(game, o.connectAddr)

	if err != nil {
		log.Printf("[ERROR] failed to connect: %v", err)
		os.Exit(1)
	}

	log.Printf("[INFO] connected to server: %s", addr)
	log.Printf("[INFO] starting game...")

	w := ui.CreateWindow(&bus.PPU.Frame, o.scale, o.verbose)
	w.SetTitle(fmt.Sprintf("%s (P2)", windowTitle))
	w.SetFrameRate(framesPerSecond)
	w.InputDelegate = sess.SendButtons
	w.ShowFPS = o.showFPS
	w.ShowPing = true

	sess.Start()

	for {
		if w.ShouldClose() {
			break
		}

		w.SetLatencyInfo(sess.Latency())
		w.HandleHotKeys()
		w.UpdateJoystick()
		sess.RunFrame()
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
		fmt.Println("usage: dendy [-scale=2] [-nosave] [-nospritelimit] [-listen=addr:port] [-connect=addr:port] romfile")
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

	apu := apupkg.New()
	apu.Enabled = o.sound

	if o.sound && (o.listenAddr != "" || o.connectAddr != "") {
		log.Printf("[WARN] sound is not supported in netplay mode")
		apu.Enabled = false
	}

	cpu := cpupkg.New()
	cpu.AllowIllegal = true

	ppu := ppupkg.New(cart)
	ppu.NoSpriteLimit = o.noSpriteLimit

	bus := &console.Bus{
		Cart: cart,
		CPU:  cpu,
		PPU:  ppu,
		APU:  apu,
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
