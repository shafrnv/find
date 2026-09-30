package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ExternalSource — площадка, с которой снята карточка места.
type ExternalSource string

const (
	SourceYandex ExternalSource = "yandex"
	SourceTwoGIS ExternalSource = "2gis"
	SourceGoogle ExternalSource = "google"
	SourceOSM    ExternalSource = "osm"
)

func (s ExternalSource) valid() bool {
	switch s {
	case SourceYandex, SourceTwoGIS, SourceGoogle, SourceOSM:
		return true
	default:
		return false
	}
}

// ExternalRef — снимок сведений о месте с внешней карты.
// Snapshot хранит то, что отдала площадка, без приведения к нашей схеме.
type ExternalRef struct {
	ID         ID
	Source     ExternalSource
	ExternalID string
	URL        string
	Snapshot   json.RawMessage
}

func NewExternalRef(id ID, source ExternalSource, externalID, rawURL string, snapshot json.RawMessage) (ExternalRef, error) {
	if id.IsZero() {
		return ExternalRef{}, fmt.Errorf("external_ref.id: обязателен")
	}
	if !source.valid() {
		return ExternalRef{}, fmt.Errorf("external_ref.source: неизвестная площадка")
	}
	externalID = strings.TrimSpace(externalID)
	if externalID == "" || len(externalID) > 200 {
		return ExternalRef{}, fmt.Errorf("external_ref.external_id: обязателен")
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL != "" && !strings.HasPrefix(rawURL, "https://") {
		return ExternalRef{}, fmt.Errorf("external_ref.url: нужен https")
	}
	if len(snapshot) > 0 && !json.Valid(snapshot) {
		return ExternalRef{}, fmt.Errorf("external_ref.snapshot: это не json")
	}
	return ExternalRef{ID: id, Source: source, ExternalID: externalID, URL: rawURL, Snapshot: snapshot}, nil
}

// Place — место, которое можно найти и за которым можно следить.
// CreatedBy — кто завёл карточку. Чужие отметки этого же места
// приходят отдельными объявлениями с намерением marking.
type Place struct {
	ID            ID
	CategoryID    ID
	SubcategoryID *ID
	Name          string
	Description   string
	Address       string
	Point         *GeoPoint
	Tags          []string
	ExternalRefs  []ExternalRef
	Photos        []PhotoRef
	CreatedBy     ID
	CreatedAt     time.Time
}

func NewPlace(id, categoryID ID, subcategoryID *ID, name, description, address string, point *GeoPoint, tags []string, refs []ExternalRef, photos []PhotoRef, createdBy ID, now time.Time) (Place, error) {
	if id.IsZero() || categoryID.IsZero() || createdBy.IsZero() {
		return Place{}, fmt.Errorf("place: нужны id, категория и автор карточки")
	}
	if subcategoryID != nil && (subcategoryID.IsZero() || *subcategoryID == categoryID) {
		return Place{}, fmt.Errorf("place.subcategory_id: некорректная подкатегория")
	}
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 160 {
		return Place{}, fmt.Errorf("place.name: от 1 до 160 символов")
	}
	description = strings.TrimSpace(description)
	if len([]rune(description)) > 8000 {
		return Place{}, fmt.Errorf("place.description: длиннее 8000 символов")
	}
	address = strings.TrimSpace(address)
	if address == "" && point == nil {
		return Place{}, fmt.Errorf("place: нужен адрес или точка на карте")
	}
	if len([]rune(address)) > 300 {
		return Place{}, fmt.Errorf("place.address: длиннее 300 символов")
	}
	normTags, err := normalizeTags(tags)
	if err != nil {
		return Place{}, err
	}
	if now.IsZero() {
		return Place{}, fmt.Errorf("place.created_at: обязателен")
	}
	return Place{
		ID:            id,
		CategoryID:    categoryID,
		SubcategoryID: subcategoryID,
		Name:          name,
		Description:   description,
		Address:       address,
		Point:         point,
		Tags:          normTags,
		ExternalRefs:  refs,
		Photos:        photos,
		CreatedBy:     createdBy,
		CreatedAt:     now,
	}, nil
}

func (p Place) Kind() SubjectKind { return SubjectPlace }
func (p Place) GetID() ID         { return p.ID }

func normalizeTags(tags []string) ([]string, error) {
	if len(tags) > 20 {
		return nil, fmt.Errorf("tags: не больше 20")
	}
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" {
			continue
		}
		if len([]rune(tag)) > 40 || strings.ContainsAny(tag, " \n\t") {
			return nil, fmt.Errorf("tags: метка короче 40 символов и без пробелов")
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out, nil
}
