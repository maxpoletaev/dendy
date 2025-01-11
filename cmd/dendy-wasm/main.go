package main

import (
	_ "embed"
	"fmt"
	"log"
	"syscall/js"
	"unsafe"

	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/system"
)

const (
	audioBufferSize = 1024 // must be multiple 128 as JS consumes samples in 128 chunks
)

//go:embed nestest.nes
var nestestROM []byte

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
	audioBuf := make([]float32, audioBufferSize)

	nes, err := create(joystick, nestestROM)
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize: %v", err)
	}

	global := js.Global()
	jsapi := global.Get("go")
	jsapi.Set("AudioBufferSize", audioBufferSize)
	jsapi.Set("AudioSampleRate", consts.AudioSamplesPerSecond)

	var (
		ticksCount  int
		sampleCount int
	)

	jsapi.Set("RunFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		buttons := args[0].Int()
		frameReady := false

		for sampleCount < len(audioBuf) {
			for ticksCount < consts.TicksPerAudioSample {
				nes.Tick()
				ticksCount++

				if nes.FrameReady() {
					joystick.SetButtons(uint8(buttons))
					frameReady = true
				}
			}

			audioBuf[sampleCount] = nes.AudioSample()
			sampleCount++
			ticksCount = 0

			if frameReady {
				return true
			}
		}

		sampleCount = 0
		return false
	}))

	jsapi.Set("GetFrameBufferPtr", js.FuncOf(func(this js.Value, args []js.Value) any {
		return uintptr(unsafe.Pointer(&nes.Frame()[0]))
	}))

	jsapi.Set("GetAudioBufferPtr", js.FuncOf(func(this js.Value, args []js.Value) any {
		return uintptr(unsafe.Pointer(&audioBuf[0]))
	}))

	jsapi.Set("LoadROM", js.FuncOf(func(this js.Value, args []js.Value) any {
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
