package relay

import (
	"fmt"
	"sync"
	"time"
)

var (
	_ Store = (*InMemoryStore)(nil)
)

type Store interface {
	DeleteSession(id string) error
	GetSession(id string) (Session, error)
	CreateSession(session Session) error
	UpdateSession(session Session) error
	ListPublicSessions() ([]Session, error)
	DeleteExpired() error
	JoinedChan(id string) chan struct{}
	NotifyJoined(id string)
}

type InMemoryStore struct {
	joined   map[string]chan struct{}
	sessions map[string]Session
	mut      sync.Mutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		sessions: make(map[string]Session),
		joined:   make(map[string]chan struct{}),
	}
}

func (s *InMemoryStore) GetSession(id string) (Session, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	if session, ok := s.sessions[id]; ok {
		return session, nil
	} else {
		return Session{}, fmt.Errorf("not found")
	}
}

func (s *InMemoryStore) CreateSession(session Session) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	if _, ok := s.sessions[session.ID]; ok {
		return fmt.Errorf("already exists")
	}

	s.sessions[session.ID] = session

	return nil
}

func (s *InMemoryStore) UpdateSession(session Session) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	if _, ok := s.sessions[session.ID]; !ok {
		return fmt.Errorf("not found")
	}

	s.sessions[session.ID] = session

	return nil
}

func (s *InMemoryStore) deleteSession(id string) {
	delete(s.sessions, id)

	if ch, ok := s.joined[id]; ok {
		delete(s.joined, id)
		close(ch)
	}
}

func (s *InMemoryStore) DeleteSession(id string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.deleteSession(id)

	return nil
}

func (s *InMemoryStore) DeleteExpired() error {
	s.mut.Lock()
	defer s.mut.Unlock()

	for id, session := range s.sessions {
		if time.Since(session.Created) > 5*time.Minute {
			s.deleteSession(id)
		}
	}

	return nil
}

func (s *InMemoryStore) JoinedChan(id string) chan struct{} {
	s.mut.Lock()
	defer s.mut.Unlock()

	if ch, ok := s.joined[id]; ok {
		return ch
	}

	ch := make(chan struct{})
	s.joined[id] = ch

	return ch
}

func (s *InMemoryStore) NotifyJoined(id string) {
	s.mut.Lock()
	defer s.mut.Unlock()

	if ch, ok := s.joined[id]; ok {
		delete(s.joined, id)
		close(ch)
	}
}

func (s *InMemoryStore) ListPublicSessions() (sessions []Session, _ error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	for _, session := range s.sessions {
		if session.Public {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}
