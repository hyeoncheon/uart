package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// Member is the main model which presents the user.
type Member struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Mobile    string    `json:"mobile" db:"mobile"`
	Icon      string    `json:"icon" db:"icon"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	Note      string    `json:"note" db:"note"`
}

// String returns pretty printable string of this model.
func (m Member) String() string {
	if m.Email == "" {
		return "Empty"
	}
	return m.Name + " (" + m.Email + ")"
}

//// actions and relational functions below:

// HasGrantFor checks if the member have granted for given app.
// additionally it increase reference count as access count.
func (m Member) HasGrantFor(appID uuid.UUID) bool {
	grant := &AccessGrant{}
	err := DB.Where("member_id = ? AND app_id = ?", m.ID, appID).
		Where("is_revoked = ?", false).
		First(grant)
	if err != nil {
		log.Error("error while getting grant apps.", err)
		return false
	}
	grant.AccessCount++
	DB.Save(grant)
	return true
}

// HasRole return true if the member has the role regardless of activated.
func (m Member) HasRole(roleID uuid.UUID) bool {
	err := DB.BelongsToThrough(&m, &RoleMap{}).Find(&Role{}, roleID)
	if err != nil {
		return false
	}
	return true
}

// GrantedApps returns the member's associcated granted apps
func (m Member) GrantedApps() *Apps {
	apps := &Apps{}
	err := DB.BelongsToThrough(&m, &AccessGrants{}).
		Where("is_revoked = ?", false).Order(membersDefaultSort).All(apps)
	if err != nil {
		log.Error("OOPS! cannot found associated apps:", err)
	}
	return apps
}

// AddRole create mapping object for the member.
func (m *Member) AddRole(tx *pop.Connection, r *Role, active ...bool) error {
	log.Infof("assign role %v to member %v", r, m)
	isActive := false
	if len(active) > 0 {
		isActive = active[0]
	}
	return tx.Create(&RoleMap{
		MemberID: m.ID,
		RoleID:   r.ID,
		IsActive: isActive,
	})
}

// RemoveRole remove rolemap between the member and given role.
func (m *Member) RemoveRole(tx *pop.Connection, r *Role) error {
	log.Debugf("decouple role %v from member %v.", r, m)
	rolemap := &RoleMap{}
	err := tx.BelongsTo(m).Where("role_id = ?", r.ID).First(rolemap)
	if err != nil {
		log.Errorf("cannot found rolemap for %v+%v(%v+%v)", m, r, m.ID, r.ID)
		return err
	}
	err = tx.Destroy(rolemap)
	if err != nil {
		log.Errorf("cannot delete the rolemap for %v: %v", rolemap.ID, err)
	}
	return err
}

// AppRoles returns associated roles of given app, assigned to the member.
func (m Member) AppRoles(app App, flag ...bool) *Roles {
	roles := &Roles{}
	Q := DB.BelongsToThrough(&m, &RoleMap{})
	if len(flag) > 0 {
		Q = Q.Where("role_maps.is_active = ?", flag[0])
	}
	err := Q.Where("roles.app_id = ?", app.ID).All(roles)
	if err != nil {
		log.Warn("cannot found associated roles: ", err)
	}
	return roles
}

// GetAppRoleCodes returns the member's active role codes of given app.
func (m Member) GetAppRoleCodes(appCode string) []string {
	ret := []string{}
	app := GetAppByCode(appCode)
	if app == nil {
		log.Error("OOPS! cannot found app with given code!")
		return ret
	}
	roles := m.AppRoles(*app, true)
	for _, r := range *roles {
		ret = append(ret, r.Code)
	}
	log.Debug("---- AppRoleCode ---- ", ret)
	return ret
}

// Roles returns the member's associcated roles
func (m Member) Roles() *Roles {
	roles := &Roles{}
	err := DB.BelongsToThrough(&m, &RoleMap{}).All(roles)
	if err != nil {
		log.Error("OOPS! cannot found associated roles:", err)
	}
	return roles
}

// Credentials returns the member's associated credentials
func (m Member) Credentials() *Credentials {
	creds := &Credentials{}
	err := DB.BelongsTo(&m).All(creds)
	if err != nil {
		log.Error("OOPS! cannot found associated credentials: ", err)
	}
	return creds
}

// CredentialCount returns count of associated credentials
func (m Member) CredentialCount() int {
	count, err := DB.BelongsTo(&m).Count(&Credentials{})
	if err != nil {
		log.Error("cannot count associated credentials: ", err)
	}
	return count
}

// AccessGrantCount returns count of associated access grants
func (m Member) AccessGrantCount() int {
	count, err := DB.BelongsTo(&m).Count(&AccessGrants{})
	if err != nil {
		log.Error("cannot count associated access grants: ", err)
	}
	return count
}

//// Generic model operation functions below:

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

// CreateMember creates a member with an associated credential
func CreateMember(cred *Credential) (*Member, error) {
	member := &Member{}
	member.IsActive = false
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
			log.Error("OOPS! cannot create member! ", err)
			return err
		}
		cred.MemberID = member.ID
		err = tx.Create(cred)
		if err != nil {
			log.Error("OOPS! cannot create credential! ", err)
			return err
		}

		uart := GetAppByCode("uart")
		if uart == nil {
			member.IsActive = true
			err = tx.Save(member)
			if err != nil {
				log.Error("OOPS! cannot activate this admin user! ", err)
			}
			uart = createUARTApp(tx)
			if uart == nil {
				log.Error("OOPS! CRITICAL! cannot create UART itself!")
				return errors.New("UART CREATION ERROR")
			}
			log.Info("FIRST FLIGHT! register my self ", uart)
			err = member.AddRole(tx, uart.GetRole(tx, RCAdmin), true)
		} else {
			err = member.AddRole(tx, uart.GetRole(tx, RCUser))
		}
		if err != nil {
			log.Errorf("OOPS! cannot assign a role to member: %v", err)
			return err
		}

		err = uart.Grant(tx, member)
		if err != nil {
			log.Errorf("OOPS! cannot grant %v to %v: %v", uart, member, err)
			return err
		}
		return nil
	})

	if err != nil {
		log.Error("transaction error while member registration: ", err)
		return member, err
	}
	log.Infof("new member %v registered successfully.", member)
	return member, nil
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

// QueryAndParams implementation (Searchable)
func (m Members) QueryAndParams(c buffalo.Context) (*pop.Query, SearchParams) {
	sp := newSearchParams(c)
	sp.DefaultSort = membersDefaultSort
	q := DB.Q()
	return q, sp
}

// Validate gets run every time you call a "pop.Validate" method.
func (m *Member) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: m.Name, Name: "Name"},
		&validators.StringIsPresent{Field: m.Email, Name: "Email"},
		&validators.StringIsPresent{Field: m.Icon, Name: "Icon"},
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
