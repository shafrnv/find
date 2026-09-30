package domain

import (
	"fmt"
	"time"
)

// MatchStatus — стадия сведения объявления «ищу» и объявления «нашёл».
type MatchStatus string

const (
	MatchProposed  MatchStatus = "proposed"
	MatchConfirmed MatchStatus = "confirmed"
	MatchRejected  MatchStatus = "rejected"
)

// Match связывает две публикации одного субъекта.
// Подтверждение встречи, выплата и скрытая примета живут следующими слоями
// и в эту сущность не входят.
type Match struct {
	ID                   ID
	SeekingPublicationID ID
	FoundPublicationID   ID
	Status               MatchStatus
	CreatedAt            time.Time
}

func NewMatch(id ID, seeking, found Publication, now time.Time) (Match, error) {
	if id.IsZero() {
		return Match{}, fmt.Errorf("match.id: обязателен")
	}
	if seeking.Intent != IntentSeeking || found.Intent != IntentFound {
		return Match{}, fmt.Errorf("match: нужны объявления «ищу» и «нашёл»")
	}
	if seeking.SubjectKind != found.SubjectKind || seeking.SubjectID != found.SubjectID {
		return Match{}, fmt.Errorf("match: объявления о разных субъектах")
	}
	if seeking.ID == found.ID {
		return Match{}, fmt.Errorf("match: объявление нельзя свести с самим собой")
	}
	if now.IsZero() {
		return Match{}, fmt.Errorf("match.created_at: обязателен")
	}
	return Match{
		ID:                   id,
		SeekingPublicationID: seeking.ID,
		FoundPublicationID:   found.ID,
		Status:               MatchProposed,
		CreatedAt:            now,
	}, nil
}
