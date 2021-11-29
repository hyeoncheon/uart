package models

// TODO REVIEW REQUIRED

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/logger"
	"github.com/gobuffalo/pop/v6"
)

// DB and others: shared variables for models
var DB *pop.Connection

var log = logger.NewLogger("Debug").WithField("category", "models")
var securityLog = log.WithField("category", "security")
var isDev = false

func init() {
	var err error
	env := envy.Get("GO_ENV", "development")
	DB, err = pop.Connect(env)
	if err != nil {
		log.Fatal(err)
	}
	pop.Debug = env == "development"

	if env == "development" {
		isDev = true
		log.Info("models initialized in development mode")
	}
}

// Logger set logger for models.
func Logger(logger buffalo.Logger) {
	log = logger.WithField("category", "models")
	log.Info("models initialized")
}
