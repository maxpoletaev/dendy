package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/maxpoletaev/dendy/console"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/ui"
)

const (
	sampleSize        = 32
	framesPerSecond   = 60
	samplesPerSecond  = 44100
	cpuTicksPerSecond = 1789773
	ticksPerSecond    = cpuTicksPerSecond * 3
	samplesPerFrame   = samplesPerSecond / framesPerSecond
	ticksPerSample    = ticksPerSecond / samplesPerSecond
	audioBufferSize   = samplesPerFrame * 3
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
	bus.InitDMA()
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

	audio := ui.CreateAudio(samplesPerSecond, sampleSize, 1, audioBufferSize)
	audioBuffer := make([]float32, audioBufferSize)
	defer audio.Close()

	defer func() {
		if err := recover(); err != nil {
			// Save state on crash to quickly reconstruct the faulty state.
			_ = saveState(bus, fmt.Sprintf("%s.crash", saveFile))
			panic(err)
		}
	}()

gameloop:
	for {
		for i := 0; i < audioBufferSize; i++ {
			for j := 0; j < ticksPerSample; j++ {
				bus.Tick()

				if bus.ScanlineComplete() {
					w.UpdateZapper()
				}

				if bus.FrameComplete() {
					if w.ShouldClose() {
						break gameloop
					}

					bus.Zapper.VBlank()
					w.UpdateJoystick()
					w.HandleHotKeys()
					w.Refresh()

					// Pause when not in focus.
					for !w.InFocus() {
						if w.ShouldClose() {
							break gameloop
						}

						w.Refresh()
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
