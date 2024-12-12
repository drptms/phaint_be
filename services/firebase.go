package services

import (
	"context"
	"fmt"
	"phaint/config"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type FireDB struct {
	app    *firebase.App
	client *firestore.Client
}

var fireDb FireDB

func (db *FireDB) Connect() error {
	context := context.Background()
	options := option.WithCredentialsFile(config.FirebaseCredentialsPath())
	config := &firebase.Config{
		ProjectID:   "phaint-ae2f2",
		DatabaseURL: config.FirebaseDBUrl(),
	}

	app, err := firebase.NewApp(context, config, options)
	if err != nil {
		fmt.Println(err)
		return err
	}
	db.app = app

	client, err := app.Firestore(context)
	if err != nil {
		fmt.Println(err)
		return err
	}

	db.client = client
	return nil
}

func (db *FireDB) GetClient() *firestore.Client {
	return db.client
}

func (db *FireDB) GetApp() *firebase.App {
	return db.app
}

func FirebaseDb() *FireDB {
	return &fireDb
}
