package controller

// Session describes a playsession
type Session struct {
	Players    []string `firestore:"players" json:"players"`
	Started    bool     `firestore:"started" json:"started"`
	Collection `firestore:"collection" json:"-"`
}

// PlayerData contains the information needed to add a new player
type PlayerData struct {
	Name       string `firestore:"name" json:"name"`
	Collection `firestore:"collection" json:"collection"`
}

// AddPlayer adds PlayerData to the session
func (s *Session) AddPlayer(playerData PlayerData) {
	// add player
	s.Players = append(s.Players, playerData.Name)

	// new collection
	if s.Collection == nil || len(s.Collection) == 0 {
		s.Collection = playerData.Collection
		return
	}

	// existing collection
	s.Collection.Intersect(playerData.Collection)

}
