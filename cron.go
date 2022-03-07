package main

import (
	"time"

	"github.com/robfig/cron/v3"
)

// Run in UTC time
var cr = cron.New(cron.WithLocation(time.UTC))

func InitializeCron() error {
	cr.Start()
	return nil
}
