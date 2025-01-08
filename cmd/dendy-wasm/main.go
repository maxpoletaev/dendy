package main

import (
	_ "embed"
	"fmt"
	"log"
	"runtime"
	"syscall/js"
	"time"
	"unsafe"

	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/system"
)

const (
	debugFrameTime = false
)

//go:embed nestest.nes
var bootROM []byte

func create(joy *input.Joystick, romData []byte) (*system.System, error) {
	rom, err := ines.NewFromBuffer(romData)
	if err != nil {
		return nil, fmt.Errorf("failed to load ROM: %v", err)
	}

	cart, err := ines.NewCartridge(rom)
	if err != nil {
		return nil, fmt.Errorf("failed to create cartridge: %v", err)
	}

	zapper := input.NewZapper()
	nes := system.New(cart, joy, zapper)
	return nes, nil
}

func main() {
	log.SetFlags(0) // disable timestamps
	joystick := input.NewJoystick()

	nes, err := create(joystick, bootROM)
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize: %v", err)
	}

	var (
		mem          runtime.MemStats
		frameTimeSum time.Duration
		frameCount   uint
	)

	js.Global().Set("runFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		buttons := args[0].Int()
		start := time.Now()
		var framePtr uintptr

		for {
			nes.Tick()
			if nes.FrameReady() {
				frame := nes.Frame()
				joystick.SetButtons(uint8(buttons))
				framePtr = uintptr(unsafe.Pointer(&frame[0]))
				break
			}
		}

		if debugFrameTime {
			frameTime := time.Since(start)
			frameTimeSum += frameTime
			frameCount++

			if frameCount%120 == 0 {
				runtime.ReadMemStats(&mem)
				avgFrameTime := frameTimeSum / time.Duration(frameCount)
				log.Printf("[INFO] frame time: %v, memory: %d", avgFrameTime, mem.HeapAlloc)
				frameTimeSum = 0
				frameCount = 0
			}
		}

		return framePtr
	}))

	js.Global().Set("uploadROM", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := js.Global().Get("Uint8Array").New(args[0])
		romData := make([]byte, data.Length())
		js.CopyBytesToGo(romData, data)

		nes2, err := create(joystick, romData)
		if err != nil {
			log.Printf("[ERROR] failed to initialize: %v", err)
			return false
		}

		nes = nes2
		return true
	}))

	select {}
}
