# Find — централизованный поиск

Поиск предметов, людей, животных, мест и событий на карте. Стек: **Go + Postgres** на сервере, **React + TypeScript + Vite** на клиенте, **Яндекс.Карты JS API 2.1** как фон.

## Документация по блокам

| Документ | О чём |
|---|---|
| [docs/01-overview.md](docs/01-overview.md) | Как устроен проект целиком, потоки данных, порты |
| [docs/02-domain.md](docs/02-domain.md) | Доменные сущности в Go: что ищется, кто пишет, видимость |
| [docs/03-database.md](docs/03-database.md) | Postgres: таблицы, связи, миграции |
| [docs/04-store.md](docs/04-store.md) | Слой записи/чтения: auth, лента карты, «я нашёл», seed |
| [docs/05-api.md](docs/05-api.md) | HTTP API, cookie-сессии, тела запросов |
| [docs/06-frontend.md](docs/06-frontend.md) | Карта, панель «В области», фильтры, форма находки |
| [docs/07-commands.md](docs/07-commands.md) | Все команды консоли: запуск, проверки, curl |
| [docs/08-dialogs.md](docs/08-dialogs.md) | Нет роли у аккаунта; кнопки отклика |
| [docs/09-chat.md](docs/09-chat.md) | Блок чата: верификация по примете, статусы, API |
| [docs/10-meeting.md](docs/10-meeting.md) | Блок встречи (якорь; место/время, handshake позже) |

## Быстрый старт

Из корня репозитория `/home/shafrnv/find` (или клона):

```bash
# 1. База
docker compose up -d

# 2. API (миграции и seed сами при старте)
go run ./cmd/api

# 3. Фронт (в другом терминале)
cd web && npm install && npm run dev
```

Открыть: http://127.0.0.1:5173/

Подробности и разбор вывода команд — в [docs/07-commands.md](docs/07-commands.md).
