package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Storage interface for creating, retrieving and modifying Sessions
type Storage interface {
	CreateSession(*http.Request, Options) (string, error)
	GetSession(*http.Request, string) (*Session, error)
	AddPlayer(*http.Request, string, PlayerRegistration) (string, *Session, error)
	UpdatePlayer(*http.Request, CardDB, string, string, PlayerUpdate) (*Session, error)
}

// Controller describes the behavior of the app
type Controller struct {
	storage Storage
	cardDB  CardDB
}

// New initializes a fresh Controller with the given storage backend
func New(storage Storage, path string) (*Controller, error) {
	cardDB, err := LoadCardDB(path)
	if err != nil {
		return nil, err
	}

	return &Controller{
		storage: storage,
		cardDB:  cardDB,
	}, nil
}

// Router sets up the API call stack
func (m *Controller) Router() http.Handler {
	// avoid errors
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	root := router.Group("")
	root.StaticFile("/", "./public/index.html")
	root.Static("/css", "./public/css")
	root.Static("/data", "./public/data")
	root.Static("/img", "./public/img")
	root.Static("/js", "./public/js")

	session := root.Group("/api/v1/sessions")

	session.POST("", m.createSession)
	session.POST("/:sessionID/players", m.addPlayer)
	session.POST("/:sessionID/players/:playerID", m.updatePlayer)

	session.GET("/:sessionID", m.getSession)
	session.GET("/:sessionID/players/:playerID/collection", m.getSessionCollection)

	return router
}

func (m *Controller) createSession(c *gin.Context) {
	var opts Options
	if err := c.BindJSON(&opts); err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session creation option data provided"})
		return
	}

	key, err := m.storage.CreateSession(c.Request, opts)
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
	sessionID := getSessionID(c.Params)
	if sessionID == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session ID provided"})
		return
	}

	playerID := getPlayerID(c.Params)
	if playerID == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no player ID provided"})
		return
	}

	session, err := m.storage.GetSession(c.Request, sessionID)
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

	player, ok := session.Players[playerID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
	}

	c.JSON(http.StatusOK, player.SessionCollection)
}

func (m *Controller) addPlayer(c *gin.Context) {
	id := getSessionID(c.Params)
	if id == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session provided"})
		return
	}

	var playerRegistration PlayerRegistration
	if err := c.BindJSON(&playerRegistration); err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no player registration data provided"})
		return
	}

	if playerRegistration.Name == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no player name provided"})
		return
	}

	if len(playerRegistration.Collection) == 0 {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "empty collection provided"})
		return
	}

	playerID, session, err := m.storage.AddPlayer(c.Request, id, playerRegistration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error adding player"})
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
	if playerID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error adding player"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": playerID})
}

func (m *Controller) updatePlayer(c *gin.Context) {
	sessionID := getSessionID(c.Params)
	if sessionID == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session provided"})
		return
	}

	playerID := getPlayerID(c.Params)
	if playerID == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no player provided"})
		return
	}

	var playerUpdate PlayerUpdate
	if err := c.BindJSON(&playerUpdate); err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no player update data provided"})
		return
	}

	session, err := m.storage.UpdatePlayer(c.Request, m.cardDB, sessionID, playerID, playerUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update player"})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if _, ok := session.Players[playerID]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	c.JSON(http.StatusOK, *session)
}

func getSessionID(params gin.Params) string {
	return params.ByName("sessionID")
}
func getPlayerID(params gin.Params) string {
	return params.ByName("playerID")
}
