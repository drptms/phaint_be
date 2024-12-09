package main

import (
	"log"
	"net/http"
	user "phaint/internal/handlers"
	"phaint/internal/services"
	crypto "phaint/internal/utils"
)

func main() {
	err := services.FirebaseDb().Connect()
	if err != nil {
		log.Println(err)
	}
	err = crypto.GenerateRSAKeys()
	if err != nil {
		log.Println(err)
	}
	log.Println("Creating the server on port 8080")
	mux := http.NewServeMux()
	// adding all the handlers
	mux.Handle("/users/register", &user.UserHandler{})

	// Run the server
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Panic(err)
	}
}
