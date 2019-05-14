package controller

// Collection represents a card collection (id -> number of cards)
type Collection map[string]byte

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
