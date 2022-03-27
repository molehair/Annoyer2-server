package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
)

var BUCKET_TRAINING_SCHEDULES = []byte("trainingSchedules")

type setTrainingPayload struct {
	// required
	DeviceId string `json:"deviceId"`

	// optional
	// Schedule in UTC time
	// If it is empty, then training will be disabled.
	// Otherwise, it must be in the form below, and the training will be set.
	// format: "hhmmwwwwwww" where 00 <= hh < 24, 00 <= mm < 60, and
	//         each w is either 0 or 1, meaning disable/enable Sun, ... , Sat
	Schedule *string `json:"schedule"`
}

func InitializeTraining() error {
	// Run trainer
	// min hour day month weekday
	_, err := cr.AddFunc("*/15 * * * *", trainer)
	if err != nil {
		return err
	}

	// create buckets
	return db.Update(func(t *bolt.Tx) error {
		_, err := t.CreateBucketIfNotExists(BUCKET_TRAINING_SCHEDULES)
		return err
	})
}

// Called every 15 minutes, it sends training messages to the devices
func trainer() {
	// Round the current time to the closest multiple of 15-minute
	close15 := time.Now().UTC().Round(15 * time.Minute)
	curHour := close15.Hour()
	curMin := close15.Minute()
	curWeekday := close15.Weekday()

	db.View(func(t *bolt.Tx) error {
		// open buckets
		trainingSchedules := t.Bucket(BUCKET_TRAINING_SCHEDULES)
		c := trainingSchedules.Cursor()

		// Iterate over the past 8 hours
		if close15.Hour() >= 8 {
			//-- the past 8 hours are within a day --//
			// Compute the boundaries of the 8 hours
			hhmmBegin := fmt.Sprintf("%02d%02d", curHour-8, curMin)
			hhmmEnd := fmt.Sprintf("%02d%02d", curHour, curMin)
			endTimeInMin := curHour*60 + curMin

			// iterate
			for k, _ := c.Seek([]byte(hhmmBegin)); k != nil && string(k[:4]) <= hhmmEnd; k, _ = c.Next() {
				hhmm := string(k[:4])
				w7 := string(k[4:11])
				hour, _ := strconv.Atoi(hhmm[:2])
				minute, _ := strconv.Atoi(hhmm[2:])

				// check weekday
				if w7[curWeekday] == '0' {
					continue
				}

				// check if the time difference is a multiple of 30 minutes
				if (minute-curMin)%30 != 0 {
					continue
				}

				// Compute daily index
				timeInMin := hour*60 + minute
				dailyIndex := (endTimeInMin - timeInMin) / 30

				// send push
				deviceId := string(k[12:])
				trainingId := int(curWeekday)
				if dailyIndex < 16 {
					//-- practice --//
					fmt.Printf("deviceId=%s, trainingId=%d, dailyIndex=%d\n", deviceId, trainingId, dailyIndex)
					sendPracticeNotification(deviceId, trainingId, dailyIndex)
				} else {
					//-- test --//
					sendTestNotification(deviceId, trainingId)
				}
			}
		} else {
			//-- the past 8 hours are not within a day --//
			// Compute the boundaries of the 8 hours
			hhmmBegin := fmt.Sprintf("%02d%02d", curHour+16, curMin)
			hhmmEnd := fmt.Sprintf("%02d%02d", curHour, curMin)
			endTimeInMin := curHour*60 + curMin

			// iterate 1: from hhmmBegin to "2400" (midnight, exclusive)
			for k, _ := c.Seek([]byte(hhmmBegin)); k != nil; k, _ = c.Next() {
				hhmm := string(k[:4])
				w7 := string(k[4:11])
				hour, _ := strconv.Atoi(hhmm[:2])
				minute, _ := strconv.Atoi(hhmm[2:])

				// check weekday
				// must check "yesterday" as the current iteration runs before midnight
				if w7[(curWeekday+6)%7] == '0' {
					continue
				}

				// check if the time difference is a multiple of 30 minutes
				if (minute-curMin)%30 != 0 {
					continue
				}

				// Compute daily index
				timeInMin := hour*60 + minute
				dailyIndex := (endTimeInMin + 24*60 - timeInMin) / 30

				// send push
				deviceId := string(k[12:])
				trainingId := int((curWeekday + 6) % 7)
				if dailyIndex < 16 {
					//-- practice --//
					fmt.Printf("deviceId=%s, trainingId=%d, dailyIndex=%d\n", deviceId, trainingId, dailyIndex)
					sendPracticeNotification(deviceId, trainingId, dailyIndex)
				} else {
					//-- test --//
					sendTestNotification(deviceId, trainingId)
				}
			}

			// iterate 2: from "0000"(midnight, inclusive) to hhmmEnd
			for k, _ := c.First(); k != nil && string(k[:4]) <= hhmmEnd; k, _ = c.Next() {
				hhmm := string(k[:4])
				w7 := string(k[4:11])
				hour, _ := strconv.Atoi(hhmm[:2])
				minute, _ := strconv.Atoi(hhmm[2:])

				// check weekday
				if w7[curWeekday] == '0' {
					continue
				}

				// check if the time difference is a multiple of 30 minutes
				if (minute-curMin)%30 != 0 {
					continue
				}

				// Compute daily index
				timeInMin := hour*60 + minute
				dailyIndex := (endTimeInMin - timeInMin) / 30

				// send push
				deviceId := string(k[12:])
				trainingId := int(curWeekday)
				if dailyIndex < 16 {
					//-- practice --//
					sendPracticeNotification(deviceId, trainingId, dailyIndex)
				} else {
					//-- test --//
					sendTestNotification(deviceId, trainingId)
				}
			}
		}

		return nil
	})
}

func sendPracticeNotification(deviceId string, trainingId int, dailyIndex int) error {
	// get token
	token := GetToken(deviceId)

	fmt.Printf("Token: %s\n", token)

	// Create payload
	data := map[string]string{
		"task":       TASK_PRACTICE,
		"trainingId": strconv.Itoa(trainingId),
		"dailyIndex": strconv.Itoa(dailyIndex),
	}

	// send
	_, err := fcmClient.Send(context.TODO(), &messaging.Message{
		Data:    data,
		Token:   token,
		Android: &messaging.AndroidConfig{Priority: "high"},
	})

	return err
}

func sendTestNotification(deviceId string, trainingId int) error {
	// get token
	token := GetToken(deviceId)

	// Create payload
	data := map[string]string{
		"task":       TASK_TEST,
		"trainingId": strconv.Itoa(trainingId),
	}

	// send
	_, err := fcmClient.Send(context.TODO(), &messaging.Message{
		Data:    data,
		Token:   token,
		Android: &messaging.AndroidConfig{Priority: "high"},
	})

	return err
}

// Handler for setting the training schedule
func HandleSetTraining(c *gin.Context) {
	// Read data
	var payload setTrainingPayload
	if err := c.BindJSON(&payload); err != nil {
		c.Status(http.StatusNotAcceptable)
		return
	}

	if payload.Schedule != nil {
		//-- Set training --//
		schedule := *payload.Schedule

		// Is schedule valid?
		if err := validateSchedule(schedule); err != nil {
			log.Printf("schedule validation error: %v", err)
			c.Status(http.StatusNotAcceptable)
			return
		}

		// clear old schedule
		if err := clearTraining(payload.DeviceId); err != nil {
			log.Printf("disableTraining failed the training error: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		// Set training
		if err := setTraining(payload.DeviceId, schedule); err != nil {
			log.Printf("enableTraining failed the training error: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
	} else {
		//-- Clear training --//
		if err := clearTraining(payload.DeviceId); err != nil {
			log.Printf("disableTraining failed the training error: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
	}
}

// Check the validity of schedule raw string
func validateSchedule(schedule string) error {
	// len == 11
	if len(schedule) != 11 {
		return errors.New("wrong length")
	}

	// hour
	hour, err := strconv.Atoi(schedule[:2])
	if err != nil {
		return err
	} else if hour < 0 || hour >= 24 {
		return errors.New("wrong hour")
	}

	// minute
	minute, err := strconv.Atoi(schedule[2:4])
	if err != nil {
		return err
	} else if minute < 0 || minute >= 60 || minute%15 != 0 {
		return errors.New("wrong minute")
	}

	// weekdays
	for i := 4; i < 11; i++ {
		if schedule[i] != '0' && schedule[i] != '1' {
			return errors.New("wrong weekday")
		}
	}

	// passed all tests
	return nil
}

// Set the training schedule for a device
// REQUIRE: `schedule` must be validated in advance.
func setTraining(deviceId string, schedule string) error {
	return db.Update(func(t *bolt.Tx) error {
		// open buckets
		devices := t.Bucket(BUCKET_DEVICES)
		trainingSchedules := t.Bucket(BUCKET_TRAINING_SCHEDULES)

		// prepare keys
		deviceIdSchedule := deviceId + DB_DELIM + "trainingSchedule"
		scheduleDeviceId := schedule + DB_DELIM + deviceId

		// add new schedule to `trainingSchedules`
		if err := trainingSchedules.Put([]byte(scheduleDeviceId), []byte("")); err != nil {
			return err
		}

		// add new schedule to `devices`
		if err := devices.Put([]byte(deviceIdSchedule), []byte(schedule)); err != nil {
			return err
		}

		return nil
	})
}

// Disable training for a device
func clearTraining(deviceId string) error {
	return db.Update(func(t *bolt.Tx) error {
		// open buckets
		devices := t.Bucket(BUCKET_DEVICES)
		trainingSchedules := t.Bucket(BUCKET_TRAINING_SCHEDULES)

		// prepare keys
		deviceIdSchedule := deviceId + DB_DELIM + "trainingSchedule"

		// read old schedule
		hhmmwwwwwww := string(devices.Get([]byte(deviceIdSchedule)))
		oldScheduleKey := hhmmwwwwwww + DB_DELIM + deviceId

		// delete from `trainingSchedules`
		if err := trainingSchedules.Delete([]byte(oldScheduleKey)); err != nil {
			return err
		}

		// delete from `devices`
		if err := devices.Delete([]byte(deviceIdSchedule)); err != nil {
			return err
		}

		return nil
	})
}
