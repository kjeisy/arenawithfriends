package mem

import (
	"net/http"
	"sync"

	"github.com/kjeisy/arenawithfriends/pkg/controller"
	"github.com/lithammer/shortuuid"
)

// Store contains an in-memory implementation of controller.Storage
type Store struct {
	mutex    sync.RWMutex
	sessions map[string]*controller.Session
}

// New initializes a new memory store
func New() *Store {
	return &Store{
		mutex:    sync.RWMutex{},
		sessions: map[string]*controller.Session{},
	}
}

// CreateSession creates a new session
func (s *Store) CreateSession(req *http.Request) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var id string
	for {
		id = shortuuid.New()
		if _, ok := s.sessions[id]; !ok {
			break
		}
	}

	s.sessions[id] = &controller.Session{}

	return id, nil
}

// GetSession checks the data for the given session (nil = not found)
func (s *Store) GetSession(req *http.Request, id string) (*controller.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.sessions[id], nil
}

// AddPlayer adds the given PlayerData to the session (nil output == session not found)
func (s *Store) AddPlayer(req *http.Request, id string, playerData controller.PlayerData) (*controller.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session := s.sessions[id]
	if session == nil {
		return nil, nil
	}

	if !session.Started {
		session.AddPlayer(playerData)
	}

	return session, nil
}

// StartSession starts the given session (session nil == not found)
func (s *Store) StartSession(req *http.Request, id string) (*controller.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session := s.sessions[id]

	if session == nil {
		return nil, nil
	}

	if !session.Started && len(session.Players) > 0 {
		session.Started = true
	}

	return session, nil
}
