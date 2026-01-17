package handlers

import (
	"chat-api/models"
	"context"
	"net/http"
	"strings"
)

type Logger interface {
	Log(method, path, remoteAddr string, statusCode int, durationMs float64)
}

type ChatService interface {
	CreateChat(ctx context.Context, title string) (*models.Chat, error)
	GetChat(ctx context.Context, id uint, limit int) (*models.Chat, error)
	DeleteChat(ctx context.Context, id uint) error
	SendMessage(ctx context.Context, chatID uint, text string) (*models.Message, error)
}

type Router struct {
	routes map[string]http.HandlerFunc
	logger Logger
}

func New(service ChatService, logger Logger) *Router {
	router := &Router{
		routes: make(map[string]http.HandlerFunc),
		logger: logger,
	}

	handler := NewChatHandler(service)
	routes := RegisterChatRoutes(handler)

	for _, route := range routes {
		key := route.Method + " " + route.Path
		router.routes[key] = route.Handler
	}

	return router
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := rt.findHandler(r)
	if handler != nil {
		handler = rt.withMiddleware(handler)
		handler(w, r)
		return
	}

	http.NotFound(w, r)
}

func (rt *Router) findHandler(r *http.Request) http.HandlerFunc {
	method := r.Method
	path := r.URL.Path

	key := method + " " + path
	if handler, exists := rt.routes[key]; exists {
		return handler
	}

	for route, handler := range rt.routes {
		if rt.matchRoute(route, method, path) {
			return handler
		}
	}

	return nil
}

func (rt *Router) matchRoute(route, method, path string) bool {
	routeParts := strings.Split(route, " ")
	if len(routeParts) != 2 || routeParts[0] != method {
		return false
	}

	routePath := routeParts[1]

	if routePath == path {
		return true
	}

	return rt.matchDynamicRoute(routePath, path)
}

func (rt *Router) matchDynamicRoute(routePath, actualPath string) bool {
	routeSegments := strings.Split(routePath, "/")
	actualSegments := strings.Split(actualPath, "/")

	if len(routeSegments) != len(actualSegments) {
		return false
	}

	for i, routeSeg := range routeSegments {
		actualSeg := actualSegments[i]

		if strings.HasPrefix(routeSeg, "{") && strings.HasSuffix(routeSeg, "}") {
			continue
		}

		if routeSeg != actualSeg {
			return false
		}
	}

	return true
}

func (rt *Router) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return LoggingMiddleware(next, rt.logger)
}
