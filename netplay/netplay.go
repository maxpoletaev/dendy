package netplay

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/maxpoletaev/dendy/internal/ringbuf"
)

const (
	maxFrameSyncFreq       = 300
	pingIntervalFrames     = 60
	initialFrameDriftLimit = 2
	maxFrameDriftLimit     = 20
)

type Netplay struct {
	game       *Game
	toRecv     chan Message
	toSend     chan Message
	conn       net.Conn
	rtt        time.Duration
	rttWindow  *ringbuf.Buffer[time.Duration]
	readerDone chan struct{}
	writerDone chan struct{}
	shouldExit bool

	syncFrame  uint32
	driftLimit uint32
}

func newNetplay(game *Game, conn net.Conn) *Netplay {
	return &Netplay{
		rttWindow:  ringbuf.New[time.Duration](10),
		toSend:     make(chan Message, 100),
		toRecv:     make(chan Message, 100),
		driftLimit: initialFrameDriftLimit,
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
	np.start()

	return np, conn.RemoteAddr(), nil
}

func Connect(game *Game, addr string) (*Netplay, net.Addr, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("netplay: failed to connect to %s: %v", addr, err)
	}

	np := newNetplay(game, conn)
	np.start()

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
			log.Printf("[ERROR] failed to write message: %v", err)
			np.shouldExit = true

			break
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

			log.Printf("[ERROR] failed to read message: %v", err)
			np.shouldExit = true

			break
		}

		np.toRecv <- msg
	}
}

func (np *Netplay) start() {
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

// HandleMessages handles incoming messages from the remote player.
func (np *Netplay) HandleMessages() {
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
}

// RunFrame progresses the game by one frame.
func (np *Netplay) RunFrame(startTime time.Time) {
	np.game.RunFrame(startTime)

	// Inject a ping message every N frames to measure latency.
	if np.game.Frame()%pingIntervalFrames == 0 {
		np.SendPing()
	}

	localFrame := np.game.Frame()
	remoteFrame := np.game.RemoteFrame()

	if localFrame < remoteFrame {
		drift := remoteFrame - localFrame

		// Ask the remote to wait if we are too far behind.
		if drift > np.driftLimit && np.syncFrame+maxFrameSyncFreq < localFrame {
			log.Printf("[INFO] asking the remote to wait for %d frames", drift)
			np.syncFrame = localFrame + uint32(rand.Int31n(maxFrameSyncFreq/10))
			np.SendWait(drift + 1) // +1 to account for the current frame

			// Gradually increase the drift limit to avoid oscillations.
			np.driftLimit = max(maxFrameDriftLimit, uint32(float32(drift)*1.25))
		}
	}
}

// RemotePing returns the ping time to the remote peer in milliseconds.
func (np *Netplay) RemotePing() int64 {
	return np.rtt.Milliseconds()
}
