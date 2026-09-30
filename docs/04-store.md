# Store (`internal/store`)

Слой между доменом и Postgres. Один тип `Store` держит `*pgxpool.Pool`.

## Файлы

| Файл | Роль |
|---|---|
| `store.go` | `Open`, `Close`, генерация UUID (`newID`), helper `idn` для seed |
| `migrate.go` | Применение `migrations/*.sql` |
| `auth.go` | Регистрация, логин, сессии |
| `profile.go` | Профиль по username, лента автора, обновление bio |
| `entries.go` | Лента карты `Entry`, `CreateFound` (`trait` + `secretTrait`) |
| `chat.go` | `ClaimFound`, верификация, список/чтение, `SendMessage` |
| `seed.go` | Первичное наполнение |

## Подключение

`store.Open(ctx, dsn)`:

1. `pgxpool.New`
2. `Ping` с таймаутом 3 с
3. При ошибке пул закрывается

`cmd/api` при старте **до 30 раз** с паузой 1 с пытается открыть пул (ждёт healthy Postgres).

## Auth

### Register

1. Пароль 8–72 символа (лимит bcrypt).
2. `domain.NewUser` (username: `^[a-z0-9_]{3,32}$`, display name 1–80).
3. `bcrypt.GenerateFromPassword`.
4. Транзакция: `INSERT users` + `INSERT identity_verifications (none)`.
5. Unique violation `23505` → `ErrUsernameTaken`.

### Login

1. Нормализация username (trim + lower).
2. Чтение `password_hash`.
3. `bcrypt.CompareHashAndPassword` — при любом провале единый `ErrInvalidLogin` (без утечки «логин существует»).

### Session

- Сырой токен: 32 случайных байта → hex (64 символа), уходит в cookie `find_session`.
- В БД: `sha256(token)` как hex.
- TTL: 30 суток (`sessionTTL`).
- `AccountBySession`: join sessions↔users, `expires_at > now()`.
- `DeleteSession`: delete по хешу (logout).

## Entry (карточка для карты)

JSON-форма, которую ест фронт:

```json
{
  "id": "<publication uuid>",
  "kind": "item|person|animal|place|event",
  "subcategoryId": "keys",
  "intent": "found",
  "title": "…",
  "author": "Мария Соколова",
  "at": "2026-09-29T…",
  "coordinates": [lon, lat],
  "place": "Манежная площадь",
  "description": "…",
  "traits": [{"name":"Цвет","value":"…"}],
  "photoCount": 2
}
```

Координаты в API и на фронте: **`[lon, lat]`** (как в приложении). Яндекс 2.1 внутри хука получает `[lat, lon]` через `toMapPoint`.

### Чтение `Entries` / `loadEntries`

Большой SQL (`entrySQL`):

- `publications` + `users`
- LEFT JOIN на `items` / `persons` / `animals` / `places` / `events` и их подкатегории
- фильтр `status in ('published','matched')`
- время: `coalesce(occurred_at, event.starts_at, created_at)`
- точка: `coalesce(pub.lon/lat, place, event)`
- заголовок из name / ФИО / nickname / place.name / event.name
- автор: `display_name` только если `author_visibility = 'public'`, иначе пустая строка

После запроса:

1. `attachTraits` — public traits → человекочитаемые имена («Цвет», «Окрас», …)
2. `attachTags` — place_tags / event_tags → строка «Теги»
3. `attachPhotos` — count public attachments

Один субъект может иметь несколько публикаций; traits крепятся ко **всем** entry с этим `kind/subject_id` (map `[]int` индексов).

### CreateFound

Только `item` | `person` | `animal`.

В транзакции:

1. Найти root + subcategory id по `slug` и `scope`.
2. Создать субъект (`items` / `persons` / `animals`).
3. Опционально traits (примета `distinctive`, пол как `pol`).
4. `NewPublication(..., IntentFound, ...)` → сразу `WithStatus(published)`.
5. `INSERT publications` с lat/lon и place_label.
6. Commit → перечитать одну карточку через `loadEntries(ctx, publicationID)`.

Вход `FoundInput`: AuthorID, Kind, Subcategory (slug), Title, Description, Place, At, Lat, Lon, Sex, Trait (public), SecretTrait (internal для чата).

## Seed

См. также [03-database.md](03-database.md). Логика в `seed.go`: детерминированные UUID, относительные времена (`now - 3*hour`), медиа-заглушки с `storage_key` вида `seed/<subject>/<n>.png` без реальных файлов на диске.

Зависимости Go

- `github.com/jackc/pgx/v5` — драйвер и pool
- `golang.org/x/crypto/bcrypt` — пароли

## Диалоги

См. [09-chat.md](09-chat.md): `ClaimFound`, answer/confirm, `ListDialogs`, `Dialog`, `SendMessage` в `chat.go`.
