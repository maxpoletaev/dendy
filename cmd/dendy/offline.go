package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/maxpoletaev/dendy/console"
	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/binario"
	"github.com/maxpoletaev/dendy/ui"
)

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

	reader := binario.NewReader(f, binary.LittleEndian)
	if err := bus.LoadState(reader); err != nil {
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

	writer := binario.NewWriter(f, binary.LittleEndian)
	if err := bus.SaveState(writer); err != nil {
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
	bus.InitDMA()
	bus.Reset()

	if o.disasm != "" {
		file, err := os.Create(o.disasm)
		if err != nil {
			log.Printf("[ERROR] failed to create disassembly file: %s", err)
			os.Exit(1)
		}

		writer := bufio.NewWriterSize(file, 1024*1024)

		defer func() {
			flushErr := writer.Flush()
			closeErr := file.Close()

			if err := errors.Join(flushErr, closeErr); err != nil {
				log.Printf("[ERROR] failed to close disassembly file: %s", err)
			}
		}()

		bus.DisasmWriter = writer
		bus.DisasmEnabled = true
	}

	if !o.noSave {
		if ok, err := loadState(bus, saveFile); err != nil {
			log.Printf("[ERROR] failed to load save file: %s", err)
			os.Exit(1)
		} else if ok {
			log.Printf("[INFO] state loaded: %s", saveFile)
		}

		if strings.HasSuffix(saveFile, ".crash") {
			log.Printf("[INFO] loaded from crash state, further saves disabled")
			o.noSave = true
		}
	}

	audio := ui.CreateAudio(consts.SamplesPerSecond, consts.SampleSize, 1, consts.AudioBufferSize)
	audioBuffer := make([]float32, consts.AudioBufferSize)
	audio.Mute(o.mute)
	defer audio.Close()

	w := ui.CreateWindow(o.scale, o.verbose)
	defer w.Close()

	w.SetFrameRate(consts.FramesPerSecond)
	w.SetTitle(windowTitle)
	w.InputDelegate = bus.Joy1.SetButtons
	w.ZapperDelegate = bus.Zapper.Update
	w.MuteDelegate = audio.ToggleMute
	w.ResetDelegate = bus.Reset
	w.ShowFPS = o.showFPS

	if !o.noCRT {
		log.Printf("[INFO] using experimental CRT effect, disable with -nocrt flag")
		w.EnableCRT()
	}

	defer func() {
		if err := recover(); err != nil {
			// Save state on crash to quickly reconstruct the faulty state,
			// unless we are already playing the crash state.
			if !strings.HasSuffix(saveFile, ".crash") {
				_ = saveState(bus, fmt.Sprintf("%s.crash", saveFile))
				log.Printf("[INFO] pre-crash state saved: %s.crash", saveFile)
			}

			panic(err)
		}
	}()

gameloop:
	for {
		for i := 0; i < consts.AudioBufferSize; i++ {
			for j := 0; j < consts.TicksPerSample; j++ {
				bus.Tick()

				if bus.ScanlineComplete() {
					w.UpdateZapper(bus.PPU.Frame)
				}

				if bus.FrameComplete() {
					if w.ShouldClose() {
						break gameloop
					}

					bus.Zapper.VBlank()
					w.UpdateJoystick()
					w.HandleHotKeys()
					w.SetGrayscale(false)
					w.Refresh(bus.PPU.Frame)

					// Pause when not in focus.
					for !w.InFocus() {
						if w.ShouldClose() {
							break gameloop
						}

						w.SetGrayscale(true)
						w.Refresh(bus.PPU.Frame)
					}
				}
			}

			audioBuffer[i] = bus.APU.Output()
		}

		audio.WaitStreamProcessed()
		audio.UpdateStream(audioBuffer)
	}

	if !o.noSave {
		if err := saveState(bus, saveFile); err != nil {
			log.Printf("[ERROR] failed to save state: %s", err)
			os.Exit(1)
		}

		log.Printf("[INFO] state saved: %s", saveFile)
	}
}
