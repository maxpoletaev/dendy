package netplay

import (
	"encoding/binary"
	"log"
	"time"
)

// SendInitialState is used by the server to send the initial state to the client.
func (np *Netplay) SendInitialState() {
	if np.game.Sleeping() {
		return
	}

	np.game.Init(nil)
	checkpoint := np.game.Checkpoint()

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypeReset,
		Frame:      checkpoint.Frame,
		Payload:    checkpoint.State,
	})
}

// SendReset restarts the game on both sides.
func (np *Netplay) SendReset() {
	if np.game.Sleeping() {
		return
	}

	np.game.Reset()
	np.game.Init(nil)
	checkpoint := np.game.Checkpoint()

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypeReset,
		Frame:      checkpoint.Frame,
		Payload:    checkpoint.State,
	})
}

func (np *Netplay) SendResync() {
	if np.game.Sleeping() {
		return
	}

	np.game.Init(nil)
	checkpoint := np.game.Checkpoint()

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypeReset,
		Frame:      checkpoint.Frame,
		Payload:    checkpoint.State,
	})
}

// SendButtons sends the local input to the remote player. Should be called every frame.
func (np *Netplay) SendButtons(buttons uint8) {
	if np.game.Sleeping() {
		return
	}

	if np.game.Frame() == 0 {
		return
	}

	np.sendMsg(Message{
		Type:       MsgTypeInput,
		Payload:    []uint8{buttons},
		Frame:      np.game.Frame(),
		Generation: np.game.Gen(),
	})

	np.game.HandleLocalInput(buttons)
}

func (np *Netplay) SendPing() {
	if np.game.Sleeping() {
		return
	}

	payload := make([]byte, 8)
	timestamp := time.Now().UnixMilli()
	binary.LittleEndian.PutUint64(payload, uint64(timestamp))

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypePing,
		Payload:    payload,
	})
}

// SendBye sends a bye message to the remote player when the game is over.
func (np *Netplay) SendBye() {
	if np.game.Sleeping() {
		return
	}

	np.sendMsg(Message{
		Type:       MsgTypeBye,
		Generation: np.game.Gen(),
	})

	// There should be no more messages sent after this,
	// so close the send channel to signal the writer to stop.
	close(np.toSend)

	// Wait for the writer to drain the send buffer,
	// so that the remote peer receives the bye message.
	<-np.writerDone

	// Close the connection. This will cause the reader to exit.
	if err := np.conn.Close(); err != nil {
		log.Printf("[ERROR] failed to close connection: %s", err)
	}

	// Clean up the reader.
	<-np.readerDone
	close(np.toRecv)
}

func (np *Netplay) SendWait(frames uint32) {
	if np.game.Sleeping() || frames == 0 {
		return
	}

	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, frames)

	np.sendMsg(Message{
		Type:       MsgTypeWait,
		Generation: np.game.Gen(),
		Payload:    payload,
	})
}
