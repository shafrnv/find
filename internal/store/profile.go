package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Profile — публичная карточка пользователя и счётчик видимых объявлений.
type Profile struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	DisplayName    string    `json:"displayName"`
	Bio            string    `json:"bio"`
	IdentityStatus string    `json:"identityStatus"`
	CreatedAt      time.Time `json:"createdAt"`
	EntryCount     int       `json:"entryCount"`
	Mine           bool      `json:"mine"`
}

func (s *Store) AccountBySession(ctx context.Context, token string) (Account, error) {
	if token == "" {
		return Account{}, ErrInvalidLogin
	}
	var account Account
	err := s.pool.QueryRow(ctx, `
		select u.id::text, u.username, u.display_name, u.bio,
			coalesce(iv.status, 'none')
		from sessions s
		join users u on u.id = s.user_id
		left join identity_verifications iv on iv.user_id = u.id
		where s.token_hash = $1 and s.expires_at > now()`, tokenHash(token)).
		Scan(&account.ID, &account.Username, &account.DisplayName, &account.Bio, &account.IdentityStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrInvalidLogin
	}
	return account, err
}

func (s *Store) ProfileByUsername(ctx context.Context, username, viewerID string) (Profile, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	var profile Profile
	err := s.pool.QueryRow(ctx, `
		select u.id::text, u.username, u.display_name, u.bio,
			coalesce(iv.status, 'none'), u.created_at
		from users u
		left join identity_verifications iv on iv.user_id = u.id
		where u.username = $1`, username).
		Scan(&profile.ID, &profile.Username, &profile.DisplayName, &profile.Bio, &profile.IdentityStatus, &profile.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, fmt.Errorf("пользователь не найден")
	}
	if err != nil {
		return Profile{}, err
	}
	profile.Mine = viewerID != "" && viewerID == profile.ID
	feed, err := s.loadEntries(ctx, entryFilter{AuthorID: profile.ID, ViewerID: viewerID})
	if err != nil {
		return Profile{}, err
	}
	profile.EntryCount = len(feed)
	return profile, nil
}

func (s *Store) UpdateProfile(ctx context.Context, userID, displayName, bio string) (Account, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || len([]rune(displayName)) > 80 {
		return Account{}, fmt.Errorf("имя: от 1 до 80 символов")
	}
	bio = strings.TrimSpace(bio)
	if len([]rune(bio)) > 500 {
		return Account{}, fmt.Errorf("о себе: длиннее 500 символов")
	}
	_, err := s.pool.Exec(ctx, `
		update users set display_name = $2, bio = $3 where id = $1`, userID, displayName, bio)
	if err != nil {
		return Account{}, err
	}
	return s.accountByID(ctx, userID)
}

func (s *Store) accountByID(ctx context.Context, id string) (Account, error) {
	var account Account
	err := s.pool.QueryRow(ctx, `
		select u.id::text, u.username, u.display_name, u.bio,
			coalesce(iv.status, 'none')
		from users u
		left join identity_verifications iv on iv.user_id = u.id
		where u.id = $1`, id).
		Scan(&account.ID, &account.Username, &account.DisplayName, &account.Bio, &account.IdentityStatus)
	return account, err
}

func (s *Store) EntriesByAuthor(ctx context.Context, username, viewerID string) ([]Entry, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	var authorID string
	err := s.pool.QueryRow(ctx, `select id::text from users where username = $1`, username).Scan(&authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("пользователь не найден")
	}
	if err != nil {
		return nil, err
	}
	return s.loadEntries(ctx, entryFilter{AuthorID: authorID, ViewerID: viewerID})
}
