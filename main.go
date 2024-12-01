package main

import (
	"log"
	"phaint/internal/services"
)

func main() {
	err := services.FirebaseDb().Connect()
	if err != nil {
		log.Println(err)
	}
	log.Println("After!")
}
