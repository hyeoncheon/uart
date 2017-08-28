package models

// TODO REVIEW REQUIRED
// Test coverage: 100% (without interface methods)

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// common constants
const (
	DefaultSortMembers = "created_at"
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
	APIKey    string    `json:"api_key" db:"api_key"`
}

// String returns pretty printable string of this model.
func (m Member) String() string {
	if m.Email == "" {
		return "Unknown"
	}
	return m.Name + " ." + m.ID.String()[0:6]
}

// IsNil return true if the member's ID is nil. otherwise return true.
func (m *Member) IsNil() bool {
	return m.ID == uuid.Nil
}

//** actions, relational accessor and functions below:

// Grant create an access grant for given member to the app
func (m Member) Grant(tx *pop.Connection, app *App, scope string) error {
	log.Infof("%v granting access right to app %v", m, app)
	grant := &AccessGrant{}
	err := tx.BelongsTo(&m).Where("app_id = ?", app.ID).First(grant)
	if err != nil {
		log.Warnf("no grant found for %v to %v: %v", app, m, err)
		return tx.Create(&AccessGrant{
			AppID:       app.ID,
			MemberID:    m.ID,
			AccessCount: 1,
			Scope:       scope,
		})
	}
	for _, s := range strings.Split(scope, " ") {
		if !strings.Contains(" "+grant.Scope+" ", " "+s+" ") {
			grant.Scope = grant.Scope + " " + s
		} //! bad but easy: just append each scopes as string
	}
	return tx.Save(grant)
}

// Revoke decouples the app and given member, returns database status
// Revoke does not consider scope.
func (m Member) Revoke(tx *pop.Connection, app *App) error {
	log.Infof("revoke access to %v by %v", app.Name, m.Name)
	grant := &AccessGrant{}
	err := tx.BelongsTo(&m).Where("app_id = ?", app.ID).First(grant)
	if err != nil {
		log.Errorf("cannot found grant for app %v to %v: %v", app, m, err)
		return errors.New("GrantNotFound")
	}
	return tx.Destroy(grant)
}

// Granted checks if the member have granted for given app.
// additionally it increase reference count as access count.
func (m Member) Granted(appID uuid.UUID, scope string) bool {
	grant := &AccessGrant{}
	err := DB.BelongsTo(&m).Where("app_id = ?", appID).First(grant)
	if err != nil {
		log.Warn("error while getting grant apps.", err)
		return false
	}

	// check each grants separately
	for _, s := range strings.Split(scope, " ") {
		if !strings.Contains(" "+grant.Scope+" ", " "+s+" ") {
			log.Warnf("grant found but no scope %v. reject!", s)
			return false
		}
	}
	grant.AccessCount++
	DB.Save(grant)
	return true
}

// Grants returns all grants of the member.
func (m Member) Grants() *AccessGrants {
	grants := &AccessGrants{}
	err := DB.BelongsTo(&m).Order(grantsDefaultSort).All(grants)
	if err != nil {
		log.Warn("no grants found: ", err)
	}
	return grants
}

// GrantedApps returns the member's associcated granted apps
func (m Member) GrantedApps() *Apps {
	apps := &Apps{}
	err := DB.BelongsToThrough(&m, &AccessGrants{}).All(apps)
	if err != nil {
		log.Warn("no associated apps found: ", err)
	}
	return apps
}

// HasRole return true if the member has the role regardless of activated.
func (m Member) HasRole(roleID uuid.UUID) bool {
	err := DB.BelongsToThrough(&m, &RoleMap{}).Find(&Role{}, roleID)
	if err != nil {
		return false
	}
	return true
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
func (m Member) AppRoles(appID uuid.UUID, flag ...bool) *Roles {
	roles := &Roles{}
	Q := DB.BelongsToThrough(&m, &RoleMap{})
	if len(flag) > 0 {
		Q = Q.Where("role_maps.is_active = ?", flag[0])
	}
	err := Q.Where("roles.app_id = ?", appID).All(roles)
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
	roles := m.AppRoles(app.ID, true)
	for _, r := range *roles {
		ret = append(ret, r.Code)
	}
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

// * related to messaging subsystem

// MessageMarkAsSent marks given message's message map for the member.
func (m *Member) MessageMarkAsSent(id uuid.UUID) error {
	link := MessageMaps{}
	DB.
		Where("member_id = ?", m.ID).
		Where("message_id = ?", id).
		All(&link)
	switch len(link) {
	case 1:
		link[0].IsSent = true
		DB.Save(&link[0])
	case 0:
		log.Errorf("OOPS! mark-as-read request for %v:%v but link not exists!", m, id)
		return errors.New("link not found")
	default:
		log.Errorf("OOPS! more than one message link! IS_IT_POSSIBLE? %v:%v", m, id)
		return errors.New("so many links")
	}
	return nil
}

// Messengers returns messengers belonging to the member.
func (m *Member) Messengers(args ...int) *Messengers {
	messengers := &Messengers{}
	q := DB.BelongsTo(m).Order(messengersDefaultSort)
	if len(args) > 0 {
		q.Where("priority = ?", args[0])
	}
	err := q.All(messengers)
	if err != nil {
		log.Warnf("cannot found messengers", err)
	}
	return messengers
}

/*
// Alerters returns messenger for alert of the member.
func (m *Member) Alerters() *Messengers {
	return m.Messengers(MsgPriAlert)
}

// Notifiers returns messenger for notification of the member.
func (m *Member) Notifiers() *Messengers {
	return m.Messengers(MsgPriNote)
}
*/

// PrimaryAlert returns primary messenger of the member.
func (m *Member) PrimaryAlert() *Messenger {
	messenger := &Messenger{}
	err := DB.BelongsTo(m).
		Where("priority = ?", MessengerPriority["Alert"]).
		Where("is_primary = ?", true).First(messenger)
	if err != nil {
		log.Warn("cannot found primary messenger ", err)
	}
	return messenger
}

// PrimaryNotifier returns primary messenger of the member.
func (m *Member) PrimaryNotifier() *Messenger {
	messenger := &Messenger{}
	err := DB.BelongsTo(m).
		Where("priority = ?", MessengerPriority["Notification"]).
		Where("is_primary = ?", true).First(messenger)
	if err != nil {
		log.Warnf("cannot found primary messenger", err)
	}
	return messenger
}

//** implementations for interfaces ---------------------------------

// GetID implements Owner interface
func (m *Member) GetID() interface{} {
	return m.ID
}

// QueryParams implements Belonging interface
func (m *Member) QueryParams() QueryParams {
	return QueryParams{}
}

// QueryParams implements Belonging interface
func (m *Members) QueryParams() QueryParams {
	return QueryParams{
		DefaultSort: DefaultSortMembers,
	}
}

// OwnedBy implements Belonging interface
func (m *Member) OwnedBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q
}

// OwnedBy implements Belonging interface
func (m *Members) OwnedBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q
}

// AccessibleBy implements Belonging interface
func (m *Member) AccessibleBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q
}

// AccessibleBy implements Belonging interface
func (m *Members) AccessibleBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q
}

//** common database/crud functions ---------------------------------

// GetMember picks a member instance with given id.
func GetMember(id interface{}) *Member {
	m := &Member{}
	if UUID, ok := id.(uuid.UUID); ok && UUID == uuid.Nil {
		return m //// to prevent database access.
	}
	err := DB.Find(m, id)
	if err != nil {
		log.Error("cannot found member with id: ", id)
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

		err = member.Grant(tx, uart, AppDefaultScope)
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

//** array model for base model --------------------------------------------

// Members is an array of Members.
type Members []Member

// String returns json marshalled representation of Members
func (m Members) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
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
