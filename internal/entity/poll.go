package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Field struct {
	ID      uint      `gorm:"primaryKey;autoIncrement"`
	PollID  uuid.UUID `gorm:"type:uuid"`
	Desc    string    `                                json:"desc"`
	Procent float32   `                                json:"procent"`
}

type Poll struct {
	ID        uuid.UUID `json:"id"         gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	AuthorID  string    `json:"author_id"`
	Desc      string    `json:"desc"`
	Fields    []Field   `json:"fields"     gorm:"foreignKey:PollID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	DeleteIn  time.Time `json:"delete_in"`
	Closed    bool      `json:"closed"`
}

func (p *Poll) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}
