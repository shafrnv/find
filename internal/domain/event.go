package domain

import (
	"fmt"
	"strings"
	"time"
)

// Event — событие с автором, временем и привязкой к месту или адресу.
// Поля намеренно короткие: расписание, билеты и участники сюда пока не входят.
type Event struct {
	ID            ID
	CategoryID    ID
	SubcategoryID *ID
	Name          string
	Description   string
	PlaceID       *ID
	Address       string
	Point         *GeoPoint
	StartsAt      time.Time
	EndsAt        *time.Time
	Tags          []string
	Photos        []PhotoRef
	CreatedBy     ID
	CreatedAt     time.Time
}

func NewEvent(id, categoryID ID, subcategoryID *ID, name, description string, placeID *ID, address string, point *GeoPoint, startsAt time.Time, endsAt *time.Time, tags []string, photos []PhotoRef, createdBy ID, now time.Time) (Event, error) {
	if id.IsZero() || categoryID.IsZero() || createdBy.IsZero() {
		return Event{}, fmt.Errorf("event: нужны id, категория и автор")
	}
	if subcategoryID != nil && (subcategoryID.IsZero() || *subcategoryID == categoryID) {
		return Event{}, fmt.Errorf("event.subcategory_id: некорректная подкатегория")
	}
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 160 {
		return Event{}, fmt.Errorf("event.name: от 1 до 160 символов")
	}
	description = strings.TrimSpace(description)
	if len([]rune(description)) > 8000 {
		return Event{}, fmt.Errorf("event.description: длиннее 8000 символов")
	}
	if placeID != nil && placeID.IsZero() {
		return Event{}, fmt.Errorf("event.place_id: пустой")
	}
	address = strings.TrimSpace(address)
	if placeID == nil && address == "" && point == nil {
		return Event{}, fmt.Errorf("event: нужно место, адрес или точка")
	}
	if len([]rune(address)) > 300 {
		return Event{}, fmt.Errorf("event.address: длиннее 300 символов")
	}
	if startsAt.IsZero() {
		return Event{}, fmt.Errorf("event.starts_at: обязателен")
	}
	if endsAt != nil && !endsAt.After(startsAt) {
		return Event{}, fmt.Errorf("event.ends_at: должен быть позже начала")
	}
	normTags, err := normalizeTags(tags)
	if err != nil {
		return Event{}, err
	}
	if now.IsZero() {
		return Event{}, fmt.Errorf("event.created_at: обязателен")
	}
	return Event{
		ID:            id,
		CategoryID:    categoryID,
		SubcategoryID: subcategoryID,
		Name:          name,
		Description:   description,
		PlaceID:       placeID,
		Address:       address,
		Point:         point,
		StartsAt:      startsAt,
		EndsAt:        endsAt,
		Tags:          normTags,
		Photos:        photos,
		CreatedBy:     createdBy,
		CreatedAt:     now,
	}, nil
}

func (e Event) Kind() SubjectKind { return SubjectEvent }
func (e Event) GetID() ID         { return e.ID }
