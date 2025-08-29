package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"phaint/internal/services"
	"phaint/internal/utils"
	"phaint/models"
)

type UserHandler struct{}

// RegisterUser Create a new user on Firebase Authentication system from an user passed in the http request body
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
	_, err = authx.CreateUser(context.Background(), user)
	if err != nil {
		log.Fatalf("Error during user craetion : %v\n", err)
	}
	h.LoginUser(w, r)
}

// LoginUser Authenticate a user passed in the http request body
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	user, err := models.NewUserFromRequest(r)
	if err != nil {
		fmt.Println(err)
	}

	token, err := utils.SignInWithPassword(user.Mail, user.Password)
	if err != nil {
		fmt.Println(err)
	}
	response := map[string]string{"UserToken": token}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/users/register":
		h.RegisterUser(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == "/users/login":
		h.LoginUser(w, r)
		return
	}
}
