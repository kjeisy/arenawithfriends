// +build appengine

package firestore

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/kjeisy/arenawithfriends/pkg/controller"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Store implements model.Storage
type Store struct{}

func getSessionCollection(ctx context.Context) (*firestore.CollectionRef, error) {
	client, err := firestore.NewClient(ctx, appengine.AppID(ctx))
	if err != nil {
		return nil, err
	}

	return client.Collection("sessions"), nil
}

// CreateSession creates a new session
func (s *Store) CreateSession(req *http.Request) (string, error) {
	ctx := appengine.NewContext(req)
	sessionCollection, err := getSessionCollection(ctx)
	if err != nil {
		log.Errorf(ctx, "could not load session collection: %v", err)
		return "", err
	}

	session := controller.Session{
		Players: []string{},
		Started: false,
	}

	docRef, _, err := sessionCollection.Add(ctx, session)
	if err != nil {
		log.Errorf(ctx, "could not create session: %v", err)
		return "", err
	}

	return docRef.ID, nil
}

// GetSession checks the data for the given session (nil = not found)
func (s *Store) GetSession(req *http.Request, id string) (*controller.Session, error) {
	ctx := appengine.NewContext(req)

	// check if cached access is possible

	if session := getSessionMem(ctx, id); session != nil {
		return session, nil
	}

	// uncached access: firestore query
	session, _, err := getSession(ctx, id)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			// not found
			return nil, nil
		}

		// other errors
		log.Errorf(ctx, "error fetching session: %v", err)
		return nil, err
	}

	return session, nil
}

// AddPlayer adds the given PlayerData to the session (nil output == session not found)
func (s *Store) AddPlayer(req *http.Request, id string, playerData controller.PlayerData) (*controller.Session, error) {
	ctx := appengine.NewContext(req)

	// check if entry is in memory store -> entry can no longer be modified
	if session := getSessionMem(ctx, id); session != nil {
		return session, nil
	}

	session, doc, err := getSession(ctx, id)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			// not found
			return nil, nil
		}

		// other errors
		log.Errorf(ctx, "error fetching session: %v", err)
		return nil, err
	}

	if !session.Started {
		session.AddPlayer(playerData)

		// update session
		if _, err := doc.Set(ctx, session); err != nil {
			log.Errorf(ctx, "could not update session: %v", err)
			return nil, err
		}
	}

	return session, nil
}

// StartSession starts the given session (session nil == not found)
func (s *Store) StartSession(req *http.Request, id string) (*controller.Session, error) {
	ctx := appengine.NewContext(req)

	// check if entry is in memory store -> entry can no longer be modified
	if session := getSessionMem(ctx, id); session != nil {
		return session, nil
	}

	session, doc, err := getSession(ctx, id)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			// not found
			return nil, nil
		}

		// other errors
		log.Errorf(ctx, "error fetching session: %v", err)
		return nil, err
	}

	if len(session.Players) == 0 {
		return session, nil
	}

	if !session.Started {
		session.Started = true

		_, err = doc.Set(ctx, session)
		if err != nil {
			log.Errorf(ctx, "could not update session: %v", err)
			return nil, err
		}

	}

	// add to memory (if we're getting here the session is started and can be cached)
	addSessionMem(ctx, id, *session)
	return session, nil
}

func getSession(ctx context.Context, id string) (*controller.Session, *firestore.DocumentRef, error) {
	sessionCollection, err := getSessionCollection(ctx)
	if err != nil {
		log.Errorf(ctx, "could not get session collection: %v", err)
		return nil, nil, err
	}

	doc := sessionCollection.Doc(id)
	if doc == nil {
		log.Errorf(ctx, "invalid ID: %s", id)
		return nil, nil, fmt.Errorf("invalid ID: %s", id)
	}

	snap, err := doc.Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	var session controller.Session
	if err := snap.DataTo(&session); err != nil {
		return nil, nil, err
	}

	// cache session if it is started
	if session.Started {
		addSessionMem(ctx, id, session)
	}

	return &session, doc, nil
}

func getSessionMem(ctx context.Context, id string) *controller.Session {
	var session controller.Session

	_, err := memcache.Gob.Get(ctx, id, &session)
	if err != nil {
		if err != memcache.ErrCacheMiss {
			log.Errorf(ctx, "could not get cached session: %v", err)
		}
		return nil
	}

	if !session.Started {
		log.Errorf(ctx, "not started but in memstore: %s", id)
		return nil
	}

	return &session
}

func addSessionMem(ctx context.Context, id string, session controller.Session) {
	// only add if session is started
	if !session.Started {
		return
	}

	item := &memcache.Item{
		Key:    id,
		Object: session,
	}
	if err := memcache.Gob.Set(ctx, item); err != nil {
		log.Errorf(ctx, "could not add new entry to memcache: %v", err)
		return
	}
	log.Infof(ctx, "added key to memcache: %s", id)
}
