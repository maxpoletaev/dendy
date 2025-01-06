package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/internal/binario"
	"github.com/maxpoletaev/dendy/system"
	"github.com/maxpoletaev/dendy/ui"
)

func loadState(nes *system.System, saveFile string) (bool, error) {
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
	if err := nes.LoadState(reader); err != nil {
		return false, err
	}

	return true, nil
}

func saveState(nes *system.System, saveFile string) error {
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
	if err := nes.SaveState(writer); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, saveFile); err != nil {
		return err
	}

	return nil
}

func runOffline(cart ines.Cartridge, opts *options, saveFile string) {
	joy1 := input.NewJoystick()
	zapper := input.NewZapper()

	nes := system.New(cart, joy1, zapper)
	nes.SetNoSpriteLimit(opts.noSpriteLimit)
	nes.SetRewindEnabled(true)

	if opts.disasm != "" {
		var file io.Writer

		if opts.disasm == "-" {
			file = os.Stdout
		} else {
			f, err := os.Create(opts.disasm)
			if err != nil {
				log.Printf("[ERROR] failed to create disassembly file: %s", err)
				os.Exit(1)
			}

			file = f
		}

		writer := bufio.NewWriterSize(file, 1024*1024)
		nes.SetDebugWriter(writer)

		defer func() {
			_ = writer.Flush()

			if f, ok := file.(*os.File); ok {
				_ = f.Close()
			}
		}()
	}

	if !opts.noSave {
		if ok, err := loadState(nes, saveFile); err != nil {
			log.Printf("[ERROR] failed to load save file: %s", err)
			os.Exit(1)
		} else if ok {
			log.Printf("[INFO] state loaded: %s", saveFile)
		}

		if strings.HasSuffix(saveFile, ".crash") {
			log.Printf("[INFO] loaded from crash state, further saves disabled")
			opts.noSave = true
		}
	}

	w := ui.CreateWindow(opts.scale, opts.verbose)
	defer w.Close()

	audio := ui.CreateAudio(consts.AudioSamplesPerSecond, consts.AudioSampleSize, 1, consts.AudioBufferSize)
	audioBuffer := make([]float32, consts.AudioBufferSize)
	audio.Mute(opts.mute)
	defer audio.Close()

	w.SetFrameRate(consts.FramesPerSecond)
	w.SetTitle(windowTitle)

	w.InputDelegate = joy1.SetButtons
	w.ZapperDelegate = zapper.Update
	w.MuteDelegate = audio.ToggleMute
	w.RewindDelegate = nes.Rewind
	w.ResetDelegate = nes.Reset
	w.ShowFPS = opts.showFPS

	if !opts.noCRT {
		log.Printf("[INFO] using experimental CRT effect, disable with -nocrt flag")
		w.EnableCRT()
	}

	defer func() {
		if err := recover(); err != nil {
			// Save state on crash to quickly reconstruct the faulty state,
			// unless we are already playing the crash state.
			if !strings.HasSuffix(saveFile, ".crash") {
				_ = saveState(nes, fmt.Sprintf("%s.crash", saveFile))
				log.Printf("[INFO] pre-crash state saved: %s.crash", saveFile)
			}

			panic(err)
		}
	}()

gameloop:
	for {
		for i := 0; i < consts.AudioBufferSize; i++ {
			for j := 0; j < consts.TicksPerAudioSample; j++ {
				nes.Tick()

				if nes.ScanlineReady() {
					w.UpdateZapper(nes.Frame())
				}

				if nes.FrameReady() {
					if w.ShouldClose() {
						break gameloop
					}

					zapper.VBlank()

					w.UpdateJoystick()
					w.HandleHotKeys()
					w.SetGrayscale(false)
					w.Refresh(nes.Frame())

					// Pause when not in focus.
					for !w.InFocus() {
						if w.ShouldClose() {
							break gameloop
						}

						w.SetGrayscale(true)
						w.Refresh(nes.Frame())
					}
				}
			}

			audioBuffer[i] = nes.AudioSample()
		}

		audio.WaitStreamProcessed()
		audio.UpdateStream(audioBuffer)
	}

	if !opts.noSave {
		if err := saveState(nes, saveFile); err != nil {
			log.Printf("[ERROR] failed to save state: %s", err)
			os.Exit(1)
		}

		log.Printf("[INFO] state saved: %s", saveFile)
	}
}
