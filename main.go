// +build !appengine
// +build !heroku

package main

import (
	"log"
	"net/http"

	"github.com/kjeisy/arenawithfriends/pkg/controller"
	"github.com/kjeisy/arenawithfriends/pkg/storage/mem"
)

func main() {
	// initialize with a pure in-memory storage (mem)
	model, err := controller.New(mem.New(), "public/data/MTGACards.json")
	if err != nil {
		log.Fatal(err)
	}

	router := model.Router()

	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}
