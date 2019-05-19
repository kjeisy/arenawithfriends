package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kjeisy/arenawithfriends/pkg/lobby"
	"github.com/kjeisy/arenawithfriends/pkg/session"
)

// Storage interface for creating, retrieving and modifying Sessions
type Storage interface {
	CreateSession(session.Options) (string, error)
	GetSession(string) (*session.Session, error)
	AddPlayer(string, session.PlayerRegistration) (string, *session.Session, error)
	RemovePlayer(string, string) *session.Session
	UpdatePlayer(session.CardDB, string, string, session.PlayerUpdate) (*session.Session, error)
}

// Controller describes the behavior of the app
type Controller struct {
	storage Storage
	cardDB  session.CardDB
	lobby   *lobby.Lobby
}

// New initializes a fresh Controller with the given storage backend
func New(storage Storage, path string) (*Controller, error) {
	cardDB, err := session.LoadCardDB(path)
	if err != nil {
		return nil, err
	}

	return &Controller{
		storage: storage,
		cardDB:  cardDB,
		lobby:   lobby.New(),
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
	//session.POST("/:sessionID/players/:playerID", m.updatePlayer)

	session.GET("/:sessionID", m.getSession)
	session.GET("/:sessionID/players/:playerID", m.getSessionWebSocket)
	session.GET("/:sessionID/players/:playerID/collection", m.getSessionCollection)

	return router
}

func (m *Controller) createSession(c *gin.Context) {
	var opts session.Options
	if err := c.BindJSON(&opts); err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session creation option data provided"})
		return
	}

	sessionid, err := m.storage.CreateSession(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}

	m.lobby.NewSession(sessionid)
	c.JSON(http.StatusOK, gin.H{"id": sessionid})
}

func (m *Controller) getSession(c *gin.Context) {
	id := getSessionID(c.Params)
	if id == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session provided"})
		return
	}

	session, err := m.storage.GetSession(id)
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

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (m *Controller) getSessionWebSocket(c *gin.Context) {
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

	s, err := m.storage.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching session"})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if _, ok := s.Players[playerID]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	if err := m.lobby.RegisterConnection(sessionID, playerID, conn); err != nil {
		conn.WriteJSON(gin.H{"error": "player already registered"})
		return
	}

	for {
		var update session.PlayerUpdate
		if err := conn.ReadJSON(&update); err != nil {
			break
		}

		session, err := m.storage.UpdatePlayer(m.cardDB, sessionID, playerID, update)
		if err != nil {
			conn.WriteJSON(gin.H{"error": "could not update player"})
			continue
		}
		if session == nil {
			conn.WriteJSON(gin.H{"error": "session not found"})
			continue
		}
		if _, ok := session.Players[playerID]; !ok {
			conn.WriteJSON(gin.H{"error": "player not found"})
			continue
		}

		// broadcast change
		m.lobby.Broadcast(sessionID, session)

	}

	m.lobby.Unregister(sessionID, playerID)
	s = m.storage.RemovePlayer(sessionID, playerID)
	if s != nil {
		m.lobby.Broadcast(sessionID, s)
	}
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

	session, err := m.storage.GetSession(sessionID)
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
	sessionID := getSessionID(c.Params)
	if sessionID == "" {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "no session provided"})
		return
	}

	var playerRegistration session.PlayerRegistration
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

	playerID, session, err := m.storage.AddPlayer(sessionID, playerRegistration)
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

	m.lobby.Broadcast(sessionID, session)

	c.JSON(http.StatusOK, gin.H{"id": playerID})
}

func getSessionID(params gin.Params) string {
	return params.ByName("sessionID")
}
func getPlayerID(params gin.Params) string {
	return params.ByName("playerID")
}
