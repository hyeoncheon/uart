package models

// TODO REVIEW REQUIRED

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/satori/go.uuid"

	"github.com/hyeoncheon/uart/utils"
)

// Belonging Interface
//
type Belonging interface {
	RelationalOwnerQuery(uuid.UUID) *pop.Query
}

// FindMy uses belonging's RelationQuery for finding path to the belonging
// and returns member's own belonging.
func FindMy(c buffalo.Context, m *Member, b Belonging, id interface{}) error {
	var q *pop.Query
	if c.Value("member_is_admin").(bool) {
		securityLog.Debug("secu: belonging is accessed by admin")
		q = DB.Q()
	} else {
		q = b.RelationalOwnerQuery(m.ID)
	}
	err := q.Find(b, id)
	if err != nil {
		log.Error("cannot found belonging", err)
		err = errors.New("Not Found")
	}
	return err
}

// PickOne find and return single belonging regardless of access right
func PickOne(tx *pop.Connection, b Belonging, id interface{}) error {
	securityLog.Debug("WARNING: this fuction do not check access right!")
	err := tx.Find(b, id)
	if err != nil {
		log.Error("cannot pick one: ", err)
		err = errors.New("Not Found")
	}
	return err
}

// SearchParams is a structure for storing paginated query
type SearchParams struct {
	Page        int
	PerPage     int
	TotalPages  int
	Sort        string
	DefaultSort string
	FilterKey   string
	FilterValue interface{}
}

func newSearchParams(c buffalo.Context) SearchParams {
	return SearchParams{
		Page:        utils.GetIntParam(c, "page", 1, 0),
		PerPage:     utils.GetIntParam(c, "pp", 10, 200),
		Sort:        utils.GetStringParam(c, "sort", ""),
		FilterKey:   utils.GetStringParam(c, "filter", ""),
		FilterValue: utils.GetParam(c, "value"),
	}
}

// Searchable Interface
//
type Searchable interface {
	SearchParams(buffalo.Context) SearchParams
	QueryAndParams(buffalo.Context) (*pop.Query, SearchParams)
}

// All returns paginated search result for given model.
func All(c buffalo.Context, m Searchable) (SearchParams, error) {
	q, sp := m.QueryAndParams(c)
	q = q.Paginate(sp.Page, sp.PerPage)
	if sp.Sort != "" {
		for _, o := range strings.Split(sp.Sort, ",") {
			q = q.Order(o)
		}
	}
	if sp.DefaultSort != "" {
		q = q.Order(sp.DefaultSort)
	}
	if sp.FilterKey != "" {
		if s, ok := sp.FilterValue.(string); ok && s != "" {
			q = q.Where(sp.FilterKey+" LIKE ?", s)
		} else {
			q = q.Where(sp.FilterKey+" = ?", sp.FilterValue)
		}
	}
	err := q.All(m)
	sp.TotalPages = q.Paginator.TotalPages
	if err != nil {
		return sp, err
	}
	return sp, nil
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
