package main

import (
	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

func InitializeFirebase() error {
	// firebase
	opts := []option.ClientOption{option.WithCredentialsFile("service-account.json")}
	app, err := firebase.NewApp(context.Background(), nil, opts...)
	if err != nil {
		return err
	}

	// FCM
	fcmClient, err = app.Messaging(context.TODO())
	return err
}
