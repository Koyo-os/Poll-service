package entity

import "github.com/google/uuid"

type Poll struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Type    string    `json:"type"`
	Payload any       `json:"payload"`
}

func NewPoll(Type string, payload any) *Poll {
	return &Poll{
		Type:    Type,
		Payload: payload,
		ID:      uuid.New(),
	}
}
