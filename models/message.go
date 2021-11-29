package models

// Test coverage: 100% (without interface methods)

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// common constants
const (
	DefaultSortMessages = "created_at desc"
)

// constants for messaging/logging subsystem
const (
	MsgFacCore  = "core"
	MsgFacAuth  = "auth"
	MsgFacApp   = "app"
	MsgFacUser  = "user"
	MsgFacMesg  = "messaging"
	MsgFacCron  = "scheduler"
	MsgFacSecu  = "security"
	MsgPriEmerg = 0 // RESERVED
	MsgPriAlert = 1 // for alert
	MsgPriCrit  = 2 // FATAL
	MsgPriErr   = 3
	MsgPriWarn  = 4
	MsgPriNote  = 5 // for notification
	MsgPriInfo  = 6
	MsgPriDebug = 7
)

// MsgPri is a map for name to code referencing of message priority
var MsgPri = map[string]int{
	"Emerg": 0,
	"Alert": 1,
	"Crit":  2,
	"Err":   3,
	"Warn":  4,
	"Note":  5,
	"Info":  6,
	"Debug": 7,
}

// MsgPriReverse is a map for code to name referencing of message priority
var MsgPriReverse = map[int]string{
	0: "Emerg",
	1: "Alert",
	2: "Crit",
	3: "Err",
	4: "Warn",
	5: "Note",
	6: "Info",
	7: "Debug",
}

// Message is a structure for messaging/logging subsystem
type Message struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	MemberID  uuid.UUID `json:"member_id" db:"member_id"`
	Subject   string    `json:"subject" db:"subject"`
	Content   string    `json:"content" db:"content"`
	AppCode   string    `json:"app_code" db:"app_code"`
	Facility  string    `json:"facility" db:"facility"`
	Priority  int       `json:"priority" db:"priority"`
	IsLog     bool      `json:"is_log" db:"is_log"`
}

//** rendering helpers for templates --------------------------------

// String returns representation of Messages
func (m Message) String() string {
	return MsgPriReverse[m.Priority] + ":" + m.Subject
}

// PriorityString returns human readable string of the message's priority
func (m Message) PriorityString() string {
	return MsgPriReverse[m.Priority]
}

// AppName returns associated app name or code itself if app is not found.
func (m Message) AppName() string {
	if app := GetAppByCode(m.AppCode); app != nil {
		return app.Name
	}
	return m.AppCode
}

// Owner returns associated member instance (owner of the message)
func (m Message) Owner() *Member {
	member := GetMember(m.MemberID)
	return member
}

// MemberMap returns message map for the message and given member.
// This method will be called directly from template(index, show)
func (m Message) MemberMap(memberID interface{}) *MessageMap {
	messageMap := &MessageMap{}
	DB.Where("member_id = ? and message_id = ?", memberID, m.ID).
		First(messageMap)
	return messageMap
}

//** relational accessor and functions ------------------------------

//** implementations for interfaces ---------------------------------

// QueryParams implements Belonging interface
func (m *Message) QueryParams() QueryParams {
	return QueryParams{}
}

// QueryParams implements Belonging interface
func (m *Messages) QueryParams() QueryParams {
	return QueryParams{
		DefaultSort: DefaultSortMessages,
	}
}

// OwnedBy implements Belonging interface
func (m *Message) OwnedBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	if len(f) == 1 {
		q = q.Where("message_maps.is_read = ?", f[0])
	}
	return q.BelongsTo(o)
}

// OwnedBy implements Belonging interface
func (m *Messages) OwnedBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	if len(f) == 1 {
		q = q.Where("message_maps.is_read = ?", f[0])
	}
	return q.BelongsTo(o)
}

// AccessibleBy implements Belonging interface
func (m *Message) AccessibleBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	if len(f) == 1 {
		q = q.Where("message_maps.is_read = ?", f[0])
	}
	return q.BelongsToThrough(o, &MessageMaps{})
}

// AccessibleBy implements Belonging interface
func (m *Messages) AccessibleBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	if len(f) == 1 {
		q = q.Where("message_maps.is_read = ?", f[0])
	}
	return q.BelongsToThrough(o, &MessageMaps{})
}

//** common database/crud functions ---------------------------------

// NewMessage creates new message with given parameters
func NewMessage(tx *pop.Connection, sndrID interface{}, rcpts, bccs *Members, subj, cont, ac, fac string, pri int, isLog bool) *Message {
	ID := uuid.UUID{}
	if aID, ok := sndrID.(uuid.UUID); ok {
		ID = aID
	}
	message := &Message{
		MemberID: ID,
		Subject:  subj,
		Content:  cont,
		AppCode:  ac,
		Facility: fac,
		Priority: pri,
		IsLog:    isLog,
	}
	if err := tx.Create(message); err != nil {
		return nil
	}
	if rcpts != nil {
		for _, member := range *rcpts {
			if err := tx.Create(&MessageMap{
				MemberID:  member.ID,
				MessageID: message.ID,
			}); err != nil {
				return nil
			}
		}
	}
	if bccs != nil {
		for _, member := range *bccs {
			if err := tx.Create(&MessageMap{
				MemberID:  member.ID,
				MessageID: message.ID,
				IsBCC:     true,
			}); err != nil {
				return nil
			}
		}
	}
	return message
}

//** array model for base model -------------------------------------

// Messages is an array of Messages
type Messages []Message

// Validate gets run every time you call a "pop.Validate" method.
func (m *Message) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: m.Subject, Name: "Subject"},
		&validators.StringIsPresent{Field: m.Content, Name: "Content"},
		&validators.StringIsPresent{Field: m.AppCode, Name: "AppCode"},
		&validators.StringIsPresent{Field: m.Facility, Name: "Facility"},
		&validators.IntIsPresent{Field: m.Priority, Name: "Priority"},
	), nil
}
