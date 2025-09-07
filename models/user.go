package models

import (
	"encoding/json"
	"net/http"

	"firebase.google.com/go/auth"
)

type User struct {
	Uid      string
	Username string
	Mail     string
	Password string
}

// NewUserFromRequest Create a new user from an http request
func NewUserFromRequest(r *http.Request) (User, error) {
	decoder := json.NewDecoder(r.Body)
	var user User
	err := decoder.Decode(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// NewFirebaseAuthUser Create a new auth.UserToCreate from an http request
func NewFirebaseAuthUser(r *http.Request) (*auth.UserToCreate, error) {
	user := auth.UserToCreate{}
	httpUser, err := NewUserFromRequest(r)
	if err != nil {
		return &user, err
	}
	user.DisplayName(httpUser.Username)
	user.Email(httpUser.Mail)
	user.Password(httpUser.Password)
	return &user, nil
}
