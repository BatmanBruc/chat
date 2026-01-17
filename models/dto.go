package models

import "time"

// CreateChatRequest represents the request to create a chat
type CreateChatRequest struct {
	Title string `json:"title" validate:"required,min=1,max=200"`
}

// CreateMessageRequest represents the request to create a message
type CreateMessageRequest struct {
	Text string `json:"text" validate:"required,min=1,max=5000"`
}

// ChatResponse represents the chat response
type ChatResponse struct {
	ID        uint              `json:"id"`
	Title     string            `json:"title"`
	CreatedAt time.Time         `json:"created_at"`
	Messages  []MessageResponse `json:"messages,omitempty"`
}

// MessageResponse represents the message response
type MessageResponse struct {
	ID        uint      `json:"id"`
	ChatID    uint      `json:"chat_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
