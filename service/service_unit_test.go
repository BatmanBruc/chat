package service

import (
	"chat-api/models"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChatRepository - мок для ChatRepository
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) Create(ctx context.Context, chat *models.Chat) (*models.Chat, error) {
	args := m.Called(ctx, chat)
	return args.Get(0).(*models.Chat), args.Error(1)
}

func (m *MockChatRepository) Get(ctx context.Context, id uint, limit int) (*models.Chat, error) {
	args := m.Called(ctx, id, limit)
	return args.Get(0).(*models.Chat), args.Error(1)
}

func (m *MockChatRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChatRepository) CreateMessage(ctx context.Context, id uint, message *models.Message) (*models.Message, error) {
	args := m.Called(ctx, id, message)
	return args.Get(0).(*models.Message), args.Error(1)
}

// TestCreateChat_EmptyTitle - тест создания чата с пустым названием
func TestCreateChat_EmptyTitle(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()

	// Тестируем пустое название
	_, err := service.CreateChat(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat title cannot be empty")

	// Тестируем название только из пробелов
	_, err = service.CreateChat(ctx, "   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat title cannot be empty")

	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateChat_TooLongTitle - тест создания чата со слишком длинным названием
func TestCreateChat_TooLongTitle(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()

	// Создаем строку длиной 201 символ
	longTitle := string(make([]byte, 201))

	_, err := service.CreateChat(ctx, longTitle)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat title cannot exceed 200 characters")

	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateChat_Success - тест успешного создания чата
func TestCreateChat_Success(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()
	inputTitle := "  Test Chat  "
	expectedTitle := "Test Chat"

	expectedChat := &models.Chat{
		ID:    1,
		Title: expectedTitle,
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(chat *models.Chat) bool {
		return chat.Title == expectedTitle
	})).Return(expectedChat, nil)

	result, err := service.CreateChat(ctx, inputTitle)

	assert.NoError(t, err)
	assert.Equal(t, expectedChat, result)
	assert.Equal(t, expectedTitle, result.Title) // Проверяем, что пробелы убраны

	mockRepo.AssertExpectations(t)
}

// TestCreateChat_RepositoryError - тест ошибки репозитория
func TestCreateChat_RepositoryError(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()
	repoError := errors.New("database error")

	mockRepo.On("Create", ctx, mock.Anything).Return((*models.Chat)(nil), repoError)

	_, err := service.CreateChat(ctx, "Valid Title")

	assert.Error(t, err)
	assert.Equal(t, repoError, err)

	mockRepo.AssertExpectations(t)
}

// TestGetChat_InvalidID - тест получения чата с некорректным ID
func TestGetChat_InvalidID(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()

	_, err := service.GetChat(ctx, 0, 20)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat ID must be greater than 0")

	mockRepo.AssertNotCalled(t, "Get")
}

// TestGetChat_InvalidLimit - тест получения чата с некорректным limit
func TestGetChat_InvalidLimit(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()

	// Отрицательный limit
	_, err := service.GetChat(ctx, 1, -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "limit cannot be negative")

	// Слишком большой limit
	_, err = service.GetChat(ctx, 1, 101)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "limit cannot exceed 100 messages")

	mockRepo.AssertNotCalled(t, "Get")
}

// TestGetChat_DefaultLimit - тест получения чата с дефолтным limit
func TestGetChat_DefaultLimit(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()
	expectedChat := &models.Chat{ID: 1, Title: "Test Chat"}

	mockRepo.On("Get", ctx, uint(1), 20).Return(expectedChat, nil)

	result, err := service.GetChat(ctx, 1, 0) // limit = 0 должен стать 20

	assert.NoError(t, err)
	assert.Equal(t, expectedChat, result)

	mockRepo.AssertExpectations(t)
}

// TestGetChat_Success - тест успешного получения чата
func TestGetChat_Success(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()
	expectedChat := &models.Chat{ID: 1, Title: "Test Chat"}

	mockRepo.On("Get", ctx, uint(1), 50).Return(expectedChat, nil)

	result, err := service.GetChat(ctx, 1, 50)

	assert.NoError(t, err)
	assert.Equal(t, expectedChat, result)

	mockRepo.AssertExpectations(t)
}

// TestGetChat_RepositoryError - тест ошибки репозитория при получении чата
func TestGetChat_RepositoryError(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()
	repoError := errors.New("chat not found")

	mockRepo.On("Get", ctx, uint(1), 20).Return((*models.Chat)(nil), repoError)

	_, err := service.GetChat(ctx, 1, 20)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get chat")
	assert.Contains(t, err.Error(), repoError.Error())

	mockRepo.AssertExpectations(t)
}

// TestDeleteChat_InvalidID - тест удаления чата с некорректным ID
func TestDeleteChat_InvalidID(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()

	err := service.DeleteChat(ctx, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat ID must be greater than 0")

	mockRepo.AssertNotCalled(t, "Delete")
}

// TestDeleteChat_Success - тест успешного удаления чата
func TestDeleteChat_Success(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()

	mockRepo.On("Delete", ctx, uint(1)).Return(nil)

	err := service.DeleteChat(ctx, 1)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestDeleteChat_RepositoryError - тест ошибки репозитория при удалении чата
func TestDeleteChat_RepositoryError(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()
	repoError := errors.New("delete failed")

	mockRepo.On("Delete", ctx, uint(1)).Return(repoError)

	err := service.DeleteChat(ctx, 1)

	assert.Error(t, err)
	assert.Equal(t, repoError, err)

	mockRepo.AssertExpectations(t)
}

// TestSendMessage_InvalidChatID - тест отправки сообщения с некорректным chat ID
func TestSendMessage_InvalidChatID(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()

	_, err := service.SendMessage(ctx, 0, "Valid message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat ID must be greater than 0")

	mockRepo.AssertNotCalled(t, "CreateMessage")
}

// TestSendMessage_EmptyText - тест отправки сообщения с пустым текстом
func TestSendMessage_EmptyText(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()

	_, err := service.SendMessage(ctx, 1, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message text cannot be empty")

	_, err = service.SendMessage(ctx, 1, "   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message text cannot be empty")

	mockRepo.AssertNotCalled(t, "CreateMessage")
}

// TestSendMessage_TooLongText - тест отправки сообщения со слишком длинным текстом
func TestSendMessage_TooLongText(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()

	// Создаем строку длиной 5001 символ
	longText := string(make([]byte, 5001))

	_, err := service.SendMessage(ctx, 1, longText)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message text cannot exceed 5000 characters")

	mockRepo.AssertNotCalled(t, "CreateMessage")
}

// TestSendMessage_Success - тест успешной отправки сообщения
func TestSendMessage_Success(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()
	inputText := "  Test message  "
	expectedText := "Test message"
	chatID := uint(1)

	expectedMessage := &models.Message{
		ID:     1,
		ChatID: chatID,
		Text:   expectedText,
	}

	mockRepo.On("CreateMessage", ctx, chatID, mock.MatchedBy(func(msg *models.Message) bool {
		return msg.ChatID == chatID && msg.Text == expectedText
	})).Return(expectedMessage, nil)

	result, err := service.SendMessage(ctx, chatID, inputText)

	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, result)
	assert.Equal(t, expectedText, result.Text) // Проверяем, что пробелы убраны

	mockRepo.AssertExpectations(t)
}

// TestSendMessage_RepositoryError - тест ошибки репозитория при отправке сообщения
func TestSendMessage_RepositoryError(t *testing.T) {
	mockRepo := new(MockChatRepository)
	service := NewChatService(mockRepo)

	ctx := context.Background()
	repoError := errors.New("create message failed")

	mockRepo.On("CreateMessage", ctx, uint(1), mock.Anything).Return((*models.Message)(nil), repoError)

	_, err := service.SendMessage(ctx, 1, "Valid message")

	assert.Error(t, err)
	assert.Equal(t, repoError, err)

	mockRepo.AssertExpectations(t)
}