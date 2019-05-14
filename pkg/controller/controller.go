package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const sessionID = "sessionID"

// Storage interface for creating, retrieving and modifying Sessions
type Storage interface {
	CreateSession(*http.Request) (string, error)
	GetSession(*http.Request, string) (*Session, error)
	AddPlayer(*http.Request, string, PlayerData) (*Session, error)
	StartSession(*http.Request, string) (*Session, error)
}

// Controller describes the behavior of the app
type Controller struct {
	storage Storage
}

// New initializes a fresh Controller with the given storage backend
func New(storage Storage) *Controller {
	return &Controller{
		storage: storage,
	}
}

// Router sets up the API call stack
func (m *Controller) Router() http.Handler {
	// avoid errors
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	session := router.Group("/api/v1/sessions")

	session.POST("", m.createSession)
	session.POST("/:sessionID/players", m.addPlayer)
	session.POST("/:sessionID/start", m.startSession)

	session.GET("/:sessionID", m.getSession)
	session.GET("/:sessionID/collection", m.getSessionCollection)

	return router
}

func (m *Controller) createSession(c *gin.Context) {
	key, err := m.storage.CreateSession(c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": key})
}

func (m *Controller) getSession(c *gin.Context) {
	id := getSessionID(c.Params)
	if id == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session provided"})
		return
	}

	session, err := m.storage.GetSession(c.Request, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching session"})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}
func (m *Controller) getSessionCollection(c *gin.Context) {
	id := getSessionID(c.Params)
	if id == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session provided"})
		return
	}

	session, err := m.storage.GetSession(c.Request, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching session"})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	if !session.Started {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "session not started"})
		return
	}

	c.JSON(http.StatusOK, session.Collection)
}
func (m *Controller) addPlayer(c *gin.Context) {
	id := getSessionID(c.Params)
	if id == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session provided"})
		return
	}

	var playerData PlayerData
	if err := c.BindJSON(&playerData); err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no player data provided"})
		return
	}

	if playerData.Name == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no player name provided"})
		return
	}

	if len(playerData.Collection) == 0 {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "empty collection provided"})
		return
	}

	session, err := m.storage.AddPlayer(c.Request, id, playerData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating session"})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if session.Started {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "session already started"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (m *Controller) startSession(c *gin.Context) {
	id := getSessionID(c.Params)
	if id == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session provided"})
		return
	}

	session, err := m.storage.StartSession(c.Request, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error starting session"})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if !session.Started {
		if len(session.Players) == 0 {
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": "session not started: no players"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "session not started: unknown"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func getSessionID(params gin.Params) string {
	return params.ByName(sessionID)
}
