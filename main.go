package main

import (
	"net/http"
	"os"

	"chat-api/database"
	"chat-api/handlers"
	"chat-api/logger"
	"chat-api/repository"
	"chat-api/service"
	"chat-api/utils"
)

func main() {
	log := logger.CreateBaseLogger("logs.log")
	log.LogInfo("Application starting", "version=1.0.0")
	db, err := database.NewDB()

	if err != nil {
		log.LogError("Start database:", err)
		os.Exit(1)
	}

	defer db.Close()

	requestLogger := logger.NewRequestLogger()
	databaseLogger := logger.NewDatabaseLogger()

	repo := repository.NewRepository(db.DB, databaseLogger)

	chatService := service.NewChatService(repo)

	router := handlers.New(chatService, requestLogger)

	port := utils.GetEnv("PORT", "8080")
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.LogError("Start http server:", err)
		os.Exit(1)
	}
}
