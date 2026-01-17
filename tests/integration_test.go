package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chat-api/handlers"
	"chat-api/logger"
	"chat-api/models"
	"chat-api/repository"
	"chat-api/service"
	"chat-api/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TODO: Сделать скрипт который берет миграции из файлов
const migrationSQL = `
CREATE TABLE IF NOT EXISTS chats (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
`

// IntegrationTestSuite - набор интеграционных тестов
type IntegrationTestSuite struct {
	suite.Suite
	db         *gorm.DB
	sqlDB      *sql.DB
	router     *handlers.Router
	testServer *httptest.Server
}

// SetupSuite - инициализация тестовой базы данных
func (suite *IntegrationTestSuite) SetupSuite() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		utils.GetEnv("DB_HOST", "localhost"),
		utils.GetEnv("DB_PORT", "5433"),
		utils.GetEnv("DB_USER", "postgres"),
		utils.GetEnv("DB_PASSWORD", "password"),
		utils.GetEnv("DB_NAME", "chat_api_test"),
		utils.GetEnv("DB_SSLMODE", "disable"),
	)

	var err error
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		suite.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	suite.Require().NoError(err, "Failed to connect to PostgreSQL after retries")

	suite.sqlDB, err = suite.db.DB()
	suite.Require().NoError(err)

	err = suite.runMigrations()
	suite.Require().NoError(err)

	databaseLogger := logger.NewDatabaseLogger()
	repo := repository.NewRepository(suite.db, databaseLogger)
	chatService := service.NewChatService(repo)
	requestLogger := logger.NewRequestLogger()

	suite.router = handlers.New(chatService, requestLogger)

	suite.testServer = httptest.NewServer(suite.router)
}

// TearDownSuite - очистка после тестов
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.sqlDB != nil {
		suite.sqlDB.Close()
	}
	if suite.testServer != nil {
		suite.testServer.Close()
	}
}

// runMigrations - выполняет миграции для тестовой БД
func (suite *IntegrationTestSuite) runMigrations() error {
	if err := suite.db.Exec(migrationSQL).Error; err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	return nil
}

// TestCreateChat - тест создания чата
func (suite *IntegrationTestSuite) TestCreateChat() {
	reqBody := models.CreateChatRequest{Title: "Test Chat"}
	reqJSON, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", suite.testServer.URL+"/chats", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	assert.Equal(suite.T(), "application/json", resp.Header.Get("Content-Type"))

	var chat models.Chat
	err = json.NewDecoder(resp.Body).Decode(&chat)
	suite.NoError(err)
	assert.Equal(suite.T(), "Test Chat", chat.Title)
	assert.NotZero(suite.T(), chat.ID)
	assert.NotZero(suite.T(), chat.CreatedAt)
}

// TestGetChat - тест получения чата
func (suite *IntegrationTestSuite) TestGetChat() {
	// Сначала создаем чат
	chat := &models.Chat{Title: "Test Get Chat"}
	err := suite.db.Create(chat).Error
	suite.NoError(err)

	message := &models.Message{ChatID: chat.ID, Text: "Test message"}
	err = suite.db.Create(message).Error
	suite.NoError(err)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/chats/%d", suite.testServer.URL, chat.ID), nil)

	resp, err := http.DefaultClient.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var chatResponse models.ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResponse)
	suite.NoError(err)
	assert.Equal(suite.T(), chat.ID, chatResponse.ID)
	assert.Equal(suite.T(), "Test Get Chat", chatResponse.Title)
	assert.Len(suite.T(), chatResponse.Messages, 1)
	assert.Equal(suite.T(), "Test message", chatResponse.Messages[0].Text)
}

// TestSendMessage - тест отправки сообщения
func (suite *IntegrationTestSuite) TestSendMessage() {
	chat := &models.Chat{Title: "Test Message Chat"}
	err := suite.db.Create(chat).Error
	suite.NoError(err)

	reqBody := models.CreateMessageRequest{Text: "Hello from test"}
	reqJSON, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/chats/%d/messages", suite.testServer.URL, chat.ID), bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var message models.Message
	err = json.NewDecoder(resp.Body).Decode(&message)
	suite.NoError(err)
	assert.Equal(suite.T(), chat.ID, message.ChatID)
	assert.Equal(suite.T(), "Hello from test", message.Text)
	assert.NotZero(suite.T(), message.ID)
}

// TestDeleteChat - тест hard delete чата с каскадным удалением сообщений
func (suite *IntegrationTestSuite) TestDeleteChat() {
	chat := &models.Chat{Title: "Test Delete Chat"}
	err := suite.db.Create(chat).Error
	suite.NoError(err)

	message := &models.Message{ChatID: chat.ID, Text: "Test message for deletion"}
	err = suite.db.Create(message).Error
	suite.NoError(err)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/chats/%d", suite.testServer.URL, chat.ID), nil)

	resp, err := http.DefaultClient.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusNoContent, resp.StatusCode)

	var deletedChat models.Chat
	err = suite.db.First(&deletedChat, chat.ID).Error
	assert.Error(suite.T(), err, "Chat should be completely deleted")

	var deletedMessages []models.Message
	err = suite.db.Where("chat_id = ?", chat.ID).Find(&deletedMessages).Error
	suite.NoError(err)
	assert.Empty(suite.T(), deletedMessages, "Messages should be deleted due to CASCADE")
}

// TestRunSuite - запуск всех тестов
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
