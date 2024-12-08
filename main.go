package main

import (
	"log"
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
	log.Println("Now trying to add a user!")
	auth.RegisterUser("nick", "n@gmail.com", "nick")
}
