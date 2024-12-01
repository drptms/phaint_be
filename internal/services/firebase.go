package services

import (
	"context"
	"phaint/config"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"google.golang.org/api/option"
)

type FireDB struct {
	*db.Client
}

var fireDb FireDB

func (db *FireDB) Connect() error {
	context := context.Background()
	option := option.WithCredentialsFile(config.FirebaseCredentialsPath())
	config := &firebase.Config{DatabaseURL: config.FirebaseDBUrl()}
	app, err := firebase.NewApp(context, config, option)
	if err != nil {
		return err
	}
	client, err := app.Database(context)
	if err != nil {
		return err
	}
	db.Client = client
	return nil
}

func FirebaseDb() *FireDB {
	return &fireDb
}
