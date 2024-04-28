package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/maxpoletaev/dendy/console"
	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/netplay"
	"github.com/maxpoletaev/dendy/relay"
	"github.com/maxpoletaev/dendy/ui"
)

func printSessionID(sessionID string) {
	const width = 20

	center := func(s string, width int) string {
		lPadding := width/2 - len(s)/2
		rPadding := width - len(s) - lPadding
		return strings.Repeat(" ", lPadding) + s + strings.Repeat(" ", rPadding)
	}

	fmt.Println(strings.Repeat("-", width+4))
	fmt.Println("|", center("Your Session ID:", width), "|")
	fmt.Println("|", center(sessionID, width), "|")
	fmt.Println(strings.Repeat("-", width+4))
}

func createSession(relayAddr string, romCRC32 uint32, public bool) (string, error) {
	log.Printf("[INFO] connecting to relay server: %s", relayAddr)

	relayClient, err := relay.Connect(relayAddr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to relay server: %w", err)
	}

	defer func() {
		if err := relayClient.Close(); err != nil {
			log.Printf("[ERROR] failed to close relay client: %s", err)
		}
	}()

	log.Printf("[INFO] creating session...")

	sessionID, err := relayClient.CreateSession(romCRC32, public)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	// Print the session ID to the console.
	printSessionID(sessionID)

	log.Printf("[INFO] waiting for the peer to join the session...")

	lAddr, rAddr, err := relayClient.GetPeerAddress()
	if err != nil {
		return "", fmt.Errorf("failed to get address: %w", err)
	}

	// Need to stop the relay client to free the port.
	if err := relayClient.Close(); err != nil {
		return "", fmt.Errorf("failed to close relay client: %w", err)
	}

	log.Printf("[INFO] peer joined: %s", rAddr.String())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := relay.HolePunchUDP(ctx, lAddr, rAddr); err != nil {
		return "", fmt.Errorf("failed to hole punch: %w", err)
	}

	return lAddr.String(), nil
}

func runAsServer(bus *console.Bus, o *opts, saveFile string) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()
	bus.InitDMA()
	bus.Reset()

	if !o.noSave {
		if ok, err := loadState(bus, saveFile); err != nil {
			log.Printf("[ERROR] failed to load save state: %s", err)
			os.Exit(1)
		} else if ok {
			log.Printf("[INFO] state loaded: %s", saveFile)
		}
	}

	audio := ui.CreateAudio(consts.SamplesPerSecond, consts.SampleSize, 1, consts.AudioBufferSize)
	defer audio.Close()
	audio.Mute(o.mute)

	game := netplay.NewGame(bus, audio)
	game.RemoteJoy = bus.Joy2
	game.LocalJoy = bus.Joy1
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

	var (
		err        error
		protocol   = o.protocol
		listenAddr = o.listenAddr
	)

	if o.createRoom {
		listenAddr, err = createSession(o.relayAddr, bus.ROM.CRC32, false)
		if err != nil {
			log.Printf("[ERROR] failed to create relay session: %s", err)
			os.Exit(1)
		}

		protocol = "udp" // relay is always UDP
	}

	log.Printf("[INFO] waiting for client to connect to %s (%s)...", listenAddr, protocol)
	sess, addr, err := netplay.Listen(protocol, listenAddr, game)

	if err != nil {
		log.Printf("[ERROR] failed to listen: %v", err)
		os.Exit(1)
	}

	log.Printf("[INFO] client connected: %s", addr)
	log.Printf("[INFO] starting game...")

	sess.SendInitialState()

	w := ui.CreateWindow(o.scale, o.verbose)
	defer w.Close()

	w.SetTitle(fmt.Sprintf("%s (P1)", windowTitle))
	w.SetFrameRate(consts.FramesPerSecond)
	w.ResyncDelegate = sess.SendResync
	w.InputDelegate = sess.SendButtons
	w.ResetDelegate = sess.SendReset
	w.MuteDelegate = audio.ToggleMute
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
			log.Printf("[INFO] client disconnected")
			break
		}

		w.HandleHotKeys()
		w.UpdateJoystick()
		w.SetPingInfo(sess.RemotePing())

		sess.HandleMessages()
		sess.RunFrame(startTime)

		w.Refresh(bus.PPU.Frame)
	}

	if !o.noSave {
		if err := saveState(bus, saveFile); err != nil {
			log.Printf("[ERROR] failed to save state: %s", err)
			os.Exit(1)
		}

		log.Printf("[INFO] state saved: %s", saveFile)
	}
}
