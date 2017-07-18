package models

import (
	"encoding/json"
	"math/rand"
	"time"
)

// Searchable Interface
//
type Searchable interface {
}

// SelectByAttrs find and store models with given search attributes
func SelectByAttrs(m Searchable, attrs map[string]interface{}) error {
	q := DB.Q()
	for attr, value := range attrs {
		q.Where(attr+" = ?", value)
	}
	return q.All(m)
}

// Object Interface (for general models)
//
type Object interface {
}

// Marshal returns JSON marshalled string from given object
func Marshal(m Object) string {
	byte, _ := json.Marshal(m)
	return string(byte)
}

// utilities

// random string generator from https://stackoverflow.com/a/31832326/1111002
var lRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = lRunes[rand.Intn(len(lRunes))]
	}
	return string(b)
}
