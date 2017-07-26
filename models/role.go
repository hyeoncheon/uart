package models

import (
	"encoding/json"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// Role is used to set member's privilege for each apps.
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	AppID       uuid.UUID `json:"app_id" db:"app_id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Description string    `json:"description" db:"description"`
	Rank        int       `json:"rank" db:"rank"`
	IsReadonly  bool      `json:"is_readonly" db:"is_readonly"`
}

// standard role codes
const (
	RCAdmin   = "admin"
	RCUser    = "user"
	RCAppMan  = "appman"
	RCUserMan = "userman"
	RCLeader  = "leader"
)

// String returns pretty printable string of this model.
func (r Role) String() string {
	return r.App().String() + "." + r.Name
}

// App returns an app instance of the role
func (r Role) App() *App {
	app := &App{}
	DB.Find(app, r.AppID)
	return app
}

// Members returns members have the role.
// if optional flag is true, only active members are returned.
func (r Role) Members(flag ...bool) *Members {
	members := &Members{}
	q := DB.BelongsToThrough(&r, &RoleMap{})
	if len(flag) > 0 {
		q = q.Where("role_maps.is_active = ?", flag[0])
	}
	err := q.All(members)
	if err != nil {
		log.Warnf("cannot found member of %v: %v", r, err)
	}
	return members
}

// MemberCount returns count of members who has the role
func (r Role) MemberCount(isActive bool) int {
	count, _ := DB.BelongsToThrough(&r, &RoleMap{}).
		Where("role_maps.is_active = ?", isActive).
		Count(&Members{})
	return count
}

// Roles is array of Role.
type Roles []Role

// String is not required by pop and may be deleted
func (r Roles) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

// Validate gets run every time you call a "pop.Validate" method.
func (r *Role) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: r.Name, Name: "Name"},
		&validators.StringIsPresent{Field: r.Code, Name: "Code"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (r *Role) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (r *Role) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

//// Role Map

// RoleMap is a mapping object for role and member.
type RoleMap struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	RoleID    uuid.UUID `json:"role_id" db:"role_id"`
	MemberID  uuid.UUID `json:"member_id" db:"member_id"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// Role returns associated role instance of this map.
func (rm RoleMap) Role() *Role {
	role := &Role{}
	err := DB.Find(role, rm.RoleID)
	if err != nil {
		log.Errorf("cannot found role for rolemap %v (%v+%v)",
			rm.ID, rm.RoleID, rm.MemberID)
	}
	return role
}

// Member returns associated member instance of this map.
func (rm RoleMap) Member() *Member {
	member := &Member{}
	err := DB.Find(member, rm.MemberID)
	if err != nil {
		log.Errorf("cannot found member for rolemap %v (%v+%v)",
			rm.ID, rm.RoleID, rm.MemberID)
	}
	return member
}
