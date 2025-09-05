package main

import (
	"log"
	"net/http"
	handlers "phaint/internal/handlers"
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
	mux.Handle("/users/", &handlers.UserHandler{})
	mux.Handle("/projects/", &handlers.ProjectHandler{})
	// WorkBoard handler
	mux.Handle("/workboard", &handlers.WorkBoardHandler{})

	// Add WebSocket handler
	mux.Handle("/connect", &handlers.WebSocketHandler{})

	// Run the server
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Panic(err)
	}
}
