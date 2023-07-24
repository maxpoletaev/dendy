package netplay

import (
	"fmt"
	"log"
	"net"
)

type Netplay struct {
	game   *Game
	toRecv chan Message
	toSend chan Message
	stop   chan struct{}
	conn   net.Conn
}

func Listen(game *Game, addr string) (*Netplay, net.Addr, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("netplay: failed to listen on %s: %v", addr, err)
	}

	conn, err := listener.Accept()
	if err != nil {
		return nil, nil, fmt.Errorf("netplay: failed to accept connection: %v", err)
	}

	np := &Netplay{
		toSend: make(chan Message, 100),
		toRecv: make(chan Message, 100),
		stop:   make(chan struct{}),
		game:   game,
		conn:   conn,
	}

	return np, conn.RemoteAddr(), nil
}

func Connect(game *Game, addr string) (*Netplay, net.Addr, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("netplay: failed to connect to %s: %v", addr, err)
	}

	np := &Netplay{
		toSend: make(chan Message, 100),
		toRecv: make(chan Message, 100),
		stop:   make(chan struct{}),
		game:   game,
		conn:   conn,
	}

	return np, conn.RemoteAddr(), nil
}

func (np *Netplay) startWriter() {
	for {
		select {
		case <-np.stop:
			return
		case msg := <-np.toSend:
			if err := writeMsg(np.conn, msg); err != nil {
				panic(fmt.Errorf("failed to write message: %v", err))
			}
		}
	}
}

func (np *Netplay) startReader() {
	for {
		select {
		case <-np.stop:
			return
		default:
			msg, err := readMsg(np.conn)
			if err != nil {
				panic(fmt.Errorf("failed to read message: %v", err))
			}

			np.toRecv <- msg
		}
	}
}

func (np *Netplay) sendMsg(msg Message) {
	select {
	case np.toSend <- msg:
	default:
		log.Printf("[WARN] send buffer is full, blocking")
		np.toSend <- msg
	}
}

func (np *Netplay) handleMessage(msg Message) {
	if msg.Generation < np.game.Generation() {
		log.Printf("[INFO] dropping message from old generation: %d", msg.Generation)
		return
	}

	switch msg.Type {
	case MsgTypeReset:
		np.game.Reset(&Checkpoint{
			Frame: msg.Frame,
			State: msg.Payload,
		})
	case MsgTypeInput:
		np.game.HandleRemoteInput(PlayerInput{
			Buttons: msg.Payload[0],
			Frame:   msg.Frame,
		})
	}
}

func (np *Netplay) Start() {
	go np.startReader()
	go np.startWriter()
}

// SendReset restarts the game on both sides, should be called by the server once the
// game is ready to start to sync the initial state.
func (np *Netplay) SendReset() {
	np.game.Reset(nil)
	cp := np.game.Checkpoint()

	np.sendMsg(Message{
		Generation: np.game.Generation(),
		Type:       MsgTypeReset,
		Frame:      cp.Frame,
		Payload:    cp.State,
	})
}

// SendButtons sends the local input to the remote player. Should be called every frame.
// The input is buffered and sent in batches to reduce the number of messages sent.
func (np *Netplay) SendButtons(buttons uint8) {
	if np.game.Frame() == 0 {
		return
	}

	np.sendMsg(Message{
		Type:       MsgTypeInput,
		Payload:    []uint8{buttons},
		Frame:      np.game.Frame(),
		Generation: np.game.Generation(),
	})

	np.game.HandleLocalInput(buttons)
}

func (np *Netplay) RunFrame() {
loop:
	for {
		select {
		case msg := <-np.toRecv:
			np.handleMessage(msg)
		default:
			break loop
		}
	}

	np.game.RunFrame()
}
