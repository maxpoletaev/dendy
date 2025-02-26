package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/genie"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/internal/loglevel"
)

const (
	windowTitle = "Dendy Emulator"
)

type options struct {
	scale         int
	noSpriteLimit bool
	saveFile      string
	noSave        bool
	showFPS       bool
	verbose       bool
	disasm        string
	memprof       string
	cpuprof       string
	protocol      string
	gg            string
	mute          bool
	noLogo        bool
	noCRT         bool

	connectAddr string
	listenAddr  string
	relayAddr   string
	joinRoom    string
	createRoom  bool
}

func (o *options) parse() *options {
	flag.IntVar(&o.scale, "scale", 2, "scale factor (default: 2)")
	flag.StringVar(&o.saveFile, "savefile", "", "save file (default: romname.save)")
	flag.BoolVar(&o.noSpriteLimit, "nospritelimit", false, "disable sprite limit (eliminates flickering)")
	flag.BoolVar(&o.noSave, "nosave", false, "disable save states")
	flag.BoolVar(&o.showFPS, "showfps", false, "show fps counter")
	flag.BoolVar(&o.mute, "mute", false, "disable apu emulation")
	flag.BoolVar(&o.noLogo, "nologo", false, "do not print logo")
	flag.BoolVar(&o.noCRT, "nocrt", false, "disable CRT effect")
	flag.StringVar(&o.gg, "gg", "", "game genie codes (comma separated)")

	flag.StringVar(&o.protocol, "protocol", "tcp", "netplay protocol (tcp, udp)")
	flag.StringVar(&o.listenAddr, "listen", "", "netplay listen address")
	flag.StringVar(&o.connectAddr, "connect", "", "netplay connect address")
	flag.StringVar(&o.relayAddr, "relay", consts.DefaultRelayAddr, "relay server address")
	flag.BoolVar(&o.createRoom, "createroom", false, "create new punlic session")
	flag.StringVar(&o.joinRoom, "joinroom", "", "join public session by id")

	// Debugging flags.
	flag.StringVar(&o.cpuprof, "cpuprof", "", "write cpu profile to file")
	flag.StringVar(&o.memprof, "memprof", "", "write memory profile to file")
	flag.StringVar(&o.disasm, "disasm", "", "write cpu disassembly to file")
	flag.BoolVar(&o.verbose, "verbose", false, "enable verbose logging")

	flag.Parse()
	return o
}

func (o *options) sanitize() {
	if o.scale < 1 {
		o.scale = 1
	}
}

func (o *options) logLevel() loglevel.Level {
	if o.verbose {
		return loglevel.LevelDebug
	}

	return loglevel.LevelInfo
}

func printLogo() {
	// $ figlet "Dendy"
	fmt.Println(" ____                 _")
	fmt.Println("|  _ \\  ___ _ __   __| |_   _")
	fmt.Println("| | | |/ _ \\ '_ \\ / _` | | | |")
	fmt.Println("| |_| |  __/ | | | (_| | |_| |")
	fmt.Println("|____/ \\___|_| |_|\\__,_|\\__, |")
	fmt.Println("                        |___/")
}

func main() {
	opts := new(options).parse()
	opts.sanitize()

	log.Default().SetFlags(0)
	log.Default().SetOutput(loglevel.New(os.Stderr, opts.logLevel()))

	if flag.NArg() != 1 {
		fmt.Println("usage: dendy [-scale=2] [-nosave] [-nospritelimit] [-listen=addr:port] [-connect=addr:port] romfile")
		os.Exit(1)
	}

	if !opts.noLogo {
		printLogo()
	}

	if opts.cpuprof != "" {
		log.Printf("[INFO] writing cpu profile to %s", opts.cpuprof)

		f, err := os.Create(opts.cpuprof)
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

	if opts.memprof != "" {
		log.Printf("[INFO] writing memory profile to %s", opts.memprof)

		f, err := os.Create(opts.memprof)
		if err != nil {
			log.Printf("[ERROR] failed to create memory profile: %v", err)
			os.Exit(1)
		}

		defer func() {
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Printf("[ERROR] failed to write memory profile: %v", err)
			}

			_ = f.Close()
		}()
	}

	romFile := flag.Arg(0)
	log.Printf("[INFO] loading rom file: %s", romFile)

	rom, err := ines.NewFromFile(romFile)
	if err != nil {
		log.Printf("[ERROR] failed to open rom file: %s", err)
		os.Exit(1)
	}

	cart, err := ines.NewCartridge(rom)
	if err != nil {
		log.Printf("[ERROR] failed to open rom file: %s", err)
		os.Exit(1)
	}

	// Game Genie was a cartridge pass-through device, and we emulate
	// it as a cartridge pass-through device. How cool is that?
	if opts.gg != "" {
		gameGenie := genie.New(cart)
		codes := strings.Split(opts.gg, ",")

		for _, code := range codes {
			if err := gameGenie.ApplyCode(code); err != nil {
				log.Printf("[ERROR] failed to apply game genie code: %s", err)
				os.Exit(1)
			}
		}

		cart = gameGenie
	}

	saveFile := opts.saveFile
	romPrefix := strings.TrimSuffix(romFile, filepath.Ext(romFile))

	switch {
	case opts.connectAddr != "" || opts.joinRoom != "":
		log.Printf("[INFO] starting client mode")
		runAsClient(cart, opts, rom)

	case opts.listenAddr != "" || opts.createRoom:
		if saveFile == "" {
			saveFile = romPrefix + ".mp.save"
		}

		log.Printf("[INFO] starting host mode")
		runAsServer(cart, opts, saveFile, rom)

	default:
		if saveFile == "" {
			saveFile = romPrefix + ".save"
		}

		log.Printf("[INFO] starting offline mode")
		runOffline(cart, opts, saveFile)
	}
}
