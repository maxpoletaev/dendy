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
	case MsgTypeWait:
		np.handleWait(msg)
	default:
		// should never reach here
		panic(fmt.Errorf("unknown message type: %d", msg.Type))
	}
}

func (np *Netplay) handleWait(msg Message) {
	frames := byteOrder.Uint32(msg.Payload)

	log.Printf("[INFO] sleeping for %d frames", frames)

	np.syncFrame = np.game.Frame() + frames

	np.noDriftFrames = 0

	np.game.SleepFrames(frames)
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
	np.game.Init(&Checkpoint{
		Frame: msg.Frame,
		State: msg.Payload,
	})
}

func (np *Netplay) handlePing(msg Message) {
	np.sendMsg(Message{
		Type:       MsgTypePong,
		Generation: np.game.Gen(),
		Payload:    msg.Payload,
	})
}

func (np *Netplay) handlePong(msg Message) {
	timeSent := time.UnixMicro(int64(byteOrder.Uint64(msg.Payload)))
	np.rttWindow.PushBackEvict(time.Since(timeSent))

	var sum time.Duration
	for i := 0; i < np.rttWindow.Len(); i++ {
		sum += np.rttWindow.At(i)
	}

	np.rtt = sum / time.Duration(np.rttWindow.Len())

	np.game.SetRTT(np.rtt)
}

func (np *Netplay) handleInput(msg Message) {
	np.game.HandleRemoteInput(msg.Payload[0], msg.Frame)
}
