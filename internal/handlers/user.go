package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"phaint/internal/services"
	crypto "phaint/internal/utils"
	"phaint/models"
)

type UserHandler struct{}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user models.User
	err := decoder.Decode(&user)
	if err != nil {
		fmt.Println(err)
	}
	cryptedPassword, err := crypto.Encrypt(user.Password)
	if err != nil {
		log.Println(err)
	}
	client := services.FirebaseDb().GetClient()
	ref := client.Collection("users").NewDoc()
	_, err = ref.Set(context.Background(), map[string]interface{}{
		"username": user.Username,
		"mail":     user.Mail,
		"password": cryptedPassword,
	})
	user.Id = ref.ID
	if err != nil {
		w.WriteHeader(http.StatusConflict)
	}
	w.WriteHeader(200)
	_, err = w.Write([]byte(ref.ID))
	if err != nil {
		log.Println(err)
	}
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("In ServeHTTP")
	switch {
	case r.Method == http.MethodPost:
		h.RegisterUser(w, r)
		return
	case r.Method == http.MethodGet:
		log.Println("In the GET")
		return
	}
}
