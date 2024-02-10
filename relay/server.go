package relay

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/xtaci/kcp-go"
)

type Session struct {
	ID         string
	Secret     string
	GameTitle  string
	HostAddr   net.UDPAddr
	ClientAddr net.UDPAddr
	Created    time.Time
	RomCRC32   uint32
	Region     string
	Public     bool
}

func NewSession() Session {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	var id [3]int
	for i := range id {
		id[i] = rnd.Intn(999)
	}

	var secret [16]byte
	for i := range secret {
		secret[i] = byte('!' + rnd.Intn('~'))
	}

	return Session{
		ID:      fmt.Sprintf("%03d-%03d-%03d", id[0], id[1], id[2]),
		Secret:  string(secret[:]),
		Created: time.Now(),
	}
}

type Server struct {
	limiter *IPLimiter
	store   Store
}

func NewServer(store Store, limiter *IPLimiter) *Server {
	return &Server{
		limiter: limiter,
		store:   store,
	}
}

func (s *Server) Listen(addr string) error {
	listener, err := kcp.Listen(addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept: %w", err)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	log.Printf("[INFO] new connection from %s", conn.RemoteAddr())

	msg, err := receive(conn)
	if err != nil {
		log.Printf("[ERROR] failed to read message: %v", err)
		return
	}

	switch m := msg.(type) {
	case *CreateSessionMsg:
		s.handleCreateSession(conn, m)
	case *JoinSessionMsg:
		s.handleJoinSession(conn, m)
	default:
		log.Printf("[ERROR] unknown message type: %t", m)
	}
}

func (s *Server) handleCreateSession(conn net.Conn, msg *CreateSessionMsg) {
	addr := conn.RemoteAddr().(*net.UDPAddr)

	session := NewSession()
	session.HostAddr = *addr
	session.RomCRC32 = msg.RomCRC32

	if ok := s.limiter.Acquire(addr.IP.String()); !ok {
		log.Printf("[WARN] rate limited: %s", addr.IP.String())
		sendError(conn, fmt.Errorf("rate limited"))
		return
	}

	if err := s.store.CreateSession(session); err != nil {
		log.Printf("[ERROR] failed to save session: %s", err)
		sendError(conn, fmt.Errorf("internal server error"))

		return
	}

	defer func() {
		if err := s.store.DeleteSession(session.ID); err != nil {
			log.Printf("[WARN] failed to delete session: %s", err)
		}

		s.limiter.Release(addr.IP.String())
	}()

	err := send(conn, &SessionCreatedMsg{ID: session.ID})
	if err != nil {
		log.Printf("[WARN] failed to write message: %v", err)
		return
	}

	log.Printf("[INFO] session %s created by %s", session.ID, session.HostAddr.String())

	expiryTimer := time.NewTimer(300 * time.Second)
	defer expiryTimer.Stop()

	select {
	case <-expiryTimer.C:
		log.Printf("[WARN] session %s was not joined in time", session.ID)
		sendError(conn, fmt.Errorf("session expired"))

		if err := s.store.DeleteSession(session.ID); err != nil {
			log.Printf("[WARN] failed to delete session: %s", err)
			return
		}

	case <-s.store.JoinedChan(session.ID):
		session, err = s.store.GetSession(session.ID)
		if err != nil {
			sendError(conn, fmt.Errorf("failed to get session: %s", err))
			log.Printf("[WARN] failed to get session: %s", err)

			return
		}

		err = send(conn, &StartGameMsg{
			IP:   session.ClientAddr.IP.To4(),
			Port: uint16(session.ClientAddr.Port),
		})

		if err != nil {
			log.Printf("[WARN] failed to write message: %v", err)
			return
		}
	}
}

func (s *Server) handleJoinSession(conn net.Conn, msg *JoinSessionMsg) {
	session, err := s.store.GetSession(msg.ID)
	if err != nil {
		sendError(conn, fmt.Errorf("could not find session: %s", err))
		return
	}

	if session.RomCRC32 != msg.RomCRC32 {
		sendError(conn, fmt.Errorf("rom mismatch"))
		return
	}

	session.ClientAddr = *conn.RemoteAddr().(*net.UDPAddr)

	if err := s.store.UpdateSession(session); err != nil {
		log.Printf("[ERROR] failed to save session: %s", err)
		sendError(conn, fmt.Errorf("internal server error"))

		return
	}

	err = send(conn, &StartGameMsg{
		IP:   session.HostAddr.IP.To4(),
		Port: uint16(session.HostAddr.Port),
	})

	if err != nil {
		log.Printf("[WARN] failed to write message: %v", err)
		return
	}

	s.store.NotifyJoined(session.ID) // wake up the host

	log.Printf("[INFO] session %s joined by %s", session.ID, session.ClientAddr.String())
}
