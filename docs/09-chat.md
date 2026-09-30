# Чат (самостоятельный блок)

Чат — **отдельный архитектурный модуль**, не часть карты. Карта только инициирует отклик (`claim`); дальше жизнь диалога живёт в блоке chat.

Встреча нашедшего и ищущего — **следующий** самостоятельный блок: `internal/meeting` (пока якорь без API). Карта и chat его не импортируют.

## Пакеты

| Слой | Путь | Назначение |
|---|---|---|
| Домен статусов | `internal/chat` | `pending_answer` → `pending_confirm` → `open` / `rejected`; `CanSendMessage` |
| Store | `internal/store/chat.go` | Claim, answer, confirm, messages |
| HTTP | `internal/httpapi/chat.go` | `/api/dialogs…`, claim |
| UI | `web/src/chat/` | `ChatsPanel`, верификация, переписка |
| Встреча (позже) | `internal/meeting` | место/время, handshake / QR |

## Верификация по отличительной особенности

При публикации находки автор может задать **скрытую** примету (`secretTrait` → `traits` с `key=distinctive`, `visibility=internal`). В карточке её нет.

```
Отклик (claim)
  ├─ нет скрытой приметы → status=open, можно писать сразу
  └─ есть скрытая примета → status=pending_answer
        → откликнувшийся POST …/answer
        → status=pending_confirm
        → автор POST …/confirm { accept: true|false }
              ├─ true  → open (+ системное сообщение)
              └─ false → rejected
```

Сообщения (`POST …/messages`) принимаются **только** при `status=open`. Сравнение строк на сервере не делается: автор сам подтверждает, что ответ верный.

## Статусы (`migrations/005_chat_verification.sql`)

| status | Смысл |
|---|---|
| `pending_answer` | ждут ответ откликнувшегося |
| `pending_confirm` | автор смотрит ответ |
| `open` | переписка |
| `rejected` | примета не подтверждена |
| `closed` | зарезервировано |

Поля: `respondent_answer`, `verified_at`.

## API

| Метод | Путь | Кто |
|---|---|---|
| `POST` | `/api/entries/{id}/claim` | отклик → создаёт/возвращает диалог |
| `GET` | `/api/dialogs` | свои диалоги (+ `status`) |
| `GET` | `/api/dialogs/{id}` | деталь: флаги `canAnswer` / `canConfirm` / `canMessage` |
| `POST` | `/api/dialogs/{id}/answer` | `{ "answer": "…" }` — только respondent |
| `POST` | `/api/dialogs/{id}/confirm` | `{ "accept": true\|false }` — только author |
| `POST` | `/api/dialogs/{id}/messages` | `{ "body": "…" }` — только при `open` |

Публикация находки: в `POST /api/found` поле `secretTrait` (опционально).

## UI

Форма «Я нашёл»: публичная примета + «Отличительная особенность (для проверки при чате)».

`web/src/chat/ChatsPanel` — до открытия сообщений показывает:

- респонденту форму ответа;
- автору ответ и кнопки «Да, это та» / «Нет, отклонить».

Карта (`MapPage`) импортирует чат как модуль: `import { ChatsPanel, claimFound } from "../chat"`.

## Демо в сиде

У «Связка ключей» есть внутренняя примета `гравировка «М+О» на кольце`. Отклик на неё без ответа на примету не откроет переписку.

## Связь с 08

[08-dialogs.md](08-dialogs.md) — принцип «нет роли у аккаунта» и кнопки отклика. Этот файл — **жизнь чата после отклика** и граница с meeting.
