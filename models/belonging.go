package models

import (
	"reflect"

	"github.com/markbates/pop"
)

type Dummy struct {
	ID interface{}
}

// QueryParams is structure contains query parameters
type QueryParams struct {
	DefaultSort string
	Sort        string
}

// MyBelonging is an interface for model which have foreign reference.
//! TODO: NEED NAME CHANGE
type MyBelonging interface {
	QueryParams() QueryParams
	OwnedBy(*pop.Query, interface{}, ...bool) *pop.Query
	AccessibleBy(*pop.Query, interface{}, ...bool) *pop.Query
}

// FindMyHaving find a single instance of belonging accessible by owner.
func FindMyHaving(q *pop.Query, m *Member, b MyBelonging, id interface{}) error {
	log.Debug("--------------------- FindMyHaving!")
	q = b.AccessibleBy(q, m)
	if err := q.Find(b, id); err != nil {
		log.Errorf("cannot found having %v %v", reflect.TypeOf(b), id)
	}
	return nil
}

// AllMyHaving collect all instances of belongings accessible by owner.
func AllMyHaving(q *pop.Query, m *Member, b MyBelonging, all ...bool) error {
	dummy := &Member{ID: m.ID}
	if len(all) == 1 {
		q = b.AccessibleBy(q, dummy, all[0])
	} else {
		q = b.AccessibleBy(q, dummy)
	}

	qp := b.QueryParams()
	if qp.DefaultSort != "" {
		q = q.Order(qp.DefaultSort)
	}

	if err := q.All(b); err != nil {
		log.Errorf("cannot found belongings %v", reflect.TypeOf(b))
	}
	return nil
}

// AllMyOwn collect all instances of belongings accessible by owner.
func AllMyOwn(q *pop.Query, m *Member, b MyBelonging, all ...bool) error {
	dummy := &Member{ID: m.ID}
	if len(all) == 1 {
		q = b.OwnedBy(q, dummy, all[0])
	} else {
		q = b.OwnedBy(q, dummy)
	}

	qp := b.QueryParams()
	if qp.DefaultSort != "" {
		q = q.Order(qp.DefaultSort)
	}

	if err := q.All(b); err != nil {
		log.Errorf("cannot found belongings %v", reflect.TypeOf(b))
	}
	return nil
}
