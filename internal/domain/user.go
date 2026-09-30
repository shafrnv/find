package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// User — аккаунт. Пропавший человек сюда не входит: это отдельный субъект Person.
// Подтверждение личности хранится отдельно и наружу отдаёт только статус.
type User struct {
	ID            ID
	Username      string
	DisplayName   string
	Bio           string
	AvatarMediaID *ID
	CreatedAt     time.Time
}

// IdentityStatus — факт проверки, без паспортных данных.
type IdentityStatus string

const (
	IdentityNone     IdentityStatus = "none"
	IdentityPending  IdentityStatus = "pending"
	IdentityVerified IdentityStatus = "verified"
)

type IdentityVerification struct {
	UserID     ID
	Status     IdentityStatus
	VerifiedAt *time.Time
}

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)

func NewUser(id ID, username, displayName, bio string, avatar *ID, now time.Time) (User, error) {
	if id.IsZero() {
		return User{}, fmt.Errorf("user.id: обязателен")
	}
	username = strings.TrimSpace(strings.ToLower(username))
	if !usernamePattern.MatchString(username) {
		return User{}, fmt.Errorf("user.username: 3–32 символа, латиница, цифры и _")
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || len([]rune(displayName)) > 80 {
		return User{}, fmt.Errorf("user.display_name: от 1 до 80 символов")
	}
	bio = strings.TrimSpace(bio)
	if len([]rune(bio)) > 500 {
		return User{}, fmt.Errorf("user.bio: длиннее 500 символов")
	}
	if now.IsZero() {
		return User{}, fmt.Errorf("user.created_at: обязателен")
	}
	return User{
		ID:            id,
		Username:      username,
		DisplayName:   displayName,
		Bio:           bio,
		AvatarMediaID: avatar,
		CreatedAt:     now,
	}, nil
}

func NewIdentityVerification(userID ID, status IdentityStatus, verifiedAt *time.Time) (IdentityVerification, error) {
	if userID.IsZero() {
		return IdentityVerification{}, fmt.Errorf("identity.user_id: обязателен")
	}
	switch status {
	case IdentityNone, IdentityPending, IdentityVerified:
	default:
		return IdentityVerification{}, fmt.Errorf("identity.status: неизвестный статус")
	}
	if status == IdentityVerified && verifiedAt == nil {
		return IdentityVerification{}, fmt.Errorf("identity.verified_at: нужен для verified")
	}
	if status != IdentityVerified && verifiedAt != nil {
		return IdentityVerification{}, fmt.Errorf("identity.verified_at: только для verified")
	}
	return IdentityVerification{UserID: userID, Status: status, VerifiedAt: verifiedAt}, nil
}

// PublicUser — то, что видно в профиле автора канала.
type PublicUser struct {
	ID             ID
	Username       string
	DisplayName    string
	Bio            string
	AvatarMediaID  *ID
	IdentityStatus IdentityStatus
}

func (u User) Public(identity IdentityStatus) PublicUser {
	if identity == "" {
		identity = IdentityNone
	}
	return PublicUser{
		ID:             u.ID,
		Username:       u.Username,
		DisplayName:    u.DisplayName,
		Bio:            u.Bio,
		AvatarMediaID:  u.AvatarMediaID,
		IdentityStatus: identity,
	}
}
