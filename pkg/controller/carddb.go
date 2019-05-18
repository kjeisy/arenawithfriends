package controller

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// CardDB contains all card details in an Arena-Centric format
type CardDB map[ArenaID]CardData

// CardData is the details of a single card
type CardData struct {
	Name            string   `json:"name"`
	CMC             uint     `json:"cmc"`
	ColorIdentity   []string `json:"color_identity"`
	Set             string   `json:"set"`
	CollectorNumber string   `json:"collector_number"`
	Rarity          string   `json:"rarity"`
}

// LoadCardDB gets the current card database from a file
func LoadCardDB(path string) (CardDB, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var output CardDB
	if err := json.Unmarshal(byteValue, &output); err != nil {
		return nil, err
	}

	return output, nil
}
