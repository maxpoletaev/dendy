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

func createSystem(joy *input.Joystick, romData []byte) (*system.System, error) {
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
	log.SetFlags(0)
	joystick := input.NewJoystick()

	nes, err := createSystem(joystick, bootROM)
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize: %v", err)
	}

	var (
		mem          runtime.MemStats
		frameTimeSum time.Duration
		frameCount   uint
	)

	js.Global().Set("runFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		start := time.Now()
		frameBuf := args[0]
		buttons := args[1].Int()

		for {
			nes.Tick()
			if nes.FrameReady() {
				frame := nes.Frame()
				frameBytes := unsafe.Slice((*byte)(unsafe.Pointer(&frame[0])), len(frame)*4)
				js.CopyBytesToJS(frameBuf, frameBytes) // TODO: can we avoid copying here?
				joystick.SetButtons(uint8(buttons))
				break
			}
		}

		if debugFrameTime {
			frameTimeSum += time.Since(start)
			frameCount++

			if frameCount%120 == 0 {
				runtime.ReadMemStats(&mem)
				elapsed := frameTimeSum / time.Duration(frameCount)
				log.Printf("[DEBUG] frame time: %s, memory: %d", elapsed, mem.Alloc)
				frameTimeSum = 0
				frameCount = 0
			}
		}

		return nil
	}))

	js.Global().Set("uploadROM", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := js.Global().Get("Uint8Array").New(args[0])
		romData := make([]byte, data.Length())
		js.CopyBytesToGo(romData, data)

		nes2, err := createSystem(joystick, romData)
		if err != nil {
			log.Printf("[ERROR] failed to initialize: %v", err)
			return false
		}

		nes = nes2
		return true
	}))

	select {}
}
