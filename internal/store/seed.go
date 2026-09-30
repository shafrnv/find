package store

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	groupUser    = 0x10
	groupRoot    = 0x20
	groupSub     = 0x30
	groupSubject = 0x40
	groupPost    = 0x50
)

func (s *Store) Seed(ctx context.Context) error {
	var n int
	if err := s.pool.QueryRow(ctx, `select count(*) from categories`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := seed(ctx, tx, time.Now()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func seed(ctx context.Context, tx pgx.Tx, now time.Time) error {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword(secret, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []struct{ username, name string }{
		{"maria", "Мария Соколова"},
		{"olga", "Ольга Новикова"},
		{"irina", "Ирина Лебедева"},
		{"alexey", "Алексей Ким"},
		{"kirill", "Кирилл Орлов"},
	}
	for i, user := range users {
		id := idn(groupUser, i+1)
		if _, err := tx.Exec(ctx, `
			insert into users (id, username, display_name, bio, created_at, password_hash)
			values ($1, $2, $3, '', $4, $5)`, id, user.username, user.name, now, string(hash)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			insert into identity_verifications (user_id, status, verified_at)
			values ($1, 'none', null)`, id); err != nil {
			return err
		}
	}

	roots := []struct{ slug, name string }{
		{"item", "Предметы"},
		{"person", "Люди"},
		{"animal", "Животные"},
		{"place", "Места"},
		{"event", "События"},
	}
	for i, root := range roots {
		if _, err := tx.Exec(ctx, `
			insert into categories (id, scope, parent_id, name, slug)
			values ($1, $2, null, $3, $4)`, idn(groupRoot, i+1), root.slug, root.name, root.slug); err != nil {
			return err
		}
	}
	subs := []struct{ slug, name, parent string }{
		{"keys", "Ключи", "item"},
		{"electronics", "Электроника", "item"},
		{"bags", "Сумки", "item"},
		{"adults", "Взрослые", "person"},
		{"children", "Дети", "person"},
		{"cats", "Кошки", "animal"},
		{"dogs", "Собаки", "animal"},
		{"yards", "Дворы", "place"},
		{"viewpoints", "Смотровые", "place"},
		{"gatherings", "Сборы", "event"},
		{"meetings", "Встречи", "event"},
	}
	rootN := map[string]int{"item": 1, "person": 2, "animal": 3, "place": 4, "event": 5}
	for i, sub := range subs {
		if _, err := tx.Exec(ctx, `
			insert into categories (id, scope, parent_id, name, slug)
			values ($1, $2, $3, $4, $5)`,
			idn(groupSub, i+1), sub.parent, idn(groupRoot, rootN[sub.parent]), sub.name, sub.slug); err != nil {
			return err
		}
	}

	hour := time.Hour
	day := 24 * hour
	maria, olga, irina, alexey, kirill := idn(groupUser, 1), idn(groupUser, 2), idn(groupUser, 3), idn(groupUser, 4), idn(groupUser, 5)
	posts := []seedPost{
		{1, "item", "keys", "found", maria, "Связка ключей", "Три ключа на металлическом кольце и круглый брелок. Лежали на скамейке у выхода к Манежной.", "Манежная площадь", 37.6208, 55.7539, -3 * hour, 2, "", []seedTrait{{"color", "серебристый металл", false}, {"distinctive", "брелок-монета", false}, {"distinctive", "гравировка «М+О» на кольце", true}}, nil},
		{2, "item", "electronics", "seeking", olga, "Телефон в сером чехле", "Пропал после пересадки. Чехол матовый, серый, без рисунка. Экран был заблокирован.", "Китай-город", 37.633, 55.7562, -6 * day, 1, "", []seedTrait{{"color", "серый", false}}, nil},
		{3, "item", "bags", "found", olga, "Чёрный рюкзак", "Небольшой городской рюкзак, оставлен у колонны в зале ожидания. Снаружи без нашивок.", "Белорусский вокзал", 37.5816, 55.7764, -26 * hour, 2, "", []seedTrait{{"color", "чёрный", false}}, nil},
		{4, "person", "adults", "seeking", maria, "Иван Петров", "Вышел из дома утром и не вышел на связь. Последний раз его видели на Тверской, шёл в сторону Пушкинской.", "Тверская", 37.6065, 55.7642, -2 * day, 1, "male", []seedTrait{{"height_cm", "180 см", false}, {"appearance", "тёмная куртка, синие джинсы", false}}, nil},
		{5, "person", "adults", "found", irina, "Мужчина в синей куртке", "Сидел на лавочке у главного входа, выглядел растерянным и не мог назвать адрес.", "Парк Горького", 37.6034, 55.7312, -9 * hour, 1, "male", []seedTrait{{"age", "около 60", false}, {"appearance", "синяя куртка, серая шапка", false}}, nil},
		{6, "animal", "cats", "seeking", alexey, "Кот Рыжик", "Выскочил из переноски у пруда. Отзывается на кличку, пугливый с незнакомцами.", "Патриаршие пруды", 37.5922, 55.7638, -5 * hour, 2, "male", []seedTrait{{"color", "рыжий", false}, {"pol", "кот", false}}, nil},
		{7, "animal", "dogs", "found", alexey, "Собака, дворняга", "Без ошейника, подошла сама и держится рядом. На ухе небольшая засечка.", "Сокольники", 37.674, 55.794, -18 * day, 1, "female", []seedTrait{{"color", "серо-белый", false}, {"pol", "сука", false}}, nil},
		{8, "place", "yards", "marking", kirill, "Тихий двор", "Вход со стороны переулка, во дворе скамейки и почти нет машин после девяти вечера.", "Арбат", 37.5928, 55.7494, -4 * day, 2, "", nil, []string{"тихий", "скамейки"}},
		{9, "place", "viewpoints", "marking", kirill, "Смотровая", "Площадка чуть в стороне от основной смотровой: меньше людей и нормальный вид на излучину.", "Воробьёвы горы", 37.545, 55.710, -40 * day, 2, "", nil, []string{"вид", "меньше людей"}},
		{10, "event", "gatherings", "marking", irina, "Сбор находок", "Короткий сбор у входа: приносят найденные вещи без хозяина и сверяют объявления.", "Лужники", 37.5542, 55.7154, -20 * hour, 1, "", nil, nil},
		{11, "event", "meetings", "marking", maria, "Встреча волонтёров", "Встреча тех, кто помогает сверять ориентировки. Сбор у главного входа, без записи.", "ВДНХ", 37.624, 55.826, 2 * day, 1, "", nil, nil},
	}
	subN := map[string]int{}
	for i, sub := range subs {
		subN[sub.slug] = i + 1
	}
	for _, post := range posts {
		if err := insertSeedPost(ctx, tx, now, post, rootN, subN); err != nil {
			return fmt.Errorf("%s: %w", post.title, err)
		}
	}
	return nil
}

type seedTrait struct {
	key, value string
	internal   bool // true → visibility=internal (для проверки в чате)
}

type seedPost struct {
	n        int
	kind     string
	sub      string
	intent   string
	author   string
	title    string
	body     string
	place    string
	lon, lat float64
	offset   time.Duration
	photos   int
	sex      string
	traits   []seedTrait
	tags     []string
}

func insertSeedPost(ctx context.Context, tx pgx.Tx, now time.Time, post seedPost, rootN, subN map[string]int) error {
	subjectID := idn(groupSubject, post.n)
	rootID := idn(groupRoot, rootN[scopeOf(post.kind)])
	subID := idn(groupSub, subN[post.sub])
	at := now.Add(post.offset)
	owner := any(nil)
	if post.intent == "seeking" {
		owner = post.author
	}
	switch post.kind {
	case "item":
		if _, err := tx.Exec(ctx, `
			insert into items (id, name, category_id, subcategory_id, description, owner_user_id, created_at)
			values ($1,$2,$3,$4,$5,$6,$7)`,
			subjectID, post.title, rootID, subID, post.body, owner, now); err != nil {
			return err
		}
	case "person":
		given, family := splitName(post.title)
		sex := post.sex
		if sex == "" {
			sex = "unknown"
		}
		if _, err := tx.Exec(ctx, `
			insert into persons (id, given_name, family_name, patronymic, sex, subcategory_id, created_at)
			values ($1,$2,$3,'',$4,$5,$6)`,
			subjectID, given, family, sex, subID, now); err != nil {
			return err
		}
	case "animal":
		sex := post.sex
		if sex == "" {
			sex = "unknown"
		}
		if _, err := tx.Exec(ctx, `
			insert into animals (id, category_id, breed_category_id, breed_text, sex, nickname, owner_user_id, created_at)
			values ($1,$2,$3,'',$4,$5,$6,$7)`,
			subjectID, rootID, subID, sex, post.title, owner, now); err != nil {
			return err
		}
	case "place":
		if _, err := tx.Exec(ctx, `
			insert into places (id, category_id, subcategory_id, name, description, address, lat, lon, created_by, created_at)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			subjectID, rootID, subID, post.title, post.body, post.place, post.lat, post.lon, post.author, now); err != nil {
			return err
		}
		for _, tag := range post.tags {
			if _, err := tx.Exec(ctx, `insert into place_tags (place_id, tag) values ($1, $2)`, subjectID, tag); err != nil {
				return err
			}
		}
	case "event":
		if _, err := tx.Exec(ctx, `
			insert into events (
				id, category_id, subcategory_id, name, description, address, lat, lon, starts_at, created_by, created_at
			) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			subjectID, rootID, subID, post.title, post.body, post.place, post.lat, post.lon, at, post.author, now); err != nil {
			return err
		}
		for _, tag := range post.tags {
			if _, err := tx.Exec(ctx, `insert into event_tags (event_id, tag) values ($1, $2)`, subjectID, tag); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("неизвестный субъект %s", post.kind)
	}

	for i, trait := range post.traits {
		traitID, err := newID()
		if err != nil {
			return err
		}
		vis := "public"
		if trait.internal {
			vis = "internal"
		}
		if _, err := tx.Exec(ctx, `
			insert into traits (id, subject_kind, subject_id, key, value, visibility, sort_order)
			values ($1,$2,$3,$4,$5,$6,$7)`,
			traitID, post.kind, subjectID, trait.key, trait.value, vis, i); err != nil {
			return err
		}
	}
	for i := 0; i < post.photos; i++ {
		mediaID, err := newID()
		if err != nil {
			return err
		}
		key := fmt.Sprintf("seed/%s/%d.png", subjectID, i+1)
		if _, err := tx.Exec(ctx, `
			insert into media (id, owner_user_id, storage_key, mime_type, created_at)
			values ($1,$2,$3,'image/png',$4)`, mediaID, post.author, key, now); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			insert into attachments (parent_kind, parent_id, media_id, visibility, sort_order)
			values ($1,$2,$3,'public',$4)`, post.kind, subjectID, mediaID, i); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
		insert into publications (
			id, author_user_id, author_visibility, intent, subject_kind, subject_id,
			body, occurred_at, lat, lon, place_label, status, created_at, updated_at
		) values ($1,$2,'public',$3,$4,$5,$6,$7,$8,$9,$10,'published',$11,$11)`,
		idn(groupPost, post.n), post.author, post.intent, post.kind, subjectID,
		post.body, at, post.lat, post.lon, post.place, now); err != nil {
		return err
	}
	return nil
}

func scopeOf(kind string) string { return kind }
