package netplay

import (
	"log"
	"time"
)

// SendInitialState is used by the server to send the initial state to the client.
func (np *Netplay) SendInitialState() {
	np.game.Init(nil)
	cp := np.game.syncState
	payload := np.pool.Buffer(cp.state.Len())
	copy(payload.Data, cp.state.Bytes())

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypeReset,
		Frame:      cp.frame,
		Buffer:     payload,
	})
}

// SendReset restarts the game on both sides.
func (np *Netplay) SendReset() {
	if np.game.Sleeping() {
		return
	}

	np.game.Reset()
	np.game.Init(nil)

	cp := np.game.syncState
	payload := np.pool.Buffer(cp.state.Len())
	copy(payload.Data, cp.state.Bytes())

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypeReset,
		Frame:      cp.frame,
		Buffer:     payload,
	})
}

func (np *Netplay) SendResync() {
	if np.game.Sleeping() {
		return
	}

	np.game.Init(nil)
	cp := np.game.syncState
	payload := np.pool.Buffer(cp.state.Len())
	copy(payload.Data, cp.state.Bytes())

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypeReset,
		Frame:      cp.frame,
		Buffer:     payload,
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

	buf := np.pool.Buffer(1)
	buf.Data[0] = buttons

	np.sendMsg(Message{
		Type:       MsgTypeInput,
		Frame:      np.game.Frame(),
		Generation: np.game.Gen(),
		Buffer:     buf,
	})

	np.game.HandleLocalInput(buttons)
}

func (np *Netplay) SendPing() {
	if np.game.Sleeping() {
		return
	}

	buf := np.pool.Buffer(8)
	timestamp := time.Now().UnixMicro()
	byteOrder.PutUint64(buf.Data, uint64(timestamp))

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypePing,
		Buffer:     buf,
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

	buf := np.pool.Buffer(4)
	byteOrder.PutUint32(buf.Data, frames)

	np.sendMsg(Message{
		Type:       MsgTypeWait,
		Generation: np.game.Gen(),
		Buffer:     buf,
	})
}
