package main

import (
	"net/http"

	"github.com/kjeisy/arenawithfriends/pkg/controller"
	"github.com/kjeisy/arenawithfriends/pkg/storage/firestore"
	"google.golang.org/appengine"
)

func main() {
	model := controller.New(&firestore.Store{})

	http.Handle("/", model.Router())
	appengine.Main()
}
