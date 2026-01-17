package models

import (
	"time"
)

type Message struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ChatID    uint      `json:"chat_id" gorm:"not null;index" validate:"required"`
	Text      string    `json:"text" gorm:"not null;size:5000" validate:"required,min=1,max=5000"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
