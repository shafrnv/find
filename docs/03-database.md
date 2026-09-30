# База данных

## Где крутится

`docker-compose.yml` поднимает `postgres:16-alpine`:

- хост: `127.0.0.1:5433` (не 5432 — на машине уже может быть чужой Postgres);
- user / password / db: `find` / `find` / `find`;
- volume: `find-pg`;
- healthcheck: `pg_isready -U find -d find`.

Данные живут, пока volume не удалили (`docker compose down -v` сотрёт всё).

## Миграции

Файлы в `migrations/`, применяются **при старте API** (`store.Migrate`):

1. Создаётся служебная таблица `schema_migrations (name, applied_at)`.
2. Все `*.sql` сортируются по имени и применяются по одному в транзакции.
3. Уже применённый файл пропускается.

| Файл | Что делает |
|---|---|
| `001_entities.sql` | Все доменные таблицы |
| `002_auth.sql` | `users.password_hash`, `persons.subcategory_id`, таблица `sessions` |
| `003_dialogs.sql` | `dialogs` + `messages` для отклика |
| `004_dialogs_any_loss.sql` | rename found→publication / finder→author / claimer→respondent |
| `005_chat_verification.sql` | статусы верификации чата, `respondent_answer`, `verified_at` |

Каталог задаётся `MIGRATIONS_DIR` (cwd процесса должен видеть путь; при `go run ./cmd/api` из корня репо — `migrations`).

## Карта таблиц

### Пользователи и вход

```
users
  id, username (unique), display_name, bio, avatar_media_id → media,
  created_at, password_hash

identity_verifications
  user_id PK → users, status (none|pending|verified), verified_at

sessions
  token_hash PK (sha256 hex от cookie), user_id → users,
  created_at, expires_at
```

Паспорт и платежи **не** хранятся. Cookie в браузере — сырой токен; в БД только хеш.

### Справочник

```
categories
  id, scope (item|person|animal|place|event),
  parent_id → categories | null, name, slug
  unique (scope, slug)
```

Корень: `parent_id IS NULL`, slug = scope. Подкатегория: `parent_id` = корень.

### Медиа

```
media
  id, owner_user_id → users, storage_key unique, mime_type, created_at

attachments
  (parent_kind, parent_id, media_id) PK
  parent_kind: item|person|animal|place|event|publication
  visibility, sort_order
```

`users.avatar_media_id` ссылается на `media` (FK добавлен после создания media).

### Субъекты

```
items          — name, category_id, subcategory_id, description, owner_user_id (internal)
persons        — given/family/patronymic, sex, subcategory_id
animals        — category_id (вид), breed_category_id / breed_text, sex, nickname, owner_user_id
places         — category/sub, name, description, address, lat/lon, created_by
place_tags     — (place_id, tag)
place_external_refs — yandex|2gis|google|osm + snapshot jsonb
events         — category/sub, name, description, place_id?, address, lat/lon, starts_at, ends_at?, created_by
event_tags     — (event_id, tag)
```

### Объявления и связи

```
publications
  author_user_id, author_visibility, intent, subject_kind, subject_id
  body, occurred_at, lat, lon, place_label, place_id?, status
  created_at, updated_at
  — subject_id БЕЗ FK: указывает в одну из пяти таблиц по subject_kind

traits
  subject_kind, subject_id, key, value, visibility, sort_order

follows
  follower_user_id, target_kind (user|place|event), target_id

matches
  seeking_publication_id, found_publication_id, status
```

Индексы: `publications_author_idx`, `publications_subject_idx`, `traits_subject_idx`, `follows_target_idx`, `sessions_user_idx`.

`matches` — доменная заготовка «свести два объявления одного субъекта». Отклик UI **«Это моё»** использует отдельные таблицы из `003_dialogs.sql`:

```
dialogs
  found_publication_id, finder_user_id, claimer_user_id
  unique (found_publication_id, claimer_user_id)

messages
  dialog_id, sender_user_id, body, created_at
```

Подробно — [08-dialogs.md](08-dialogs.md) (отклик) и [09-chat.md](09-chat.md) (чат и верификация).

## Связи, которые важно помнить

1. **Publication → Subject** — полиморфная: `subject_kind` + `subject_id`.
2. **Publication → User** — автор; для seeking автор может быть internal.
3. **Item/Animal.owner_user_id** — внутренний владелец, не автор объявления.
4. **Place/Event.created_by** — кто создал карточку места/события.
5. **Attachments** вешаются на субъект (в seed) или потенциально на publication; UI считает `photoCount` по public attachments субъекта.
6. **Session** → User; удаление пользователя каскадом сносит сессии.
7. **Dialog** → found publication + finder + claimer; сообщения каскадом при удалении диалога.

## Seed (наполнение)

Не SQL-файл, а `store.Seed` после миграций. Если `categories` уже не пусты — seed пропускается. Иначе:

- 5 демо-пользователей (`maria`, `olga`, `irina`, `alexey`, `kirill`) со случайным bcrypt-паролем (войти как они **нельзя** без смены хеша — для UI регистрируют своего);
- дерево категорий;
- 11 публикаций, совпадающих со старым демо-списком на карте (ключи, телефон, рюкзак, Иван Петров, …).

Идентификаторы в seed детерминированные: `idn(group, n)` → UUID вида `018fXX00-0000-7000-8000-…………`.
