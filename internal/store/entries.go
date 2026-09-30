package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"find/internal/domain"
)

// Entry — карточка на карте: объявление вместе с субъектом и автором.
type Entry struct {
	ID             string      `json:"id"`
	Kind           string      `json:"kind"`
	SubcategoryID  string      `json:"subcategoryId"`
	Intent         string      `json:"intent"`
	Title          string      `json:"title"`
	Author         string      `json:"author"`
	AuthorUserID   *string     `json:"authorUserId"`
	AuthorUsername *string     `json:"authorUsername"`
	At             time.Time   `json:"at"`
	Coordinates    [2]float64  `json:"coordinates"`
	Place          string      `json:"place"`
	Description    string      `json:"description"`
	Traits         []TraitLine `json:"traits"`
	PhotoCount     int         `json:"photoCount"`
}

type entryFilter struct {
	OnlyID   string
	AuthorID string
	ViewerID string
}

type TraitLine struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type FoundInput struct {
	AuthorID    string
	Kind        string
	Subcategory string
	Title       string
	Description string
	Place       string
	At          time.Time
	Lat         float64
	Lon         float64
	Sex         string
	Trait       string // публичная примета (видна на карточке)
	SecretTrait string // скрытая «отличительная особенность» для верификации чата
}

func (s *Store) Entries(ctx context.Context) ([]Entry, error) {
	return s.loadEntries(ctx, entryFilter{})
}

func (s *Store) CreateFound(ctx context.Context, in FoundInput) (Entry, error) {
	kind := domain.SubjectKind(in.Kind)
	if !kind.LostFound() {
		return Entry{}, fmt.Errorf("находка бывает у предмета, человека или животного")
	}
	point, err := domain.NewGeoPoint(in.Lat, in.Lon)
	if err != nil {
		return Entry{}, err
	}
	if in.At.IsZero() {
		return Entry{}, fmt.Errorf("время находки обязательно")
	}
	sex := domain.Sex(in.Sex)
	if sex == "" {
		sex = domain.SexUnknown
	}
	switch sex {
	case domain.SexFemale, domain.SexMale, domain.SexUnknown:
	default:
		return Entry{}, fmt.Errorf("пол: female, male или unknown")
	}
	title := strings.TrimSpace(in.Title)
	description := strings.TrimSpace(in.Description)
	place := strings.TrimSpace(in.Place)
	if place == "" {
		return Entry{}, fmt.Errorf("место обязательно")
	}

	var traits []domain.Trait
	if strings.TrimSpace(in.Trait) != "" {
		trait, err := domain.NewTrait(domain.TraitDistinctive, in.Trait, domain.VisibilityPublic)
		if err != nil {
			return Entry{}, err
		}
		traits = append(traits, trait)
	}
	if strings.TrimSpace(in.SecretTrait) != "" {
		trait, err := domain.SecretDistinctive(in.SecretTrait)
		if err != nil {
			return Entry{}, err
		}
		traits = append(traits, trait)
	}
	if sex != domain.SexUnknown && (kind == domain.SubjectPerson || kind == domain.SubjectAnimal) {
		trait, err := domain.NewTrait("pol", sexLabel(string(kind), string(sex)), domain.VisibilityPublic)
		if err != nil {
			return Entry{}, err
		}
		traits = append(traits, trait)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Entry{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rootID, subID, err := subcategory(ctx, tx, string(kind), in.Subcategory)
	if err != nil {
		return Entry{}, err
	}
	subjectID, err := newID()
	if err != nil {
		return Entry{}, err
	}
	now := time.Now()
	switch kind {
	case domain.SubjectItem:
		item, err := domain.NewItem(domain.ID(subjectID), title, domain.ID(rootID), idPtr(subID), description, nil, nil, nil, now)
		if err != nil {
			return Entry{}, err
		}
		_, err = tx.Exec(ctx, `
			insert into items (id, name, category_id, subcategory_id, description, owner_user_id, created_at)
			values ($1, $2, $3, $4, $5, null, $6)`,
			string(item.ID), item.Name, rootID, subID, item.Description, item.CreatedAt)
		if err != nil {
			return Entry{}, err
		}
	case domain.SubjectPerson:
		given, family := splitName(title)
		person, err := domain.NewPerson(domain.ID(subjectID), given, family, "", sex, nil, nil, now)
		if err != nil {
			return Entry{}, err
		}
		_, err = tx.Exec(ctx, `
			insert into persons (id, given_name, family_name, patronymic, sex, subcategory_id, created_at)
			values ($1, $2, $3, $4, $5, $6, $7)`,
			string(person.ID), person.GivenName, person.FamilyName, person.Patronymic, string(person.Sex), subID, person.CreatedAt)
		if err != nil {
			return Entry{}, err
		}
	case domain.SubjectAnimal:
		animal, err := domain.NewAnimal(domain.ID(subjectID), domain.ID(rootID), idPtr(subID), "", sex, title, nil, nil, nil, now)
		if err != nil {
			return Entry{}, err
		}
		_, err = tx.Exec(ctx, `
			insert into animals (id, category_id, breed_category_id, breed_text, sex, nickname, owner_user_id, created_at)
			values ($1, $2, $3, $4, $5, $6, null, $7)`,
			string(animal.ID), rootID, subID, animal.BreedText, string(animal.Sex), animal.Nickname, animal.CreatedAt)
		if err != nil {
			return Entry{}, err
		}
	}

	for i, trait := range traits {
		traitID, err := newID()
		if err != nil {
			return Entry{}, err
		}
		if _, err := tx.Exec(ctx, `
			insert into traits (id, subject_kind, subject_id, key, value, visibility, sort_order)
			values ($1, $2, $3, $4, $5, $6, $7)`,
			traitID, string(kind), subjectID, trait.Key, trait.Value, string(trait.Visibility), i); err != nil {
			return Entry{}, err
		}
	}

	pubID, err := newID()
	if err != nil {
		return Entry{}, err
	}
	occurred := in.At
	publication, err := domain.NewPublication(
		domain.ID(pubID), domain.ID(in.AuthorID), domain.IntentFound, kind, domain.ID(subjectID),
		description, &occurred, &point, place, nil, nil, now,
	)
	if err != nil {
		return Entry{}, err
	}
	publication, err = publication.WithStatus(domain.StatusPublished, now)
	if err != nil {
		return Entry{}, err
	}
	if _, err := tx.Exec(ctx, `
		insert into publications (
			id, author_user_id, author_visibility, intent, subject_kind, subject_id,
			body, occurred_at, lat, lon, place_label, status, created_at, updated_at
		) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		string(publication.ID), string(publication.AuthorUserID), string(publication.AuthorVisibility),
		string(publication.Intent), string(publication.SubjectKind), subjectID,
		publication.Body, publication.OccurredAt, point.Lat, point.Lon, publication.PlaceLabel,
		string(publication.Status), publication.CreatedAt, publication.UpdatedAt,
	); err != nil {
		return Entry{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Entry{}, err
	}
	list, err := s.loadEntries(ctx, entryFilter{OnlyID: string(publication.ID)})
	if err != nil {
		return Entry{}, err
	}
	if len(list) != 1 {
		return Entry{}, fmt.Errorf("объявление не прочиталось после записи")
	}
	return list[0], nil
}

func subcategory(ctx context.Context, tx pgx.Tx, kind, slug string) (rootID, subID string, err error) {
	err = tx.QueryRow(ctx, `
		select p.id::text, c.id::text
		from categories c
		join categories p on p.id = c.parent_id
		where c.slug = $1 and c.scope = $2 and p.scope = $2 and p.parent_id is null`,
		strings.TrimSpace(slug), kind).Scan(&rootID, &subID)
	if errorsIsNoRows(err) {
		return "", "", fmt.Errorf("подкатегория не относится к этой категории")
	}
	return rootID, subID, err
}

func (s *Store) loadEntries(ctx context.Context, filter entryFilter) ([]Entry, error) {
	reveal := filter.AuthorID != "" && filter.AuthorID == filter.ViewerID
	query := entrySQL
	args := []any{reveal}
	n := 2
	if filter.OnlyID != "" {
		query += fmt.Sprintf(` and pub.id = $%d`, n)
		args = append(args, filter.OnlyID)
		n++
	}
	if filter.AuthorID != "" {
		query += fmt.Sprintf(` and pub.author_user_id = $%d`, n)
		args = append(args, filter.AuthorID)
		n++
		if filter.ViewerID == "" {
			query += ` and pub.author_visibility = 'public'`
		} else {
			query += fmt.Sprintf(` and (pub.author_visibility = 'public' or pub.author_user_id = $%d)`, n)
			args = append(args, filter.ViewerID)
			n++
		}
	}
	query += ` order by coalesce(pub.occurred_at, ev.starts_at, pub.created_at) desc`
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	subjects := map[string][]int{}
	for rows.Next() {
		var entry Entry
		var subjectID string
		var lat, lon *float64
		var authorID, authorUsername, authorName *string
		if err := rows.Scan(
			&entry.ID, &entry.Kind, &entry.Intent, &subjectID,
			&authorName, &authorID, &authorUsername, &entry.At,
			&lon, &lat, &entry.Place, &entry.Description, &entry.Title, &entry.SubcategoryID,
		); err != nil {
			return nil, err
		}
		if lat == nil || lon == nil || entry.Title == "" {
			continue
		}
		if authorName != nil {
			entry.Author = *authorName
		}
		entry.AuthorUserID = authorID
		entry.AuthorUsername = authorUsername
		entry.Coordinates = [2]float64{*lon, *lat}
		entry.Traits = []TraitLine{}
		key := entry.Kind + "/" + subjectID
		subjects[key] = append(subjects[key], len(entries))
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.attachTraits(ctx, entries, subjects); err != nil {
		return nil, err
	}
	if err := s.attachTags(ctx, entries, subjects); err != nil {
		return nil, err
	}
	if err := s.attachPhotos(ctx, entries, subjects); err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []Entry{}
	}
	return entries, nil
}

const entrySQL = `
select
    pub.id::text,
    pub.subject_kind,
    pub.intent,
    pub.subject_id::text,
    case when pub.author_visibility = 'public' or $1::boolean then u.display_name end,
    case when pub.author_visibility = 'public' or $1::boolean then u.id::text end,
    case when pub.author_visibility = 'public' or $1::boolean then u.username end,
    coalesce(pub.occurred_at, ev.starts_at, pub.created_at),
    coalesce(pub.lon, pl.lon, ev.lon),
    coalesce(pub.lat, pl.lat, ev.lat),
    coalesce(nullif(pub.place_label, ''), nullif(pl.address, ''), nullif(ev.address, ''), ''),
    coalesce(nullif(pub.body, ''), nullif(it.description, ''), nullif(pl.description, ''), nullif(ev.description, ''), ''),
    coalesce(
        nullif(it.name, ''),
        nullif(trim(both from pe.given_name || ' ' || pe.family_name), ''),
        nullif(an.nickname, ''),
        nullif(pl.name, ''),
        nullif(ev.name, ''),
        ''
    ),
    coalesce(itc.slug, pec.slug, anc.slug, plc.slug, evc.slug, '')
from publications pub
join users u on u.id = pub.author_user_id
left join items it on pub.subject_kind = 'item' and it.id = pub.subject_id
left join categories itc on itc.id = it.subcategory_id
left join persons pe on pub.subject_kind = 'person' and pe.id = pub.subject_id
left join categories pec on pec.id = pe.subcategory_id
left join animals an on pub.subject_kind = 'animal' and an.id = pub.subject_id
left join categories anc on anc.id = an.breed_category_id
left join places pl on pub.subject_kind = 'place' and pl.id = pub.subject_id
left join categories plc on plc.id = pl.subcategory_id
left join events ev on pub.subject_kind = 'event' and ev.id = pub.subject_id
left join categories evc on evc.id = ev.subcategory_id
where pub.status in ('published', 'matched')`

func (s *Store) attachTraits(ctx context.Context, entries []Entry, subjects map[string][]int) error {
	rows, err := s.pool.Query(ctx, `
		select subject_kind, subject_id::text, key, value
		from traits
		where visibility = 'public'
		order by sort_order, key`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var kind, id, key, value string
		if err := rows.Scan(&kind, &id, &key, &value); err != nil {
			return err
		}
		for _, index := range subjects[kind+"/"+id] {
			entries[index].Traits = append(entries[index].Traits, TraitLine{Name: traitName(kind, key), Value: value})
		}
	}
	return rows.Err()
}

func (s *Store) attachTags(ctx context.Context, entries []Entry, subjects map[string][]int) error {
	rows, err := s.pool.Query(ctx, `
		select 'place', place_id::text, tag from place_tags
		union all
		select 'event', event_id::text, tag from event_tags
		order by 1, 2, 3`)
	if err != nil {
		return err
	}
	defer rows.Close()
	grouped := map[string][]string{}
	for rows.Next() {
		var kind, id, tag string
		if err := rows.Scan(&kind, &id, &tag); err != nil {
			return err
		}
		grouped[kind+"/"+id] = append(grouped[kind+"/"+id], tag)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for key, tags := range grouped {
		if len(tags) == 0 {
			continue
		}
		line := TraitLine{Name: "Теги", Value: strings.Join(tags, ", ")}
		for _, index := range subjects[key] {
			entries[index].Traits = append(entries[index].Traits, line)
		}
	}
	return nil
}

func (s *Store) attachPhotos(ctx context.Context, entries []Entry, subjects map[string][]int) error {
	rows, err := s.pool.Query(ctx, `
		select parent_kind, parent_id::text, count(*)
		from attachments
		where visibility = 'public'
		group by parent_kind, parent_id`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var kind, id string
		var count int
		if err := rows.Scan(&kind, &id, &count); err != nil {
			return err
		}
		for _, index := range subjects[kind+"/"+id] {
			entries[index].PhotoCount = count
		}
	}
	return rows.Err()
}

func traitName(kind, key string) string {
	if key == "color" && kind == "animal" {
		return "Окрас"
	}
	switch key {
	case "color":
		return "Цвет"
	case "distinctive":
		return "Отличительная черта"
	case "height_cm":
		return "Рост"
	case "weight_kg":
		return "Вес"
	case "nationality":
		return "Национальность"
	case "appearance":
		return "Одежда"
	case "age":
		return "Возраст"
	case "pol":
		return "Пол"
	default:
		return key
	}
}

func sexLabel(kind, sex string) string {
	if kind == "animal" {
		if sex == "female" {
			return "самка"
		}
		return "самец"
	}
	if sex == "female" {
		return "женщина"
	}
	return "мужчина"
}

func splitName(title string) (string, string) {
	parts := strings.Fields(title)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func idPtr(id string) *domain.ID {
	value := domain.ID(id)
	return &value
}

func errorsIsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
