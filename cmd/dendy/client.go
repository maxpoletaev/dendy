package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/maxpoletaev/dendy/console"
	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/netplay"
	"github.com/maxpoletaev/dendy/relay"
	"github.com/maxpoletaev/dendy/ui"
)

func joinSession(relayAddr string, sessionID string, romCRC32 uint32) (string, string, error) {
	log.Printf("[INFO] connecting to relay server: %s", relayAddr)

	relayClient, err := relay.Connect(relayAddr)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to relay server: %w", err)
	}

	defer func() {
		if err := relayClient.Close(); err != nil {
			log.Printf("[ERROR] failed to close relay client: %s", err)
		}
	}()

	log.Printf("[INFO] joining session...")

	err = relayClient.JoinSession(sessionID, romCRC32)
	if err != nil {
		return "", "", fmt.Errorf("failed to join session: %w", err)
	}

	lAddr, rAddr, err := relayClient.GetPeerAddress()
	if err != nil {
		return "", "", fmt.Errorf("failed to get address: %w", err)
	}

	log.Printf("[INFO] peer joined: %s", rAddr.String())

	// Need to stop the relay client to free the port.
	if err := relayClient.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close relay client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	time.Sleep(1 * time.Second) // give the host some time to start up

	err = relay.HolePunchUDP(ctx, lAddr, rAddr)
	if err != nil {
		return "", "", fmt.Errorf("failed to hole punch: %w", err)
	}

	return lAddr.String(), rAddr.String(), nil
}

func runAsClient(bus *console.Bus, o *opts) {
	bus.Joy1 = input.NewJoystick()
	bus.Joy2 = input.NewJoystick()
	bus.InitDMA()
	bus.Reset()

	audio := ui.CreateAudio(consts.SamplesPerSecond, consts.SampleSize, 1, consts.AudioBufferSize)
	defer audio.Close()
	audio.Mute(o.mute)

	game := netplay.NewGame(bus, audio)
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

	var (
		err      error
		protocol = o.protocol
		rAddr    = o.connectAddr
		lAddr    = o.listenAddr
	)

	if o.joinRoom != "" {
		lAddr, rAddr, err = joinSession(o.relayAddr, o.joinRoom, bus.ROM.CRC32)
		if err != nil {
			log.Printf("[ERROR] failed to join relay session: %s", err)
			os.Exit(1)
		}

		protocol = "udp" // always use UDP for relay
	}

	if rAddr == "" {
		log.Printf("[ERROR] no host address provided")
		os.Exit(1)
	}

	log.Printf("[INFO] connecting to %s (%s)...", rAddr, protocol)
	sess, addr, err := netplay.Connect(protocol, rAddr, lAddr, game)

	if err != nil {
		log.Printf("[ERROR] failed to connect: %v", err)
		os.Exit(1)
	}

	log.Printf("[INFO] connected to server: %s", addr)
	log.Printf("[INFO] starting game...")

	w := ui.CreateWindow(o.scale, o.verbose)
	defer w.Close()

	w.SetTitle(fmt.Sprintf("%s (P2)", windowTitle))
	w.SetFrameRate(consts.FramesPerSecond)
	w.InputDelegate = sess.SendButtons
	w.MuteDelegate = audio.ToggleMute
	w.ShowFPS = o.showFPS
	w.ShowPing = true

	if !o.noCRT {
		log.Printf("[INFO] using experimental CRT effect, disable with -nocrt flag")
		w.EnableCRT()
	}

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
		w.SetPingInfo(sess.RemotePing())

		sess.HandleMessages()
		sess.RunFrame(startTime)

		w.Refresh(bus.PPU.Frame)
	}
}
