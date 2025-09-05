package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"phaint/internal/services"
	"phaint/internal/utils"
	"phaint/models"

	"google.golang.org/api/iterator"
)

type UserHandler struct{}

// RegisterUser Create a new user on Firebase Authentication system from an user passed in the http request body
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	app := services.FirebaseDb().GetApp()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	user, err := models.NewFirebaseAuthUser(r)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	userReq, err := models.NewUserFromRequest(r)

	ctx := context.Background()
	client := services.FirebaseDb().GetClient()

	_, _, err = client.Collection("users").Add(ctx, map[string]interface{}{
		"UID":      userReq.Uid,
		"mail":     userReq.Mail,
		"username": userReq.Username,
	})

	if err != nil {
		fmt.Println(err)
	}
	authx, err := app.Auth(ctx)
	if err != nil {
		fmt.Println(err)
	}
	_, err = authx.CreateUser(ctx, user)
	if err != nil {
		log.Fatalf("Error during user craetion : %v\n", err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
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

	ctx := context.Background()
	client := services.FirebaseDb().GetClient()
	iter := client.Collection("users").Documents(ctx)
	defer iter.Stop()

	var arr []map[string]interface{}
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Println("Err during collection iteration")
		}
		if doc.Data()["mail"] == user.Mail {
			arr = append(arr, doc.Data())
		}
	}

	response := map[string]string{"UserToken": token, "username": arr[0]["username"].(string)}

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
	default:
		http.NotFound(w, r)
	}
}
