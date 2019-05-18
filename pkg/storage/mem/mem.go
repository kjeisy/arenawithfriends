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
func (s *Store) CreateSession(req *http.Request, opts controller.Options) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var id string
	for {
		id = shortuuid.New()
		if _, ok := s.sessions[id]; !ok {
			break
		}
	}

	s.sessions[id] = &controller.Session{
		Options: opts,
		Players: map[string]*controller.PlayerData{},
	}

	return id, nil
}

// GetSession checks the data for the given session (nil = not found)
func (s *Store) GetSession(req *http.Request, id string) (*controller.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.sessions[id], nil
}

// AddPlayer adds the given PlayerData to the session (nil output == session not found)
func (s *Store) AddPlayer(req *http.Request, sessionID string, playerRegistration controller.PlayerRegistration) (string, *controller.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session := s.sessions[sessionID]
	if session == nil {
		return "", nil, nil
	}

	if session.Started {
		return "", session, nil
	}

	var playerID string
	for {
		playerID = shortuuid.New()
		if _, ok := session.Players[playerID]; !ok {
			break
		}
	}

	session.Players[playerID] = &controller.PlayerData{
		PlayerName:         playerRegistration.PlayerName,
		CompleteCollection: playerRegistration.Collection,
	}

	return playerID, session, nil
}

// UpdatePlayer sets
func (s *Store) UpdatePlayer(req *http.Request, cardDB controller.CardDB, sessionID string, playerID string, update controller.PlayerUpdate) (*controller.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session := s.sessions[sessionID]
	if session == nil {
		return nil, nil
	}

	session.UpdatePlayer(cardDB, playerID, update)

	return session, nil
}
