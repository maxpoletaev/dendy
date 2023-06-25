package netplay

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/maxpoletaev/dendy/internal/rolling"
)

type Netplay struct {
	game       *Game
	toRecv     chan Message
	toSend     chan Message
	stop       chan struct{}
	inputBatch InputBatch
	remoteConn net.Conn
	batchSize  int
	ping       int
	pingJitter int
}

func Listen(game *Game, addr string, opts ...Options) (*Netplay, net.Addr, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("netplay: failed to listen on %s: %v", addr, err)
	}

	conn, err := listener.Accept()
	if err != nil {
		return nil, nil, fmt.Errorf("netplay: failed to accept connection: %v", err)
	}

	np := &Netplay{
		toSend:     make(chan Message, 100),
		toRecv:     make(chan Message, 100),
		stop:       make(chan struct{}),
		batchSize:  defaultInputBatch,
		game:       game,
		remoteConn: conn,
	}

	for _, opt := range opts {
		withOptions(np, opt)
	}

	return np, conn.RemoteAddr(), nil
}

func Connect(game *Game, addr string, opts ...Options) (*Netplay, net.Addr, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("netplay: failed to connect to %s: %v", addr, err)
	}

	np := &Netplay{
		toSend:     make(chan Message, 100),
		toRecv:     make(chan Message, 100),
		stop:       make(chan struct{}),
		batchSize:  defaultInputBatch,
		game:       game,
		remoteConn: conn,
	}

	for _, opt := range opts {
		withOptions(np, opt)
	}

	return np, conn.RemoteAddr(), nil
}

func (np *Netplay) simulateLatency() {
	if np.ping > 0 {
		sleep := time.Duration(np.ping) * time.Millisecond

		if np.pingJitter > 0 {
			sleep += time.Duration(rand.Intn(np.pingJitter)) * time.Millisecond
		}

		time.Sleep(sleep)
	}
}

func (np *Netplay) startWriter() {
	for {
		select {
		case <-np.stop:
			return
		case msg := <-np.toSend:
			np.simulateLatency()

			if err := writeMsg(np.remoteConn, msg); err != nil {
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
			msg, err := readMsg(np.remoteConn)
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
	if msg.Incarnation < np.game.Incarnation() {
		log.Printf("[INFO] dropping message from old incarnation: %d", msg.Incarnation)
		return
	}

	switch msg.Type {
	case MsgTypeReset:
		np.resetInputBatch(msg.Frame)
		np.game.Reset(&Checkpoint{
			Frame: msg.Frame,
			State: msg.Payload,
		})

	case MsgTypeInput:
		np.game.AddRemoteInput(InputBatch{
			Input:      msg.Payload,
			StartFrame: msg.Frame,
		})

		// If we're too far behind, ask the other side to wait for us.
		endFrame := np.game.Frame() + int32(np.batchSize*10)
		if rolling.GreaterThan(msg.Frame, endFrame) {
			log.Printf("[INFO] asking other side to sleep")

			np.sendMsg(Message{
				Incarnation: np.game.Incarnation(),
				Type:        MsgTypeSleep,
				Frame:       msg.Frame,
			})
		}

	case MsgTypeSleep:
		if d := int(msg.Frame - np.game.Frame()); d > 0 {
			log.Printf("[INFO] sleeping for %d frames", d)
			np.game.Sleep(d)
		}
	}

}

func (np *Netplay) Start() {
	go np.startReader()
	go np.startWriter()
}

func (np *Netplay) resetInputBatch(startFrame int32) {
	np.inputBatch = InputBatch{
		StartFrame: startFrame,
		Input:      make([]uint8, 0, np.batchSize),
	}
}

// SendReset restarts the game on both sides, should be called by the server once the
// game is ready to start to sync the initial state.
func (np *Netplay) SendReset() {
	np.game.Reset(nil)
	np.resetInputBatch(0)
	cp := np.game.Checkpoint()

	np.sendMsg(Message{
		Incarnation: np.game.Incarnation(),
		Type:        MsgTypeReset,
		Frame:       cp.Frame,
		Payload:     cp.State,
	})
}

// SendInput sends the local input to the remote player. Should be called every frame.
// The input is buffered and sent in batches to reduce the number of messages sent.
func (np *Netplay) SendInput(buttons uint8) {
	np.game.AddLocalInput(buttons)
	np.inputBatch.Add(buttons)

	if len(np.inputBatch.Input) >= np.batchSize {
		np.sendMsg(Message{
			Type:        MsgTypeInput,
			Payload:     np.inputBatch.Input,
			Frame:       np.inputBatch.StartFrame,
			Incarnation: np.game.Incarnation(),
		})

		np.inputBatch = InputBatch{
			StartFrame: np.game.Frame() + 1,
			Input:      make([]uint8, 0, np.batchSize),
		}
	}
}

func (np *Netplay) RunFrame() {
	select {
	case msg := <-np.toRecv:
		np.handleMessage(msg)
	default:
	}

	np.game.RunFrame()
}
