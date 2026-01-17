package models

import (
	"time"
)

type Chat struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null;size:200" validate:"required,min=1,max=200"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []Message `json:"messages,omitempty" gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
}
