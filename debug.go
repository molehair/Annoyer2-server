package main

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
)

func addDebugFunctions(r *gin.Engine) {
	r.GET("/test1", func(c *gin.Context) {
		deviceId := c.Query("deviceId")
		dailyIndex, _ := strconv.Atoi(c.Query("dailyIndex"))
		if err := sendPracticeNotification(deviceId, 0, dailyIndex); err != nil {
			fmt.Printf("%v", err)
		} else {
			println("sent prac noti")
		}
	})
	r.GET("/test2", func(c *gin.Context) {
		deviceId := c.Query("deviceId")
		if err := sendTestNotification(deviceId, 0); err != nil {
			fmt.Printf("%v", err)
		} else {
			println("sent test noti")
		}
	})
	r.GET("/debugTrainingSchedules", func(c *gin.Context) {
		db.View(func(t *bolt.Tx) error {
			trainingSchedules := t.Bucket(BUCKET_TRAINING_SCHEDULES)
			trainingSchedules.ForEach(func(k, _ []byte) error {
				fmt.Printf("%s\n", k)
				return nil
			})
			return nil
		})
	})

	r.GET("/debugDevices", func(c *gin.Context) {
		db.View(func(t *bolt.Tx) error {
			devices := t.Bucket(BUCKET_DEVICES)
			devices.ForEach(func(k, v []byte) error {
				fmt.Printf("%s %s\n", k, v)
				return nil
			})
			return nil
		})
	})
}
