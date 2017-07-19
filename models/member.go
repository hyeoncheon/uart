package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// MemberStatus is pseudo constant for member's current status
var MemberStatus = map[string]string{
	"New":    "new",
	"Active": "active",
	"Locked": "lockec",
}

// Member is the main model which presents the user.
type Member struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Mobile    string    `json:"mobile" db:"mobile"`
	Icon      string    `json:"icon" db:"icon"`
	Status    string    `json:"status" db:"status"`
	Note      string    `json:"note" db:"note"`
}

// String returns pretty printable string of this model.
func (m Member) String() string {
	return m.Name + " (" + m.Email + ")"
}

// GetMember picks a member instance with given id.
func GetMember(id interface{}) *Member {
	m := &Member{}
	err := DB.Find(m, id)
	if err != nil {
		log.Error("cannot found member with id: ", id)
		return nil
	}
	return m
}

// AddRole create mapping object for the member.
func (m *Member) AddRole(r *Role) error {
	log.Infof("assign role %v to member %v", r, m)
	return DB.Create(&RoleMap{
		MemberID: m.ID,
		RoleID:   r.ID,
	})
}

// GetAppRoleCodes returns the member's role codes of given app.
func (m Member) GetAppRoleCodes(app string) []string {
	rs := &Roles{}
	rmap := &RoleMap{}
	err := DB.BelongsToThrough(&m, rmap).All(rs)
	if err != nil {
		log.Warn("cannot found associated roles: ", err)
	}
	roles := []string{}
	for _, r := range *rs {
		roles = append(roles, r.Code)
	}
	log.Debug("-----------------------------------", roles)
	return roles
}

// Credentials returns the member's associated credentials
func (m Member) Credentials() *Credentials {
	creds := &Credentials{}
	err := DB.BelongsTo(&m).All(creds)
	if err != nil {
		log.Error("cannot found associated credentials: ", err)
	}
	return creds
}

// CreateMember creates a member with an associated credential
func CreateMember(cred *Credential) *Member {
	member := &Member{}
	member.Status = MemberStatus["New"]
	member.Icon = cred.AvatarURL
	member.Name = cred.Name
	member.Email = cred.Email

	excl := []string{}
	if cred.AvatarURL == "" {
		excl = append(excl, "icon")
	}

	err := DB.Transaction(func(tx *pop.Connection) error {
		err := tx.Create(member, excl...)
		if err != nil {
			return err
		}
		cred.MemberID = member.ID
		err = tx.Create(cred)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Error("transaction error while member registration: ", err)
	}

	uart := GetAppByName("UART")
	if uart == nil {
		uart = createUARTApp()
		log.Info("FIRST FLIGHT! register my self ", uart)
		err = member.AddRole(uart.GetRole("admin"))
	} else {
		err = member.AddRole(uart.GetRole("user"))
	}
	if err != nil {
		// TODO admin alert for failed role assignment
		log.Errorf("add role to member %v failed: %v", member, err)
	}

	err = uart.Grant(member)
	if err != nil {
		// TODO admin alert for failed access grant
		log.Errorf("access grant failed for %v to %v: %v", uart, member, err)
	}

	log.Infof("new member %v registered successfully.", member)
	// TODO admin notification for member registration
	return member
}

// Members is an array of Members.
type Members []Member

// String is not required by pop and may be deleted
func (m Members) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

const membersDefaultSort = "created_at"

// SearchParams implementation (Searchable)
func (m Members) SearchParams(c buffalo.Context) SearchParams {
	sp := newSearchParams(c)
	sp.DefaultSort = membersDefaultSort
	return sp
}

// Validate gets run every time you call a "pop.Validate" method.
func (m *Member) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: m.Name, Name: "Name"},
		&validators.StringIsPresent{Field: m.Email, Name: "Email"},
		&validators.StringIsPresent{Field: m.Icon, Name: "Icon"},
		&validators.StringIsPresent{Field: m.Status, Name: "Status"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (m *Member) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (m *Member) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
