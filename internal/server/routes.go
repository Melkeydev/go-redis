package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-auth-testing/internal/auth"
	_ "go-auth-testing/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", s.helloWorldHandler)

	r.Get("/auth/{provider}/callback", s.getAuthProviderCallback)
	r.Get("/logout/{provider}", s.logoutProvider)
	r.Get("/auth/{provider}", s.beginAuthProviderCallback)
	r.Get("/debug/providers", s.debugProvidersHandler)

	return r
}

func (s *Server) helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	w.Write(jsonResp)
}

func (s *Server) getAuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Printf("This is the error from completing user auth %v", err)
		return
	}

	sessionID := uuid.New().String()
	expirationDuration := time.Duration(auth.MaxAge) * time.Second

	session := auth.Session{
		UserID: user.UserID,
	}

	sessionJsonData, err := json.Marshal(session)
	if err != nil {
		fmt.Printf("Error marshaling session: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = redisClient.Set(ctx, sessionID, sessionJsonData, expirationDuration).Err()
	if err != nil {
		fmt.Printf("Error setting session in Redis: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   auth.MaxAge,
		HttpOnly: true,
		Secure:   auth.IsProd,
	})

	http.Redirect(w, r, "http://localhost:5173", http.StatusFound)
}

func (s *Server) beginAuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))
	gothic.BeginAuthHandler(w, r)
}

// TODO: delete session after explicit logout
func (s *Server) logoutProvider(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))
	gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) debugProvidersHandler(w http.ResponseWriter, r *http.Request) {
	providers := goth.GetProviders()
	for name, _ := range providers {
		fmt.Println("Configured provider:", name)
	}
}

func (s *Server) validateSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Session cookie not found", http.StatusUnauthorized)
		return
	}

	_, err = redisClient.Get(ctx, cookie.Value).Result()
	if err != nil {
		fmt.Printf("Error retrieving session from Redis: %v\n", err)
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}
}
