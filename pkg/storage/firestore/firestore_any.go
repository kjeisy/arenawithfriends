package firestore

import (
	"net/http"

	"github.com/kjeisy/arenawithfriends/pkg/controller"
)

// Store implements model.Storage
type Store struct{}

// CreateSession creates a new session
func (s *Store) CreateSession(req *http.Request) (string, error) {
	panic(ErrPlatformNotSupported)
}

// GetSession checks the data for the given session (nil = not found)
func (s *Store) GetSession(req *http.Request, id string) (*controller.Session, error) {
	panic(ErrPlatformNotSupported)
}

// AddPlayer adds the given PlayerData to the session (nil output == session not found)
func (s *Store) AddPlayer(req *http.Request, id string, playerData controller.PlayerData) (*controller.Session, error) {
	panic(ErrPlatformNotSupported)
}

// StartSession starts the given session (session nil == not found)
func (s *Store) StartSession(req *http.Request, id string) (*controller.Session, error) {
	panic(ErrPlatformNotSupported)
}
