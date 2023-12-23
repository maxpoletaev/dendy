package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/maxpoletaev/dendy/console"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/netplay"
	"github.com/maxpoletaev/dendy/ui"
)

func runAsServer(bus *console.Bus, o *opts) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()
	bus.InitDMA()

	game := netplay.NewGame(bus)
	game.BufferSize = o.bufsize
	game.RemoteJoy = bus.Joy2
	game.LocalJoy = bus.Joy1
	game.Reset(nil)

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
		bus.DisasmEnabled = false // will be controlled by the game
		game.DisasmEnabled = true
	}

	log.Printf("[INFO] waiting for client...")
	sess, addr, err := netplay.Listen(game, o.listenAddr)

	if err != nil {
		log.Printf("[ERROR] failed to listen: %v", err)
		os.Exit(1)
	}

	log.Printf("[INFO] client connected: %s", addr)
	log.Printf("[INFO] starting game...")

	w := ui.CreateWindow(&bus.PPU.Frame, o.scale, o.verbose)
	w.SetTitle(fmt.Sprintf("%s (P1)", windowTitle))
	w.SetFrameRate(framesPerSecond)
	w.InputDelegate = sess.SendButtons
	w.ResetDelegate = sess.SendReset
	w.ShowFPS = o.showFPS
	w.ShowPing = true

	sess.SendReset()
	sess.Start()

	for {
		if w.ShouldClose() {
			break
		}

		w.SetLatencyInfo(sess.Latency())
		w.HandleHotKeys()
		w.UpdateJoystick()

		sess.RunFrame()

		w.Refresh()
	}
}
