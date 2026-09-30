# HTTP API (`cmd/api` + `internal/httpapi`)

## Запуск процесса

`cmd/api/main.go`:

1. Читает `DATABASE_URL`, `MIGRATIONS_DIR`.
2. Ждёт Postgres (до 30 попыток).
3. `Migrate` → `Seed`.
4. Слушает **только** `127.0.0.1:8080` (не снаружи машины).
5. `ReadHeaderTimeout: 5s`.

Лог успеха: `api http://127.0.0.1:8080`.

## Маршруты

Регистрация через Go 1.22+ pattern: `"METHOD /path"`.

| Метод | Путь | Auth | Назначение |
|---|---|---|---|
| `GET` | `/api/entries` | нет | Все опубликованные карточки для карты |
| `GET` | `/api/me` | cookie | Текущий аккаунт (id, username, displayName, bio, identityStatus) |
| `PATCH` | `/api/me` | cookie | Обновить displayName и bio |
| `POST` | `/api/register` | нет → ставит cookie | Регистрация |
| `POST` | `/api/login` | нет → ставит cookie | Вход |
| `POST` | `/api/logout` | cookie optional | Удалить сессию |
| `POST` | `/api/found` | cookie | Создать объявление «нашёл» |
| `POST` | `/api/entries/{id}/claim` | cookie | «Это моё» → создать/вернуть диалог |
| `GET` | `/api/dialogs` | cookie | Список чатов (+ status) |
| `GET` | `/api/dialogs/{id}` | cookie | Диалог: флаги canAnswer/canConfirm/canMessage |
| `POST` | `/api/dialogs/{id}/answer` | cookie | Ответ на отличительную особенность |
| `POST` | `/api/dialogs/{id}/confirm` | cookie | Подтверждение/отклонение ответа автором |
| `POST` | `/api/dialogs/{id}/messages` | cookie | Сообщение (только status=open) |
| `GET` | `/api/users/{username}` | cookie optional | Публичный профиль + `entryCount`, флаг `mine` |
| `GET` | `/api/users/{username}/entries` | cookie optional | Лента объявлений автора |

Ответы: `Content-Type: application/json; charset=utf-8`.  
Ошибки: `{"error":"текст на русском"}`.

Тело запроса ограничено `MaxBytesReader` **1 MiB**.

## Cookie-сессия

- Имя: `find_session`
- `HttpOnly`, `SameSite=Lax`, `Path=/`
- Значение: сырой токен; в Postgres — SHA-256
- При register/login выставляется `Expires` ≈ +30 дней
- Logout: delete в БД + cookie с `MaxAge: -1`

Кросс-домен без прокси не настроен: фронт ходит на тот же origin через Vite proxy `/api` → `:8080`, cookie идёт на `127.0.0.1:5173`.

## Контракты

### `GET /api/entries` → `200` массив Entry

См. форму в [04-store.md](04-store.md).

### `GET /api/me` → `200` Account | `401`

```json
{ "id": "…", "username": "finder", "displayName": "Нашёдший" }
```

### `POST /api/register`

```json
{ "username": "finder", "displayName": "Нашёдший", "password": "secret-pass" }
```

- `201` + Account + Set-Cookie  
- `409` имя занято  
- `400` валидация domain/пароля  

### `POST /api/login`

```json
{ "username": "finder", "password": "secret-pass" }
```

- `200` + cookie  
- `401` неверный логин/пароль  

### `POST /api/logout`

Тело `{}` или пустое после decode; `200` `{"ok":true}`.

### `POST /api/found`

Требует сессию. Тело:

```json
{
  "kind": "item",
  "subcategory": "keys",
  "title": "Синий зонт",
  "description": "…",
  "place": "Манежная площадь",
  "at": "2026-09-29T12:45",
  "lat": 55.7558,
  "lon": 37.6173,
  "sex": "unknown",
  "trait": "деревянная ручка"
}
```

`at` парсится как `2006-01-02T15:04` в локальной TZ, иначе RFC3339.  
`201` → готовый Entry.  
`401` без сессии.  
`400` ошибки domain/store.

Для person/animal поле `sex`: `female` | `male` | `unknown`. Для item фронт шлёт `unknown`.

## Цепочка вызова «нашёл»

```
browser FoundForm
  → api.createFound (fetch POST /api/found, credentials include cookies по same-origin)
  → httpapi.found
      → current() читает cookie
      → store.CreateFound
  → JSON Entry
  → MapPage: setEntries + select(id) → ObjectCard
```

Полные контракты claim/dialogs — в [08-dialogs.md](08-dialogs.md) и [09-chat.md](09-chat.md).
