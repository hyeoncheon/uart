package models

// Test coverage: 100% (without interface methods)

import (
	"reflect"

	"github.com/markbates/pop"
)

// Owner is an interface for model which have belongings
type Owner interface {
	GetID() interface{}
}

// QueryParams is structure contains query parameters
type QueryParams struct {
	DefaultSort string
	Sort        string

	FilterKey   string
	FilterValue interface{}
}

// Belonging is an interface for model which have foreign reference.
type Belonging interface {
	QueryParams() QueryParams
	OwnedBy(*pop.Query, Owner, ...bool) *pop.Query
	AccessibleBy(*pop.Query, Owner, ...bool) *pop.Query
}

// FindMy find a single instance of belonging accessible by owner.
func FindMy(q *pop.Query, m *Member, b Belonging, id interface{}) error {
	log.Debug("--------------------- FindMy!")
	q = b.AccessibleBy(q, m)
	if err := q.Find(b, id); err != nil {
		log.Errorf("cannot found having %v %v", reflect.TypeOf(b), id)
		log.Errorf("query error: %v", err)
		return err
	}
	return nil
}

// FindMyOwn find a single instance of belonging accessible by owner.
func FindMyOwn(q *pop.Query, m *Member, b Belonging, id interface{}) error {
	log.Debug("--------------------- FindMyOwn!")
	q = b.OwnedBy(q, m)
	if err := q.Find(b, id); err != nil {
		log.Errorf("cannot found having %v %v", reflect.TypeOf(b), id)
		log.Errorf("query error: %v", err)
		return err
	}
	return nil
}

// AllMy collect all instances of belongings accessible by owner.
func AllMy(q *pop.Query, m *Member, b Belonging, all ...bool) error {
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
func AllMyOwn(q *pop.Query, m *Member, b Belonging, all ...bool) error {
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
