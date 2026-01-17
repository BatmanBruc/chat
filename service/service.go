package service

import (
	"chat-api/models"
	"context"
	"fmt"
	"strings"
)

type ChatRepository interface {
	Create(ctx context.Context, chat *models.Chat) (*models.Chat, error)
	Get(ctx context.Context, id uint, limit int) (*models.Chat, error)
	Delete(ctx context.Context, id uint) error
	CreateMessage(ctx context.Context, id uint, message *models.Message) (*models.Message, error)
}

type ChatService interface {
	CreateChat(ctx context.Context, title string) (*models.Chat, error)
	GetChat(ctx context.Context, id uint, limit int) (*models.Chat, error)
	DeleteChat(ctx context.Context, id uint) error
	SendMessage(ctx context.Context, chatID uint, text string) (*models.Message, error)
}

type service struct {
	repo ChatRepository
}

func NewChatService(repo ChatRepository) ChatService {
	return &service{
		repo,
	}
}

func (s *service) CreateChat(ctx context.Context, title string) (*models.Chat, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("chat title cannot be empty")
	}
	if len(title) > 200 {
		return nil, fmt.Errorf("chat title cannot exceed 200 characters")
	}

	chat := &models.Chat{
		Title: title,
	}

	return s.repo.Create(ctx, chat)
}

func (s *service) GetChat(ctx context.Context, id uint, limit int) (*models.Chat, error) {
	if id == 0 {
		return nil, fmt.Errorf("chat ID must be greater than 0")
	}
	if limit == 0 {
		limit = 20
	} else if limit < 0 {
		return nil, fmt.Errorf("limit cannot be negative")
	} else if limit > 100 {
		return nil, fmt.Errorf("limit cannot exceed 100 messages")
	}

	chat, err := s.repo.Get(ctx, id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	return chat, nil
}

func (s *service) DeleteChat(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("chat ID must be greater than 0")
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) SendMessage(ctx context.Context, chatID uint, text string) (*models.Message, error) {
	if chatID == 0 {
		return nil, fmt.Errorf("chat ID must be greater than 0")
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("message text cannot be empty")
	}
	if len(text) > 5000 {
		return nil, fmt.Errorf("message text cannot exceed 5000 characters")
	}

	message := &models.Message{
		ChatID: chatID,
		Text:   text,
	}

	return s.repo.CreateMessage(ctx, chatID, message)
}
