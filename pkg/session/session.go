package session

// PlayerData contains the player's session information
type PlayerData struct {
	PlayerName
	PlayerUpdate
	CompleteCollection Collection `firestore:"complete_collection" json:"-"`
	SessionCollection  Collection `firestore:"session_collection" json:"-"`
}

// PlayerName is a placeholder for a player's name
type PlayerName struct {
	Name string `firestore:"name" json:"name"`
}

// PlayerRegistration contains the data needed for adding a new player
type PlayerRegistration struct {
	PlayerName
	Collection `firestore:"collection" json:"collection"`
}

// PlayerUpdate contains data provided when sending an update
type PlayerUpdate struct {
	Ready bool `firestore:"ready" json:"ready"`
}

// Options describes which kind of game is played
type Options struct {
	Singleton    bool `firestore:"singleton" json:"singleton"`
	Pauper       bool `firestore:"pauper" json:"pauper"`
	ColorOptions `json:"color"`
	Set          string `firestore:"set" json:"set"`
}

// ColorOptions contains all settings related to colors
type ColorOptions struct {
	White     bool `json:"white"`
	Blue      bool `json:"blue"`
	Black     bool `json:"black"`
	Red       bool `json:"red"`
	Green     bool `json:"green"`
	Colorless bool `json:"colorless"`
}

// Lookup does a lookup with the given color string
func (c ColorOptions) Lookup(color string) bool {
	switch color {
	case "W":
		return c.White
	case "U":
		return c.Blue
	case "B":
		return c.Black
	case "R":
		return c.Red
	case "G":
		return c.Green
	}

	return false
}

func allColors() ColorOptions { return ColorOptions{true, true, true, true, true, true} }

// Session describes a playsession
type Session struct {
	Players map[string]*PlayerData `firestore:"players" json:"players"`
	Started bool                   `firestore:"started" json:"started"`
	Options
}

// UpdatePlayer updates the given player based on the PlayerUpdate
func (s *Session) UpdatePlayer(cardDB CardDB, playerID string, update PlayerUpdate) {
	if s.Started {
		return
	}

	player, ok := s.Players[playerID]
	if !ok {
		return
	}

	player.PlayerUpdate = update

	// Check if the update made the session "startable"; start if yes
	s.startCheck(cardDB)
}

// RemovePlayer removes a player from the session
func (s *Session) RemovePlayer(playerID string) {
	if _, ok := s.Players[playerID]; !ok {
		return
	}

	delete(s.Players, playerID)
	// un-ready all players
	for playerID := range s.Players {
		s.Players[playerID].Ready = false
	}
}

func (s *Session) startCheck(cardDB CardDB) {
	if len(s.Players) < 2 {
		return
	}

	// check if all players are ready
	for _, player := range s.Players {
		if !player.Ready {
			return
		}
	}

	// start session
	// for now, only do "constructed. will change!
	s.constructed(cardDB)

	s.Started = true
}

func (s *Session) constructed(cardDB CardDB) {
	// create intersection
	var collection Collection
	for _, player := range s.Players {
		if collection == nil {
			collection = player.CompleteCollection.Copy()
			continue
		}

		collection.Intersect(player.CompleteCollection)
	}

	// filter colors
	if s.Options.ColorOptions != allColors() {
		collection.FilterColors(cardDB, s.Options.ColorOptions)
	}

	// set filter
	if s.Options.Set != "" {
		collection.FilterSet(cardDB, s.Options.Set)
	}

	// generally never show more than 4 per name
	max := byte(4)
	if s.Options.Singleton {
		max = 1
	}
	collection.MaxPerCard(cardDB, max)

	if s.Options.Pauper {
		collection.FilterRarities(cardDB, "common")
	}

	//  write back for each player
	for _, player := range s.Players {
		player.SessionCollection = collection
	}
}
