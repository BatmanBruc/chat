package handlers

import (
	"net/http"
)

type RouteDefinition struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

type IChatHandler interface {
	HealthCheck(w http.ResponseWriter, r *http.Request)
	CreateChat(w http.ResponseWriter, r *http.Request)
	SendMessage(w http.ResponseWriter, r *http.Request)
	GetMessages(w http.ResponseWriter, r *http.Request)
	DeleteChat(w http.ResponseWriter, r *http.Request)
}

func RegisterChatRoutes(h IChatHandler) []RouteDefinition {
	return []RouteDefinition{
		{
			Method:  "GET",
			Path:    "/health",
			Handler: h.HealthCheck,
		},
		{
			Method:  "POST",
			Path:    "/chats",
			Handler: h.CreateChat,
		},
		{
			Method:  "POST",
			Path:    "/chats/{id}/messages",
			Handler: h.SendMessage,
		},
		{
			Method:  "GET",
			Path:    "/chats/{id}",
			Handler: h.GetMessages,
		},
		{
			Method:  "DELETE",
			Path:    "/chats/{id}",
			Handler: h.DeleteChat,
		},
	}
}
