package domain

import (
	"fmt"
	"strings"
	"time"
)

// Intent — кем говорит автор объявления.
type Intent string

const (
	IntentSeeking Intent = "seeking" // ищущий
	IntentFound   Intent = "found"   // нашедший
	IntentMarking Intent = "marking" // авторская отметка
)

func (i Intent) valid() bool {
	return i == IntentSeeking || i == IntentFound || i == IntentMarking
}

// PublicationStatus — жизнь объявления, не жизнь субъекта.
type PublicationStatus string

const (
	StatusDraft     PublicationStatus = "draft"
	StatusPublished PublicationStatus = "published"
	StatusMatched   PublicationStatus = "matched"
	StatusClosed    PublicationStatus = "closed"
	StatusHidden    PublicationStatus = "hidden"
)

func (s PublicationStatus) valid() bool {
	switch s {
	case StatusDraft, StatusPublished, StatusMatched, StatusClosed, StatusHidden:
		return true
	default:
		return false
	}
}

// Publication — пост автора о субъекте.
// Для «нашёл» автор публичен: это нашедший.
// Для «ищу» автор по умолчанию скрыт: это и есть внутренний потерявший,
// если он пишет сам. Место и время относятся к этому сообщению.
type Publication struct {
	ID               ID
	AuthorUserID     ID
	AuthorVisibility Visibility
	Intent           Intent
	SubjectKind      SubjectKind
	SubjectID        ID
	Body             string
	OccurredAt       *time.Time
	Point            *GeoPoint
	PlaceLabel       string
	PlaceID          *ID
	Status           PublicationStatus
	Photos           []PhotoRef
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewPublication(id, author ID, intent Intent, kind SubjectKind, subjectID ID, body string, occurredAt *time.Time, point *GeoPoint, placeLabel string, placeID *ID, photos []PhotoRef, now time.Time) (Publication, error) {
	if id.IsZero() || author.IsZero() || subjectID.IsZero() {
		return Publication{}, fmt.Errorf("publication: нужны id, автор и субъект")
	}
	if !intent.valid() {
		return Publication{}, fmt.Errorf("publication.intent: неизвестное намерение")
	}
	if !kind.valid() {
		return Publication{}, fmt.Errorf("publication.subject_kind: неизвестный субъект")
	}
	if (intent == IntentSeeking || intent == IntentFound) && !kind.LostFound() {
		return Publication{}, fmt.Errorf("publication.intent: место и событие отмечают, а не ищут как пропажу")
	}
	body = strings.TrimSpace(body)
	if len([]rune(body)) > 8000 {
		return Publication{}, fmt.Errorf("publication.body: длиннее 8000 символов")
	}
	placeLabel = strings.TrimSpace(placeLabel)
	if len([]rune(placeLabel)) > 300 {
		return Publication{}, fmt.Errorf("publication.place_label: длиннее 300 символов")
	}
	if placeID != nil && placeID.IsZero() {
		return Publication{}, fmt.Errorf("publication.place_id: пустой")
	}
	if (intent == IntentSeeking || intent == IntentFound) && occurredAt == nil && placeLabel == "" && point == nil && placeID == nil {
		return Publication{}, fmt.Errorf("publication: у пропажи нужно время или место")
	}
	if now.IsZero() {
		return Publication{}, fmt.Errorf("publication.created_at: обязателен")
	}
	visibility := VisibilityPublic
	if intent == IntentSeeking {
		visibility = VisibilityInternal
	}
	return Publication{
		ID:               id,
		AuthorUserID:     author,
		AuthorVisibility: visibility,
		Intent:           intent,
		SubjectKind:      kind,
		SubjectID:        subjectID,
		Body:             body,
		OccurredAt:       occurredAt,
		Point:            point,
		PlaceLabel:       placeLabel,
		PlaceID:          placeID,
		Status:           StatusDraft,
		Photos:           photos,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// RevealAuthor открывает автора объявления «ищу», если он сам этого хочет.
func (p Publication) RevealAuthor() (Publication, error) {
	if p.Intent != IntentSeeking {
		return Publication{}, fmt.Errorf("publication: открывать автора есть смысл у объявления «ищу»")
	}
	p.AuthorVisibility = VisibilityPublic
	return p, nil
}

func (p Publication) WithStatus(status PublicationStatus, now time.Time) (Publication, error) {
	if !status.valid() {
		return Publication{}, fmt.Errorf("publication.status: неизвестный статус")
	}
	if now.IsZero() {
		return Publication{}, fmt.Errorf("publication.updated_at: обязателен")
	}
	p.Status = status
	p.UpdatedAt = now
	return p, nil
}

// PublicationView — внешний пост. AuthorUserID пуст, если автор скрыт.
type PublicationView struct {
	ID           ID
	AuthorUserID *ID
	Intent       Intent
	SubjectKind  SubjectKind
	SubjectID    ID
	Body         string
	OccurredAt   *time.Time
	Point        *GeoPoint
	PlaceLabel   string
	PlaceID      *ID
	Status       PublicationStatus
	PhotoIDs     []ID
	CreatedAt    time.Time
}

func (p Publication) Public() PublicationView {
	view := PublicationView{
		ID:          p.ID,
		Intent:      p.Intent,
		SubjectKind: p.SubjectKind,
		SubjectID:   p.SubjectID,
		Body:        p.Body,
		OccurredAt:  p.OccurredAt,
		Point:       p.Point,
		PlaceLabel:  p.PlaceLabel,
		PlaceID:     p.PlaceID,
		Status:      p.Status,
		PhotoIDs:    publicPhotoIDs(p.Photos),
		CreatedAt:   p.CreatedAt,
	}
	if p.AuthorVisibility == VisibilityPublic {
		author := p.AuthorUserID
		view.AuthorUserID = &author
	}
	return view
}
