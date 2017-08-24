package jobs

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
)

// constants
const (
	DefaultQueue = "default"
)

// Handler is structure storing handler statistics data
type Handler struct {
	Name    string
	Running int
	Success int
}

func (h Handler) String() string {
	return h.Name + " " + strconv.Itoa(h.Success) + "/" + strconv.Itoa(h.Running)
}

// Handlers is a handler registry (originaly a dummy type for holding methods)
type Handlers map[string]*Handler

var handlers = Handlers{}
var env string
var w worker.Worker
var logger buffalo.Logger

// RegisterAll register all workers
func RegisterAll(app *buffalo.App) {
	env = app.Env
	w = app.Worker
	logger = app.Logger.WithField("category", "worker")

	x := reflect.TypeOf(&Handler{})
	for i := 0; i < x.NumMethod(); i++ {
		method := x.Method(i)
		if !strings.HasPrefix(method.Name, "Register") {
			continue
		}
		logger.Debugf("invoking %v...", method.Name)

		h := Handler{Name: method.Name}
		r := method.Func.Call([]reflect.Value{reflect.ValueOf(&h)})
		if err := r[0].Interface(); err != nil {
			logger.Error("OOPS! registration failed: ", handlers[method.Name])
		}
		handlers[h.Name] = &h
		logger.Infof("new background job handler %v registered", h)
	}
	logger.Infof("jobs registration completed! (%v handlers)", len(handlers))
	return
}

func countRunning(name string) {
	if h, ok := handlers[name]; ok == true {
		h.Running++
	}
}

func countSuccess(name string) {
	if h, ok := handlers[name]; ok == true {
		h.Success++
	}
}

func printStatistics(name string) {
	if h, ok := handlers[name]; ok == true {
		logger.Infof("statistics: %v", h)
	}
}
