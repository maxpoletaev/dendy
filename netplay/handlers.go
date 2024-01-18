package netplay

import (
	"fmt"
	"log"
	"time"
)

func (np *Netplay) handleMessage(msg Message) {
	if msg.Generation < np.game.Gen() {
		log.Printf("[INFO] dropping message from old generation: %d", msg.Generation)
		return
	}

	switch msg.Type {
	case MsgTypeReset:
		np.handleReset(msg)
	case MsgTypePing:
		np.handlePing(msg)
	case MsgTypePong:
		np.handlePong(msg)
	case MsgTypeInput:
		np.handleInput(msg)
	case MsgTypeBye:
		np.handleBye(msg)
	default:
		// should never reach here
		panic(fmt.Errorf("unknown message type: %d", msg.Type))
	}
}

func (np *Netplay) handleBye(msg Message) {
	// The remote peer doesn't care about further messages.
	close(np.toSend)

	// Wait for the writer to drain the send buffer and stops.
	<-np.writerDone

	// Set the shouldExit flag to signal the game loop to exit.
	np.shouldExit = true

	// Close the connection. This will cause the reader to exit.
	if err := np.conn.Close(); err != nil {
		log.Printf("[ERROR] failed to close connection: %v", err)
	}

	// Clean up the reader.
	<-np.readerDone
	close(np.toRecv)
}

func (np *Netplay) handleReset(msg Message) {
	np.game.ResetState()

	np.game.SetCheckpoint(&Checkpoint{
		Frame: msg.Frame,
		State: msg.Payload,
	})
}

func (np *Netplay) handlePing(msg Message) {
	np.sendMsg(Message{
		Type:       MsgTypePong,
		Generation: np.game.Gen(),
	})
}

func (np *Netplay) handlePong(msg Message) {
	np.latency = time.Since(np.pingSent)
}

func (np *Netplay) handleInput(msg Message) {
	np.game.HandleRemoteInput(msg.Payload[0])
}
