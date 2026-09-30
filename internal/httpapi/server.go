package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"find/internal/store"
)

const sessionCookie = "find_session"

type Server struct {
	store *store.Store
}

func New(db *store.Store) *Server {
	return &Server{store: db}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/entries", s.entries)
	mux.HandleFunc("GET /api/me", s.me)
	mux.HandleFunc("PATCH /api/me", s.updateMe)
	mux.HandleFunc("POST /api/register", s.register)
	mux.HandleFunc("POST /api/login", s.login)
	mux.HandleFunc("POST /api/logout", s.logout)
	mux.HandleFunc("POST /api/found", s.found)
	mux.HandleFunc("POST /api/entries/{id}/claim", s.claim)
	mux.HandleFunc("GET /api/dialogs", s.dialogs)
	mux.HandleFunc("GET /api/dialogs/{id}", s.dialog)
	mux.HandleFunc("POST /api/dialogs/{id}/messages", s.sendMessage)
	mux.HandleFunc("POST /api/dialogs/{id}/answer", s.answerDialog)
	mux.HandleFunc("POST /api/dialogs/{id}/confirm", s.confirmDialog)
	mux.HandleFunc("GET /api/users/{username}", s.profile)
	mux.HandleFunc("GET /api/users/{username}/entries", s.userEntries)
	return mux
}

func (s *Server) entries(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.Entries(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось прочитать объявления")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	account, ok := s.requireAccount(w, r, "войдите в аккаунт")
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, account)
}

func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	account, ok := s.requireAccount(w, r, "войдите в аккаунт")
	if !ok {
		return
	}
	var body struct {
		DisplayName string `json:"displayName"`
		Bio         string `json:"bio"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	updated, err := s.store.UpdateProfile(r.Context(), account.ID, body.DisplayName, body.Bio)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username    string `json:"username"`
		DisplayName string `json:"displayName"`
		Password    string `json:"password"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	account, err := s.store.Register(r.Context(), body.Username, body.DisplayName, body.Password)
	if errors.Is(err, store.ErrUsernameTaken) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !s.startSession(w, account) {
		return
	}
	writeJSON(w, http.StatusCreated, account)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	account, err := s.store.Login(r.Context(), body.Username, body.Password)
	if errors.Is(err, store.ErrInvalidLogin) {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось войти")
		return
	}
	if !s.startSession(w, account) {
		return
	}
	writeJSON(w, http.StatusOK, account)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		_ = s.store.DeleteSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) found(w http.ResponseWriter, r *http.Request) {
	account, ok := s.requireAccount(w, r, "войдите, чтобы оставить объявление")
	if !ok {
		return
	}
	var body struct {
		Kind        string  `json:"kind"`
		Subcategory string  `json:"subcategory"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Place       string  `json:"place"`
		At          string  `json:"at"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Sex         string  `json:"sex"`
		Trait       string  `json:"trait"`
		SecretTrait string  `json:"secretTrait"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	at, err := time.ParseInLocation("2006-01-02T15:04", body.At, time.Local)
	if err != nil {
		at, err = time.Parse(time.RFC3339, body.At)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "время находки обязательно")
		return
	}
	entry, err := s.store.CreateFound(r.Context(), store.FoundInput{
		AuthorID:    account.ID,
		Kind:        body.Kind,
		Subcategory: body.Subcategory,
		Title:       body.Title,
		Description: body.Description,
		Place:       body.Place,
		At:          at,
		Lat:         body.Lat,
		Lon:         body.Lon,
		Sex:         body.Sex,
		Trait:       body.Trait,
		SecretTrait: body.SecretTrait,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

func (s *Server) profile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	viewerID := ""
	if account, ok := s.optionalAccount(r); ok {
		viewerID = account.ID
	}
	profile, err := s.store.ProfileByUsername(r.Context(), username, viewerID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "не удалось открыть профиль")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) userEntries(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	viewerID := ""
	if account, ok := s.optionalAccount(r); ok {
		viewerID = account.ID
	}
	list, err := s.store.EntriesByAuthor(r.Context(), username, viewerID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "не удалось открыть ленту")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) requireAccount(w http.ResponseWriter, r *http.Request, message string) (store.Account, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		writeError(w, http.StatusUnauthorized, message)
		return store.Account{}, false
	}
	account, err := s.store.AccountBySession(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, message)
		return store.Account{}, false
	}
	return account, true
}

func (s *Server) optionalAccount(r *http.Request) (store.Account, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return store.Account{}, false
	}
	account, err := s.store.AccountBySession(r.Context(), cookie.Value)
	if err != nil {
		return store.Account{}, false
	}
	return account, true
}

func (s *Server) startSession(w http.ResponseWriter, account store.Account) bool {
	token, expires, err := s.store.CreateSession(context.Background(), account.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось открыть сессию")
		return false
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return true
}

func readJSON(w http.ResponseWriter, r *http.Request, dest any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный запрос")
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
