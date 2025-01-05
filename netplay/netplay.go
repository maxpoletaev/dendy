package netplay

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/maxpoletaev/dendy/internal/binario"
	"github.com/maxpoletaev/dendy/internal/bytepool"
	"github.com/maxpoletaev/dendy/internal/ringbuf"
)

const (
	maxFrameSyncFreq    = 300
	pingIntervalFrames  = 120
	maxFrameDriftWindow = 20
	minFrameDriftWindow = 3    // should not be <3 as int(2*1.35)=2
	driftWindowFactor   = 1.35 // factor to increase/decrease the drift window
	maxPoolItemSize     = 8
	maxMessageBatch     = 10
)

var (
	byteOrder = binary.LittleEndian
)

type Netplay struct {
	game       *Game
	toRecv     chan Message
	toSend     chan Message
	conn       net.Conn
	rtt        time.Duration
	pool       *bytepool.BytePool
	rttWindow  *ringbuf.Buffer[time.Duration]
	readerDone chan struct{}
	writerDone chan struct{}
	shouldExit bool
	isHost     bool

	driftWindow   int
	syncFrame     uint32
	noDriftFrames uint32
	pingCounter   uint32
}

func newNetplay(game *Game, conn net.Conn) *Netplay {
	pool := bytepool.New(maxPoolItemSize)

	return &Netplay{
		rttWindow:   ringbuf.New[time.Duration](10),
		toSend:      make(chan Message, 100),
		toRecv:      make(chan Message, 100),
		driftWindow: minFrameDriftWindow,
		readerDone:  make(chan struct{}),
		writerDone:  make(chan struct{}),
		pool:        pool,
		game:        game,
		conn:        conn,
	}
}

func (np *Netplay) startWriter() {
	w := binario.NewWriter(np.conn, byteOrder)

	defer close(np.writerDone)

	for {
		msg, ok := <-np.toSend
		if !ok {
			break
		}

		if err := writeMsg(w, &msg); err != nil {
			log.Printf("[ERROR] failed to write message: %v", err)
			np.shouldExit = true
			break
		}
	}
}

func (np *Netplay) startReader() {
	r := binario.NewReader(np.conn, byteOrder)

	defer close(np.readerDone)

	for {
		msg := Message{}

		if err := readMsg(r, &msg, np.pool); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				break
			}

			log.Printf("[ERROR] failed to read message: %v", err)
			np.shouldExit = true
			break
		}

		select {
		case np.toRecv <- msg:
		default:
			log.Printf("[WARN] recv buffer is full, blocking")
		}
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
	for i := 0; i < maxMessageBatch; i++ {
		select {
		case msg, ok := <-np.toRecv:
			if ok {
				np.handleMessage(msg)
				msg.Buffer.Free()
			} else {
				break loop
			}
		default:
			break loop
		}
	}
}

// handleFrameDrift makes sure both emulators are running approximately at the same speed,
// by asking the remote side to wait if it detects a difference in the frame count.
func (np *Netplay) handleFrameDrift() {
	localFrame := np.game.Frame()
	driftFrames := np.game.DriftFrames()

	if driftFrames < 0 {
		driftFrames = -driftFrames

		if driftFrames > np.driftWindow && np.syncFrame+maxFrameSyncFreq < localFrame {
			log.Printf("[INFO] asking the remote to wait for %d frames", driftFrames)

			// We drifted, reset the counter and set the next sync frame.
			np.syncFrame = localFrame + uint32(rand.Int31n(maxFrameSyncFreq/10))
			np.noDriftFrames = 0

			// Ask the remote to wait if we are too far behind.
			np.SendWait(uint32(driftFrames))

			// Gradually increase the window to avoid oscillations.
			if np.driftWindow < maxFrameDriftWindow {
				np.driftWindow = min(maxFrameDriftWindow, int(float32(np.driftWindow)*driftWindowFactor))
				log.Printf("[DEBUG] drift window increased to %d", np.driftWindow)
			}
		}
	}

	// Start shrinking the window if everything is fine.
	if np.driftWindow > minFrameDriftWindow && np.noDriftFrames > maxFrameSyncFreq*10 {
		np.driftWindow = max(minFrameDriftWindow, int(float32(np.driftWindow)/driftWindowFactor))
		log.Printf("[DEBUG] drift window decreased to %d", np.driftWindow)
		np.noDriftFrames = 0
	}

	np.noDriftFrames++
}

// RunFrame progresses the game by one frame.
func (np *Netplay) RunFrame(startTime time.Time) {
	np.game.RunFrame(startTime)

	// Inject a ping message every N frames to measure latency.
	if np.game.Frame()%pingIntervalFrames == 0 {
		np.SendPing()
	}

	np.handleFrameDrift()
}

// RemotePing returns the ping time to the remote peer in milliseconds.
func (np *Netplay) RemotePing() int64 {
	return np.rtt.Milliseconds()
}
