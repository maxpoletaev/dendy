package netplay

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const (
	frameDuration      = time.Second / 60
	pingIntervalFrames = 60
)

type Netplay struct {
	game       *Game
	latency    time.Duration
	toRecv     chan Message
	toSend     chan Message
	pingSent   time.Time
	conn       net.Conn
	readerDone chan struct{}
	writerDone chan struct{}
	shouldExit bool
}

func newNetplay(game *Game, conn net.Conn) *Netplay {
	return &Netplay{
		toSend:     make(chan Message, 100),
		toRecv:     make(chan Message, 100),
		readerDone: make(chan struct{}),
		writerDone: make(chan struct{}),
		game:       game,
		conn:       conn,
	}
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

	np := newNetplay(game, conn)
	np.Start()

	return np, conn.RemoteAddr(), nil
}

func Connect(game *Game, addr string) (*Netplay, net.Addr, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("netplay: failed to connect to %s: %v", addr, err)
	}

	np := newNetplay(game, conn)

	np.Start()

	return np, conn.RemoteAddr(), nil
}

func (np *Netplay) startWriter() {
	defer close(np.writerDone)

	for {
		msg, ok := <-np.toSend
		if !ok {
			break
		}

		if err := writeMsg(np.conn, msg); err != nil {
			panic(fmt.Errorf("failed to write message: %v", err))
		}
	}
}

func (np *Netplay) startReader() {
	defer close(np.readerDone)

	for {
		msg, err := readMsg(np.conn)

		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				break
			}

			panic(fmt.Errorf("failed to read message: %v", err))
		}

		np.toRecv <- msg
	}
}

func (np *Netplay) Start() {
	go np.startReader()
	go np.startWriter()
}

func (np *Netplay) sendMsg(msg Message) {
	select {
	case np.toSend <- msg:
	default:
		log.Printf("[WARN] send buffer is full, blocking")
		np.toSend <- msg
	}
}

// ShouldExit indicates whether the game loop should exit.
func (np *Netplay) ShouldExit() bool {
	return np.shouldExit
}

// RunFrame runs a single frame of the game and handles any incoming messages.
func (np *Netplay) RunFrame() {
loop:
	for {
		select {
		case msg, ok := <-np.toRecv:
			if ok {
				np.handleMessage(msg)
			} else {
				break loop
			}
		default:
			break loop
		}
	}

	// Inject a ping message every N frames to measure latency.
	if np.game.Frame()%pingIntervalFrames == 0 {
		np.sendMsg(Message{
			Generation: np.game.Gen(),
			Type:       MsgTypePing,
		})

		np.pingSent = time.Now()
	}

	np.game.RunFrame()
}

// Latency returns the current latency in milliseconds between the players.
func (np *Netplay) Latency() int64 {
	return np.latency.Milliseconds()
}
