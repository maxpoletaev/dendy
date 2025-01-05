package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/maxpoletaev/dendy/consts"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/netplay"
	"github.com/maxpoletaev/dendy/relay"
	"github.com/maxpoletaev/dendy/system"
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

func runAsClient(cart ines.Cartridge, opts *options, rom *ines.ROM) {
	joy1 := input.NewJoystick()
	joy2 := input.NewJoystick()

	nes := system.New(cart, joy1, joy2)
	nes.SetNoSpriteLimit(opts.noSpriteLimit)

	audio := ui.CreateAudio(consts.AudioSamplesPerSecond, consts.AudioSampleSize, 1, consts.AudioBufferSize)
	defer audio.Close()
	audio.Mute(opts.mute)

	game := netplay.NewGame(nes, audio, joy2, joy1)
	game.Init(nil)

	if opts.disasm != "" {
		file, err := os.Create(opts.disasm)
		if err != nil {
			log.Printf("[ERROR] failed to create disassembly file: %s", err)
			os.Exit(1)
		}

		writer := bufio.NewWriterSize(file, 1024*1024)
		game.SetDebugOutput(writer)

		defer func() {
			flushErr := writer.Flush()
			closeErr := file.Close()

			if err := errors.Join(flushErr, closeErr); err != nil {
				log.Printf("[ERROR] failed to close disassembly file: %s", err)
			}
		}()
	}

	var (
		err      error
		protocol = opts.protocol
		rAddr    = opts.connectAddr
		lAddr    = opts.listenAddr
	)

	if opts.joinRoom != "" {
		lAddr, rAddr, err = joinSession(opts.relayAddr, opts.joinRoom, rom.CRC32)
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

	win := ui.CreateWindow(opts.scale, opts.verbose)
	defer win.Close()

	win.SetTitle(fmt.Sprintf("%s (P2)", windowTitle))
	win.SetFrameRate(consts.FramesPerSecond)
	win.InputDelegate = sess.SendButtons
	win.MuteDelegate = audio.ToggleMute
	win.ShowFPS = opts.showFPS
	win.ShowPing = true

	if !opts.noCRT {
		log.Printf("[INFO] using experimental CRT effect, disable with -nocrt flag")
		win.EnableCRT()
	}

	for {
		startTime := time.Now()

		if win.ShouldClose() {
			log.Printf("[INFO] saying goodbye...")
			sess.SendBye()
			break
		}

		if sess.ShouldExit() {
			log.Printf("[INFO] server disconnected")
			break
		}

		win.HandleHotKeys()
		win.UpdateJoystick()
		win.SetPingInfo(sess.RemotePing())

		sess.HandleMessages()
		sess.RunFrame(startTime)

		win.Refresh(nes.Frame())
	}
}
