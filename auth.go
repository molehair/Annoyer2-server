package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
)

var BUCKET_DEVICES = []byte("devices")

type refreshTokenPayload struct {
	// required
	DeviceId string `json:"deviceId"`

	// required
	Token string `json:"token"`
}

type fcmToken struct {
	Token     string
	Timestamp time.Time
}

func InitializeAuth() error {
	// create buckets
	return db.Update(func(t *bolt.Tx) error {
		_, err := t.CreateBucketIfNotExists(BUCKET_DEVICES)
		return err
	})
}

// Receive token for push message and deviceId
// If token is valid, save it to DB.
func HandleRefreshToken(c *gin.Context) {
	// read data
	var payload refreshTokenPayload
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusNotAcceptable)
		return
	}

	// test if token is valid
	_, err := fcmClient.SendDryRun(context.TODO(), &messaging.Message{Token: payload.Token})
	if err == nil {
		//-- token is valid --//
		// timestamp
		now := time.Now()

		// Add to database
		key := payload.DeviceId + DB_DELIM + "fcmToken"
		value, _ := json.Marshal(fcmToken{payload.Token, now})
		err = db.Update(func(t *bolt.Tx) error {
			devices := t.Bucket(BUCKET_DEVICES)
			devices.Put([]byte(key), value)
			return nil
		})
		if err != nil {
			log.Printf("/refreshToken: %v", err)
		}
	} else {
		//-- token is not valid --//
		c.Status(http.StatusNotAcceptable)
	}
}

// Get the token for push notification from deviceId
func GetToken(deviceId string) string {
	var token string

	db.View(func(t *bolt.Tx) error {
		// open bucket
		devices := t.Bucket(BUCKET_DEVICES)
		if devices == nil {
			return errors.New("empty bucket: devices")
		}

		// get token
		var v fcmToken
		json.Unmarshal(devices.Get([]byte(deviceId+DB_DELIM+"fcmToken")), &v)
		token = v.Token

		return nil
	})

	return token
}
