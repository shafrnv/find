# Команды консоли

Все команды ниже — из корня репозитория, если не сказано иначе. Путь примера: `/home/shafrnv/find`.

## 1. Postgres

```bash
# Поднять контейнер (порт 5433, volume find-pg)
docker compose up -d

# Статус и health
docker compose ps

# Логи Postgres
docker compose logs -f postgres

# Остановить (данные в volume сохранятся)
docker compose stop

# Остановить и удалить volume (полная очистка БД)
docker compose down -v
```

Зайти в psql внутри контейнера (клиента `psql` на хосте может не быть):

```bash
docker exec -it find-postgres-1 psql -U find -d find
```

Полезные SQL внутри сессии:

```sql
\dt
select name from schema_migrations order by name;
select count(*) from publications;
select username, display_name from users order by created_at;
select intent, subject_kind, place_label from publications order by created_at desc limit 20;
```

Одной командой с хоста:

```bash
docker exec find-postgres-1 psql -U find -d find -c 'select count(*) from publications;'
```

## 2. Go API

Нужны Go 1.26+ и модули из `go.mod` (первый раз скачает зависимости).

```bash
# Подтянуть зависимости (если нужно)
go mod tidy

# Юнит-тесты домена
go test ./internal/domain/

# Сборка бинарника
go build -o /tmp/find-api ./cmd/api

# Запуск (миграции + seed + listen :8080)
go run ./cmd/api
# или
/tmp/find-api
```

Переменные окружения:

```bash
DATABASE_URL='postgres://find:find@127.0.0.1:5433/find?sslmode=disable' \
MIGRATIONS_DIR=migrations \
go run ./cmd/api
```

Успешный лог: `api http://127.0.0.1:8080`.  
Если Postgres ещё не готов: строки `postgres: …` раз в секунду, потом fatal.

Проверка API без браузера:

```bash
# Лента
curl -sS http://127.0.0.1:8080/api/entries | python3 -m json.tool | head

# Регистрация + cookie-jar
curl -sS -c /tmp/find.cookies -H 'Content-Type: application/json' \
  -d '{"username":"demo_user","displayName":"Демо","password":"secret-pass"}' \
  http://127.0.0.1:8080/api/register

# Кто я
curl -sS -b /tmp/find.cookies http://127.0.0.1:8080/api/me

# Создать находку (координаты Москвы)
curl -sS -b /tmp/find.cookies -H 'Content-Type: application/json' \
  -d '{
    "kind":"item",
    "subcategory":"keys",
    "title":"Тестовый зонт",
    "description":"Проверка API",
    "place":"Центр",
    "at":"2026-09-29T12:40",
    "lat":55.7558,
    "lon":37.6173,
    "sex":"unknown",
    "trait":"ручка"
  }' \
  http://127.0.0.1:8080/api/found

# Выход
curl -sS -b /tmp/find.cookies -c /tmp/find.cookies \
  -H 'Content-Type: application/json' -d '{}' \
  http://127.0.0.1:8080/api/logout
```

Отклик «Это моё» и переписка — примеры в [08-dialogs.md](08-dialogs.md).

## 3. Фронтенд

```bash
cd web

# Зависимости (один раз)
npm install

# Проверка типов
npx tsc --noEmit

# Dev-сервер на 127.0.0.1:5173 с proxy /api
npm run dev
# эквивалент:
npx vite --host 127.0.0.1 --port 5173

# Production-сборка
npm run build
npm run preview
```

Проверка, что Vite проксирует API:

```bash
curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:5173/api/entries
# ожидается 200, если API уже запущен
```

Открыть в браузере: http://127.0.0.1:5173/

## 4. Типовой порядок «с нуля»

Три терминала:

```bash
# Терминал A
cd /home/shafrnv/find
docker compose up -d
docker compose ps   # STATUS должен стать healthy

# Терминал B
cd /home/shafrnv/find
go run ./cmd/api

# Терминал C
cd /home/shafrnv/find/web
npm install         # если ещё не ставили
npm run dev
```

## 5. Что вызывал агент при разработке (шпаргалка)

Эти команды реально использовались при сборке текущего состояния:

```bash
docker compose up -d
docker inspect kisselenko-postgres-1   # чужой Postgres на :5432 — поэтому свой на :5433
go get github.com/jackc/pgx/v5@v5.7.4 golang.org/x/crypto@v0.36.0
go get github.com/jackc/pgx/v5/pgxpool@v5.7.4
gofmt -w internal/store/*.go internal/httpapi/*.go cmd/api/main.go
go test ./internal/domain/
go build ./...
go build -o /tmp/find-api ./cmd/api && /tmp/find-api
npx tsc --noEmit
npx vite --host 127.0.0.1 --port 5173
curl -sS http://127.0.0.1:8080/api/entries
curl … /api/register && curl … /api/found
docker exec find-postgres-1 psql -U find -d find -c '…'
```

Яндекс ключ вшит в `web/index.html` (query `apikey=`). Для проверки API 2.1 раньше смотрели HTTP 200 у `api-maps.yandex.ru/2.1/…`.

## 6. Диагностика

| Симптом | Что проверить |
|---|---|
| `postgres: connection refused` в логе API | `docker compose ps`, порт 5433 |
| Фронт «Не удалось открыть объявления» | Запущен ли `:8080`, proxy в `vite.config.ts` |
| Пустая карта / 0 из 0 | Bounds ещё не пришли; ждём ymaps; смотрим Network → `/api/entries` |
| 401 на «Опубликовать» / «Это моё» | Нет cookie — зарегистрироваться |
| Подкатегория «не относится» | slug должен быть из seed (`keys`, `cats`, …) и scope = kind |
| «нельзя откликнуться на свою находку» | Открыта своя карточка `found` |
| Пустые Чаты | Ещё не было claim; смотрите `dialogs` в Postgres |
| Старые демо-данные пропали | Был `down -v` или ручной TRUNCATE; перезапуск API снова сделает seed только если categories пусты |

Сброс только seed при живой схеме: удалить данные вручную через psql **или** `docker compose down -v && docker compose up -d` и снова `go run ./cmd/api`.
