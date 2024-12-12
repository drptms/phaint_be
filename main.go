package main

import (
	"log"
	"net/http"
	user "phaint/internal/handlers"
	"phaint/internal/services"
)

func main() {
	err := services.FirebaseDb().Connect()
	if err != nil {
		log.Println(err)
	}
	if err != nil {
		log.Println(err)
	}
	log.Println("Creating the server on port 8080")
	mux := http.NewServeMux()
	// adding all the handlers
	mux.Handle("/users/register", &user.UserHandler{})
	mux.Handle("/users/login", &user.UserHandler{})

	// Run the server
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Panic(err)
	}
}
