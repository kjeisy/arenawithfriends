package lobby

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/kjeisy/arenawithfriends/pkg/session"
)

type Lobby struct {
	mutex    sync.Mutex
	sessions map[string]map[string]*websocket.Conn
}

func New() *Lobby {
	return &Lobby{
		mutex:    sync.Mutex{},
		sessions: map[string]map[string]*websocket.Conn{},
	}
}

func (l *Lobby) NewSession(sessionID string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.sessions[sessionID] = map[string]*websocket.Conn{}
}

func (l *Lobby) RegisterConnection(sessionID string, playerID string, connection *websocket.Conn) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, ok := l.sessions[sessionID]; !ok {
		return ErrSessionNotFound
	}

	if _, ok := l.sessions[sessionID][playerID]; ok {
		return ErrPlayerAlreadyRegistered
	}

	l.sessions[sessionID][playerID] = connection
	return nil
}

func (l *Lobby) Broadcast(sessionID string, sessionData *session.Session) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	sessions, ok := l.sessions[sessionID]
	if !ok {
		return
	}

	for playerID, broadcastConn := range sessions {
		if err := broadcastConn.WriteJSON(*sessionData); err != nil {
			broadcastConn.Close()
			delete(l.sessions[sessionID], playerID)
		}
	}
}

func (l *Lobby) Unregister(sessionID string, playerID string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, ok := l.sessions[sessionID]; !ok {
		return
	}
	delete(l.sessions[sessionID], playerID)

	// if this was the last player, remove the whole session
	if len(l.sessions[sessionID]) == 0 {
		delete(l.sessions, sessionID)
	}
}
