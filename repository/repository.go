package repository

import (
	"chat-api/models"
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ChatRepository interface {
	Create(ctx context.Context, chat *models.Chat) (*models.Chat, error)
	Get(ctx context.Context, id uint, limit int) (*models.Chat, error)
	Delete(ctx context.Context, id uint) error
	CreateMessage(ctx context.Context, id uint, message *models.Message) (*models.Message, error)
}

type Logger interface {
	Log(operation, table string, details string, durationMs float64, err error)
}

type Repository struct {
	db     *gorm.DB
	logger Logger
}

func NewRepository(db *gorm.DB, logger Logger) ChatRepository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) Create(ctx context.Context, chat *models.Chat) (*models.Chat, error) {
	start := time.Now()

	result := r.db.WithContext(ctx).Create(chat)
	err := result.Error

	duration := time.Since(start)
	durationMs := float64(duration.Nanoseconds()) / 1e6

	r.logger.Log("Create", "chats", fmt.Sprintf("chat: %+v", chat), durationMs, result.Error)

	if err != nil {
		return nil, fmt.Errorf("failed create chat: %w", err)
	}
	return chat, nil
}

func (r *Repository) Get(ctx context.Context, id uint, limit int) (*models.Chat, error) {
	start := time.Now()

	var chat models.Chat

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		resultC := tx.First(&chat, id)

		if resultC.Error != nil {
			return fmt.Errorf("failed get chat: %w", resultC.Error)
		}

		return tx.Where("chat_id = ?", id).Order("updated_at DESC").Limit(limit).Find(&chat.Messages).Error
	}, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead, // Предотвращает изменения во время транзакции
		ReadOnly:  true,
	})

	duration := time.Since(start)
	durationMs := float64(duration.Nanoseconds()) / 1e6

	r.logger.Log("Transaction: get, get", "chats, messages", fmt.Sprintf("chat_id: %d, limit: %d", id, limit), durationMs, err)

	return &chat, err
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	start := time.Now()

	var chat models.Chat
	result := r.db.WithContext(ctx).Delete(&chat, id)
	err := result.Error

	duration := time.Since(start)
	durationMs := float64(duration.Nanoseconds()) / 1e6

	r.logger.Log("Delete", "chats", fmt.Sprintf("chat_id: %d", id), durationMs, result.Error)

	if err != nil {
		return fmt.Errorf("failed delete chat: %w", err)
	}
	return nil
}

func (r *Repository) CreateMessage(ctx context.Context, id uint, message *models.Message) (*models.Message, error) {
	start := time.Now()

	result := r.db.WithContext(ctx).Create(message)
	err := result.Error

	duration := time.Since(start)
	durationMs := float64(duration.Nanoseconds()) / 1e6

	r.logger.Log("Create", "messages", fmt.Sprintf("chat_id: %d, message: %+v", id, message), durationMs, result.Error)

	if err != nil {
		return nil, fmt.Errorf("failed create message: %w", err)
	}
	return message, nil
}
