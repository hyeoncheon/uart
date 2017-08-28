package models

// Test coverage: 100% (without interface methods)

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"
)

// MessageMap is structure for connection messages and members
type MessageMap struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	MemberID  uuid.UUID `json:"member_id" db:"member_id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	IsSent    bool      `json:"is_sent" db:"is_sent"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	IsBCC     bool      `json:"is_bcc" db:"is_bcc"`
}

// String returns human readable string of model message map.
func (m MessageMap) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// MessageMaps is an array of message map
type MessageMaps []MessageMap

// String returns human readable string of model message map.
func (m MessageMaps) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}
