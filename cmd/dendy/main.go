package main

import (
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
	"github.com/maxpoletaev/dendy/internal/loglevel"
)

const (
	windowTitle = "Dendy Emulator"
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
	dmcReverse    bool
	mute          bool
}

func (o *opts) parse() *opts {
	flag.IntVar(&o.scale, "scale", 2, "scale factor (default: 2)")
	flag.BoolVar(&o.noSpriteLimit, "nospritelimit", false, "disable sprite limit")
	flag.BoolVar(&o.dmcReverse, "dmcreverse", false, "reverse dmc samples")
	flag.StringVar(&o.connectAddr, "connect", "", "netplay connect address")
	flag.StringVar(&o.listenAddr, "listen", "", "netplay listen address)")
	flag.IntVar(&o.bufsize, "bufsize", 0, "netplay input buffer size (default: 0)")
	flag.BoolVar(&o.noSave, "nosave", false, "disable save states")
	flag.BoolVar(&o.showFPS, "showfps", false, "show fps counter")
	flag.BoolVar(&o.mute, "mute", false, "disable apu emulation")

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
	o := new(opts).parse()
	o.sanitize()

	log.Default().SetFlags(0)
	log.Default().SetOutput(loglevel.New(os.Stderr, o.logLevel()))

	if flag.NArg() != 1 {
		fmt.Println("usage: dendy [-scale=2] [-nosave] [-nospritelimit] [-listen=addr:port] [-connect=addr:port] romfile")
		os.Exit(1)
	}

	printLogo()

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
	apu.Enabled = !o.mute

	if apu.Enabled && (o.listenAddr != "" || o.connectAddr != "") {
		log.Printf("[WARN] sound is not yet supported in netplay mode")
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
