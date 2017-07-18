package models

import (
	"github.com/gobuffalo/envy"
	"github.com/markbates/pop"
	"github.com/sirupsen/logrus"
)

var DB *pop.Connection
var log = logrus.New().WithField("category", "model")
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
		log.Logger.Level = logrus.DebugLevel
		isDev = true
		log.Info("models initialized with log level ", log.Logger.Level)
	}
}
