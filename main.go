package main

import (
	"log"
	"net/http"
	auth "phaint/internal/handlers"
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
	mux.Handle("/users/register", &auth.UserHandler{})

	// Run the server
	http.ListenAndServe(":8080", mux)
}
