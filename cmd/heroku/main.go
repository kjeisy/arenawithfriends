package main

import (
	"net/http"
	"os"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/kjeisy/arenawithfriends/pkg/controller"
	"github.com/kjeisy/arenawithfriends/pkg/storage/mem"
)

func main() {
	// initialize with a pure in-memory storage (mem)
	model := controller.New(mem.New())

	http.Handle("/", model.Router())

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, nil)
}
