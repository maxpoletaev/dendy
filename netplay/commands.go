package netplay

import "log"

// SendInitialState is used by the server to send the initial state to the client.
func (np *Netplay) SendInitialState() {
	np.game.Init(nil)

	checkpoint := np.game.Checkpoint()
	state := make([]uint8, len(checkpoint.State))
	copy(state, checkpoint.State)
	frame := checkpoint.Frame

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypeReset,
		Frame:      frame,
		Payload:    state,
	})
}

// SendReset restarts the game on both sides.
func (np *Netplay) SendReset() {
	np.game.Reset()
	np.game.Init(nil)

	checkpoint := np.game.Checkpoint()
	state := make([]uint8, len(checkpoint.State))
	copy(state, checkpoint.State)
	frame := checkpoint.Frame

	np.sendMsg(Message{
		Generation: np.game.Gen(),
		Type:       MsgTypeReset,
		Frame:      frame,
		Payload:    state,
	})
}

// SendButtons sends the local input to the remote player. Should be called every frame.
func (np *Netplay) SendButtons(buttons uint8) {
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

// SendBye sends a bye message to the remote player when the game is over.
func (np *Netplay) SendBye() {
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
