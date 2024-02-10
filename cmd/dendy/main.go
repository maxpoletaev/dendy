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
	"github.com/maxpoletaev/dendy/consts"
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
	saveFile      string
	noSave        bool
	showFPS       bool
	verbose       bool
	disasm        string
	cpuprof       string
	protocol      string
	mute          bool
	noLogo        bool

	connectAddr string
	listenAddr  string
	relayAddr   string
	joinRoom    string
	createRoom  bool
}

func (o *opts) parse() *opts {
	flag.IntVar(&o.scale, "scale", 2, "scale factor (default: 2)")
	flag.StringVar(&o.saveFile, "savefile", "", "save file (default: romname.save)")
	flag.BoolVar(&o.noSpriteLimit, "nospritelimit", false, "disable sprite limit (eliminates flickering)")
	flag.BoolVar(&o.noSave, "nosave", false, "disable save states")
	flag.BoolVar(&o.showFPS, "showfps", false, "show fps counter")
	flag.BoolVar(&o.mute, "mute", false, "disable apu emulation")
	flag.BoolVar(&o.noLogo, "nologo", false, "do not print logo")

	flag.StringVar(&o.protocol, "protocol", "tcp", "netplay protocol (tcp, udp)")
	flag.StringVar(&o.listenAddr, "listen", "", "netplay listen address")
	flag.StringVar(&o.connectAddr, "connect", "", "netplay connect address")
	flag.StringVar(&o.relayAddr, "relay", consts.DefaultRelayAddr, "relay server address")
	flag.BoolVar(&o.createRoom, "createroom", false, "create new punlic session")
	flag.StringVar(&o.joinRoom, "joinroom", "", "join public session by id")

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

	if !o.noLogo {
		printLogo()
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

	rom, err := ines.OpenROM(romFile)
	if err != nil {
		log.Printf("[ERROR] failed to open rom file: %s", err)
		os.Exit(1)
	}

	cart, err := ines.NewCartridge(rom)
	if err != nil {
		log.Printf("[ERROR] failed to open rom file: %s", err)
		os.Exit(1)
	}

	log.Printf("[INFO] loaded rom: mapper:%d crc32:%08X", rom.MapperID, rom.CRC32)

	apu := apupkg.New()
	cpu := cpupkg.New()
	ppu := ppupkg.New(cart)
	ppu.NoSpriteLimit = o.noSpriteLimit

	bus := &console.Bus{
		ROM:  rom,
		Cart: cart,
		CPU:  cpu,
		PPU:  ppu,
		APU:  apu,
	}

	saveFile := o.saveFile
	romPrefix := strings.TrimSuffix(romFile, filepath.Ext(romFile))

	switch {
	case o.connectAddr != "" || o.joinRoom != "":
		log.Printf("[INFO] starting client mode")
		runAsClient(bus, o)

	case o.listenAddr != "" || o.createRoom:
		if saveFile == "" {
			saveFile = romPrefix + ".mp.save"
		}

		log.Printf("[INFO] starting host mode")
		runAsServer(bus, o, saveFile)

	default:
		if saveFile == "" {
			saveFile = romPrefix + ".save"
		}

		log.Printf("[INFO] starting offline mode")
		runOffline(bus, o, saveFile)
	}
}
