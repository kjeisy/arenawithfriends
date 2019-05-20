package session

// Collection represents a card collection (id -> number of cards)
type Collection map[ArenaID]byte

// ArenaID is the unique identifier for a card, used in MTG Arena
type ArenaID string

// Intersect reduces collection to the intersection of c and i
func (c Collection) Intersect(i Collection) {
	for key, count := range c {
		countIn, ok := i[key]
		// if i doesn't contain the entry, remove it
		if !ok {
			delete(c, key)
			continue
		}

		// if i has less copies of that card, reduce it
		if count > countIn {
			c[key] = countIn
		}
	}
}

// MaxPerCard sets the card of the same name to at most the given number
func (c Collection) MaxPerCard(cardDB CardDB, max byte) {
	// create a map from card name to count (there are duplicates between sets, and we want to avoid that)
	nameMap := map[string]ArenaID{}

	for arenaID, count := range c {
		cardDetails, ok := cardDB[arenaID]
		if !ok {
			delete(c, arenaID)
			continue
		}

		// card name does not exist yet
		mainArenaID, ok := nameMap[cardDetails.Name]
		if !ok {
			nameMap[cardDetails.Name] = arenaID
			mainArenaID = arenaID
		} else {
			// card already exists, add count
			c[mainArenaID] += count
			delete(c, arenaID)
		}

		if c[mainArenaID] > max {
			c[mainArenaID] = max
		}
	}
}

// FilterRarities removes all cards that are not part of the given rarities
func (c Collection) FilterRarities(cardDB CardDB, rarities ...string) {
	keepRarity := map[string]struct{}{}
	for _, rarity := range rarities {
		keepRarity[rarity] = struct{}{}
	}

	for arenaID := range c {
		cardDetails, ok := cardDB[arenaID]
		if !ok {
			delete(c, arenaID)
			continue
		}

		if _, ok := keepRarity[cardDetails.Rarity]; !ok {
			delete(c, arenaID)
		}
	}
}

// FilterSet only returns the cards that are part of the given set
// TODO also allow the same cards from a different set
func (c Collection) FilterSet(cardDB CardDB, set string) {
	for arenaID := range c {
		cardDetails, ok := cardDB[arenaID]
		if !ok {
			delete(c, arenaID)
			continue
		}

		if cardDetails.Set != set {
			delete(c, arenaID)
		}
	}
}

// FilterColors removes non-configured colors from the collection
func (c Collection) FilterColors(cardDB CardDB, colors ColorOptions) {
	for arenaID := range c {
		cardDetails, ok := cardDB[arenaID]
		if !ok {
			delete(c, arenaID)
			continue
		}

		if len(cardDetails.ColorIdentity) == 0 {
			if !colors.Colorless {
				delete(c, arenaID)
			}
			continue
		}

		// all colors need to match for it to stay in the collection
		for _, color := range cardDetails.ColorIdentity {
			match := colors.Lookup(color)

			if !match {
				delete(c, arenaID)
				break
			}
		}
	}
}

// Copy creates a duplicate collection and returns it
func (c Collection) Copy() Collection {
	out := Collection{}
	for key, count := range c {
		out[key] = count
	}
	return out
}
