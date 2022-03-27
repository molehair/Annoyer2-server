package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const DEBUG_MODE = false
const SESSION_NAME = "mysession"
const LOG_DIR = "logs" // used on running docker
const LOG_GIN_FILE = "gin.log"
const LOG_OTHERS_FILE = "others.log"

var COOKIE_KEY = []byte("r9e3wihfe2i0")

func Initialization() {
	// The order MATTERS.

	// log
	if !DEBUG_MODE {
		os.MkdirAll(LOG_DIR, os.ModePerm)
		f, err := os.Create(filepath.Join(LOG_DIR, LOG_OTHERS_FILE))
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// database
	if err := InitializeDatabase(); err != nil {
		log.Fatalf("%v", err)
		return
	}

	// firebase
	if err := InitializeFirebase(); err != nil {
		log.Fatalf("%v", err)
		return
	}

	// auth
	if err := InitializeAuth(); err != nil {
		log.Fatalf("%v", err)
		return
	}

	// cron
	if err := InitializeCron(); err != nil {
		log.Fatalf("%v", err)
		return
	}

	// training
	if err := InitializeTraining(); err != nil {
		log.Fatalf("%v", err)
		return
	}
}

func main() {
	// Initialization
	Initialization()

	// release mode for gin?
	if !DEBUG_MODE {
		gin.SetMode(gin.ReleaseMode)
	}

	// Disable Console Color, you don't need console color when writing the logs to file.
	gin.DisableConsoleColor()

	// Logging
	os.MkdirAll(LOG_DIR, os.ModePerm)
	f, _ := os.Create(filepath.Join(LOG_DIR, LOG_GIN_FILE))
	if DEBUG_MODE {
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	} else {
		gin.DefaultWriter = io.MultiWriter(f)
	}

	// Agent: "just like a new"
	r := gin.New()

	// no proxy allowed
	r.SetTrustedProxies(nil)

	// basic security measure
	r.Use(helmet.Default())

	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// TODO: https://github.com/gin-contrib/cors

	// enable sessions
	store := cookie.NewStore(COOKIE_KEY)
	r.Use(sessions.Sessions(SESSION_NAME, store))

	r.POST("/refreshToken", HandleRefreshToken)
	r.POST("/setTraining", HandleSetTraining)

	// test functions for debugging
	if DEBUG_MODE {
		addDebugFunctions(r)
	}

	r.Run()
}
