package main

import (
	_ "embed"
	"log"
	"runtime"
	"syscall/js"
	"time"
	"unsafe"

	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/netplay"
	"github.com/maxpoletaev/dendy/system"
)

//go:embed test.nes
var initialRomData []byte

func createSystem(joy1, joy2 *input.Joystick, romData []byte) *system.System {
	rom, err := ines.NewFromBuffer(romData)
	if err != nil {
		log.Fatalf("failed to load ROM: %v", err)
		return nil
	}

	cart, err := ines.NewCartridge(rom)
	if err != nil {
		log.Fatalf("failed to create cartridge: %v", err)
		return nil
	}

	//zapper := input.NewZapper()
	nes := system.New(cart, joy1, joy2)
	nes.DisableAPU()

	return nes
}

func main() {
	global := js.Global()
	localJoy := input.NewJoystick()
	remoteJoy := input.NewJoystick()

	var (
		nes        *system.System
		sess       *netplay.Netplay
		conn       *webrtcConn
		isHost     bool
		frameCount int
	)

	global.Set("uploadROM", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := js.Global().Get("Uint8Array").New(args[0])
		romData := make([]byte, data.Length())
		js.CopyBytesToGo(romData, data)
		nes = createSystem(localJoy, remoteJoy, romData)
		return nil
	}))

	global.Set("startSession", js.FuncOf(func(this js.Value, args []js.Value) any {
		conn = newWebrtcConn(args[0])
		isHost = args[1].Bool()

		if isHost {
			game := netplay.NewGame(nes, nil, localJoy, remoteJoy)
			game.Init(nil)
			sess = netplay.New(game, conn, true)
			sess.SendInitialState()
		} else {
			game := netplay.NewGame(nes, nil, remoteJoy, localJoy)
			game.Init(nil)
			sess = netplay.New(game, conn, false)
		}

		//runtime.GC()
		sess.Start()
		return nil
	}))

	global.Set("runFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		startTime := time.Now()

		frameBuf := args[0]
		buttons := args[1].Int()

		sess.SendButtons(uint8(buttons))
		sess.HandleMessages()
		sess.RunFrame(startTime)

		frame := nes.Frame()
		frameBytes := unsafe.Slice((*byte)(unsafe.Pointer(&frame[0])), len(frame)*4)
		js.CopyBytesToJS(frameBuf, frameBytes) // TODO: can we avoid copying here?

		frameCount++

		if frameCount%300 == 0 {
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)
			log.Printf("HeapAlloc: %d", stats.HeapAlloc)
		}

		return nil
	}))

	select {}
}
