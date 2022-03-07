package main

import (
	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

func InitializeDatabase() error {
	// Open the database file in your current directory.
	// It will be created if it doesn't exist.
	var err error
	db, err = bolt.Open("my.db", 0600, nil)
	return err
}
