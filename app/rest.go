package app

import (
	"go-json-rest"
	"log"
	"net/http"
)

type User struct {
	Id   string
	Name string
}

func getUser(w *rest.ResponseWriter, req *rest.Request) {
	user := User{
		Id:   req.PathParam("id"),
		Name: "Antoine",
	}
	w.WriteJson(&user)
}

func getFarm(w *rest.ResponseWriter, req *rest.Request) {
	farm := Farm()
	w.WriteJson(farm)
}

func StartRestServer(listen string) {
	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.Route{"GET", "/users/:id", getUser},
		rest.Route{"GET", "/farm", getFarm},
	)
	http.ListenAndServe(listen, &handler)
	log.Println("Admin rest server started: ", listen)
}
