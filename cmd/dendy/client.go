package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/maxpoletaev/dendy/console"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/netplay"
	"github.com/maxpoletaev/dendy/ui"
)

func runAsClient(bus *console.Bus, o *opts) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()
	bus.InitDMA()
	bus.Reset()

	game := netplay.NewGame(bus)
	game.RemoteJoy = bus.Joy1
	game.LocalJoy = bus.Joy2
	game.Init(nil)

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

	log.Printf("[INFO] connecting to server...")
	sess, addr, err := netplay.Connect(game, o.connectAddr)

	if err != nil {
		log.Printf("[ERROR] failed to connect: %v", err)
		os.Exit(1)
	}

	log.Printf("[INFO] connected to server: %s", addr)
	log.Printf("[INFO] starting game...")

	w := ui.CreateWindow(&bus.PPU.Frame, o.scale, o.verbose)
	defer w.Close()

	w.SetTitle(fmt.Sprintf("%s (P2)", windowTitle))
	w.SetFrameRate(framesPerSecond)
	w.InputDelegate = sess.SendButtons
	w.ShowFPS = o.showFPS
	w.ShowPing = true

	for {
		startTime := time.Now()

		if w.ShouldClose() {
			log.Printf("[INFO] saying goodbye...")
			sess.SendBye()
			break
		}

		if sess.ShouldExit() {
			log.Printf("[INFO] server disconnected")
			break
		}

		w.HandleHotKeys()
		w.UpdateJoystick()
		w.SetGrayscale(game.Sleeping())
		w.SetPingInfo(sess.RemotePing())

		sess.HandleMessages()
		sess.RunFrame(startTime)

		w.Refresh()
	}
}
