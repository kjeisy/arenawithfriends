// +build !appengine
// +build !heroku

package main

import (
	"net/http"

	"github.com/kjeisy/arenawithfriends/pkg/controller"
	"github.com/kjeisy/arenawithfriends/pkg/storage/mem"
)

func main() {
	// initialize with a pure in-memory storage (mem)
	model := controller.New(mem.New())

	http.Handle("/", model.Router())
	http.ListenAndServe(":8081", nil)
}
