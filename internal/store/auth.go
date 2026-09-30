package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"find/internal/domain"
)

const sessionTTL = 30 * 24 * time.Hour

var (
	ErrInvalidLogin  = errors.New("неверный логин или пароль")
	ErrUsernameTaken = errors.New("это имя уже занято")
)

type Account struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	DisplayName    string `json:"displayName"`
	Bio            string `json:"bio"`
	IdentityStatus string `json:"identityStatus"`
}

func (s *Store) Register(ctx context.Context, username, displayName, password string) (Account, error) {
	if len([]rune(password)) < 8 || len(password) > 72 {
		return Account{}, fmt.Errorf("пароль: от 8 до 72 символов")
	}
	id, err := newID()
	if err != nil {
		return Account{}, err
	}
	user, err := domain.NewUser(domain.ID(id), username, displayName, "", nil, time.Now())
	if err != nil {
		return Account{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Account{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Account{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		insert into users (id, username, display_name, bio, created_at, password_hash)
		values ($1, $2, $3, $4, $5, $6)`,
		string(user.ID), user.Username, user.DisplayName, user.Bio, user.CreatedAt, string(hash))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Account{}, ErrUsernameTaken
		}
		return Account{}, err
	}
	if _, err := tx.Exec(ctx, `
		insert into identity_verifications (user_id, status, verified_at)
		values ($1, 'none', null)`, string(user.ID)); err != nil {
		return Account{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Account{}, err
	}
	return Account{
		ID:             string(user.ID),
		Username:       user.Username,
		DisplayName:    user.DisplayName,
		Bio:            user.Bio,
		IdentityStatus: "none",
	}, nil
}

func (s *Store) Login(ctx context.Context, username, password string) (Account, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	var account Account
	var hash string
	err := s.pool.QueryRow(ctx, `
		select u.id::text, u.username, u.display_name, u.bio,
			coalesce(iv.status, 'none'), u.password_hash
		from users u
		left join identity_verifications iv on iv.user_id = u.id
		where u.username = $1`, username).
		Scan(&account.ID, &account.Username, &account.DisplayName, &account.Bio, &account.IdentityStatus, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrInvalidLogin
	}
	if err != nil {
		return Account{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return Account{}, ErrInvalidLogin
	}
	return account, nil
}

func (s *Store) CreateSession(ctx context.Context, userID string) (string, time.Time, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(raw[:])
	expires := time.Now().Add(sessionTTL)
	_, err := s.pool.Exec(ctx, `
		insert into sessions (token_hash, user_id, created_at, expires_at)
		values ($1, $2, now(), $3)`, tokenHash(token), userID, expires)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `delete from sessions where token_hash = $1`, tokenHash(token))
	return err
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
