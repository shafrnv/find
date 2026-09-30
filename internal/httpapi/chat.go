package httpapi

import (
	"errors"
	"net/http"

	"find/internal/store"
)

// Маршруты блока chat (верификация + переписка). Встреча — отдельный пакет meeting.

func (s *Server) claim(w http.ResponseWriter, r *http.Request) {
	account, ok := s.requireAccount(w, r, "войдите, чтобы откликнуться")
	if !ok {
		return
	}
	dialog, err := s.store.ClaimFound(r.Context(), r.PathValue("id"), account.ID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "объявление не найдено")
		return
	}
	if errors.Is(err, store.ErrBadClaim) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, dialog)
}

func (s *Server) dialogs(w http.ResponseWriter, r *http.Request) {
	account, ok := s.requireAccount(w, r, "войдите, чтобы открыть чаты")
	if !ok {
		return
	}
	list, err := s.store.ListDialogs(r.Context(), account.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось открыть чаты")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) dialog(w http.ResponseWriter, r *http.Request) {
	account, ok := s.requireAccount(w, r, "войдите, чтобы открыть чат")
	if !ok {
		return
	}
	detail, err := s.store.Dialog(r.Context(), r.PathValue("id"), account.ID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "чат не найден")
		return
	}
	if errors.Is(err, store.ErrForbidden) {
		writeError(w, http.StatusForbidden, "нет доступа к этому чату")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось открыть чат")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) sendMessage(w http.ResponseWriter, r *http.Request) {
	account, ok := s.requireAccount(w, r, "войдите, чтобы написать")
	if !ok {
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	msg, err := s.store.SendMessage(r.Context(), r.PathValue("id"), account.ID, body.Body)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "чат не найден")
		return
	}
	if errors.Is(err, store.ErrForbidden) {
		writeError(w, http.StatusForbidden, "нет доступа к этому чату")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

func (s *Server) answerDialog(w http.ResponseWriter, r *http.Request) {
	account, ok := s.requireAccount(w, r, "войдите, чтобы ответить на примету")
	if !ok {
		return
	}
	var body struct {
		Answer string `json:"answer"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	detail, err := s.store.SubmitDialogAnswer(r.Context(), r.PathValue("id"), account.ID, body.Answer)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "чат не найден")
		return
	}
	if errors.Is(err, store.ErrForbidden) {
		writeError(w, http.StatusForbidden, "нет доступа к этому чату")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) confirmDialog(w http.ResponseWriter, r *http.Request) {
	account, ok := s.requireAccount(w, r, "войдите, чтобы подтвердить примету")
	if !ok {
		return
	}
	var body struct {
		Accept bool `json:"accept"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	detail, err := s.store.ConfirmDialogAnswer(r.Context(), r.PathValue("id"), account.ID, body.Accept)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "чат не найден")
		return
	}
	if errors.Is(err, store.ErrForbidden) {
		writeError(w, http.StatusForbidden, "нет доступа к этому чату")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}
