// +build heroku

package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/heroku/x/hmetrics/onload"
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

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, nil)
}
