package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"phaint/models"
	"phaint/services"
	"phaint/utils"
)

type UserHandler struct{}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	app := services.FirebaseDb().GetApp()
	user, err := models.NewFirebaseAuthUser(r)
	if err != nil {
		fmt.Println(err)
	}
	authx, err := app.Auth(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	userRecord, err := authx.CreateUser(context.Background(), user)
	if err != nil {
		log.Fatalf("Errore durante la creazione dell'utente: %v\n", err)
	}
	fmt.Printf("Utente creato: %v\n", userRecord.UID)
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	user, err := models.NewUserFromRequest(r)
	if err != nil {
		fmt.Println(err)
	}

	token, err := utils.SignInWithPassword(user.Mail, user.Password)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Utente loggato: %v\n", token)
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		h.RegisterUser(w, r)
		return
	case r.Method == http.MethodGet:
		h.LoginUser(w, r)
		return
	}
}
