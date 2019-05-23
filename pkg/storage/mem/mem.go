package mem

import (
	"sync"

	"github.com/kjeisy/arenawithfriends/pkg/session"
	"github.com/lithammer/shortuuid"
)

// Store contains an in-memory implementation of controller.Storage
type Store struct {
	mutex    sync.RWMutex
	sessions map[string]*session.Session
}

// New initializes a new memory store
func New() *Store {
	return &Store{
		mutex:    sync.RWMutex{},
		sessions: map[string]*session.Session{},
	}
}

// CreateSession creates a new session
func (st *Store) CreateSession(opts session.Options) (string, error) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	var id string
	for {
		id = shortuuid.New()
		if _, ok := st.sessions[id]; !ok {
			break
		}
	}

	st.sessions[id] = &session.Session{
		Options: opts,
		Players: map[string]*session.PlayerData{},
	}

	return id, nil
}

// GetSession checks the data for the given session (nil = not found)
func (st *Store) GetSession(id string) (*session.Session, error) {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	return st.sessions[id], nil
}

// AddPlayer adds the given PlayerData to the session (nil output == session not found)
func (st *Store) AddPlayer(sessionID string, playerRegistration session.PlayerRegistration) (string, *session.Session, error) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	s := st.sessions[sessionID]
	if s == nil {
		return "", nil, nil
	}

	if s.Started {
		return "", s, nil
	}

	var playerID string
	for {
		playerID = shortuuid.New()
		if _, ok := s.Players[playerID]; !ok {
			break
		}
	}

	s.Players[playerID] = &session.PlayerData{
		PlayerName:         playerRegistration.PlayerName,
		CompleteCollection: playerRegistration.Collection,
	}

	return playerID, s, nil
}

// UpdatePlayer sets
func (st *Store) UpdatePlayer(cardDB session.CardDB, sessionID string, playerID string, update session.PlayerUpdate) (*session.Session, error) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	session := st.sessions[sessionID]
	if session == nil {
		return nil, nil
	}

	session.UpdatePlayer(cardDB, playerID, update)

	return session, nil
}

// RemovePlayer removes a player
func (st *Store) RemovePlayer(sessionID string, playerID string) *session.Session {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	session, ok := st.sessions[sessionID]
	if !ok {
		return session
	}

	session.RemovePlayer(playerID)

	if len(session.Players) == 0 {
		delete(st.sessions, sessionID)
		return nil
	}

	return session
}
