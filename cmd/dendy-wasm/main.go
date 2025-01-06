package main

import (
	_ "embed"
	"log"
	"syscall/js"
	"unsafe"

	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/system"
)

//go:embed nestest.nes
var initialRomData []byte

func createSystem(joy *input.Joystick, romData []byte) *system.System {
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

	zapper := input.NewZapper()
	nes := system.New(cart, joy, zapper)

	return nes
}

func main() {
	joy1 := input.NewJoystick()
	nes := createSystem(joy1, initialRomData)

	js.Global().Set("runFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		frameBuf := args[0]
		buttons := args[1].Int()

		for {
			nes.Tick()
			if nes.FrameReady() {
				frame := nes.Frame()
				frameBytes := unsafe.Slice((*byte)(unsafe.Pointer(&frame[0])), len(frame)*4)
				js.CopyBytesToJS(frameBuf, frameBytes) // TODO: can we avoid copying here?
				joy1.SetButtons(uint8(buttons))
				break
			}
		}

		return nil
	}))

	js.Global().Set("uploadROM", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := js.Global().Get("Uint8Array").New(args[0])
		romData := make([]byte, data.Length())
		js.CopyBytesToGo(romData, data)
		nes = createSystem(joy1, romData)
		return nil
	}))

	js.Global().Set("getAudioSample", js.FuncOf(func(this js.Value, args []js.Value) any {
		return nes.AudioSample()
	}))

	select {}
}
