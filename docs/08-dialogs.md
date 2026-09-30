# Отклики «Это моё» и чаты

## Главный принцип: роли нет у человека

У аккаунта **нет** постоянной роли «нашедший» или «потерявший».

Один и тот же пользователь может:

- публиковать **находку** (`intent = found`);
- публиковать **поиск** (`intent = seeking`);
- **отмечать** места и события (`intent = marking`).

Роль существует только **у конкретного объявления** (`publications.intent`). В карточке подписи вроде «Кто нашёл» / «Кто ищет» относятся к автору **этого** поста, а не к типу аккаунта.

Отклик и диалог работают для **предметов, людей и животных** — и на объявлениях «Нашли», и на «Ищут». Места и события не откликаются.

| Объявление | Кнопка |
|---|---|
| Находка предмета / животного | **Это моё** |
| Находка / ориентировка человека | **Это мой человек** |
| Ищут предмет / животное | **Я нашёл** |
| Ищут человека | **Я видел** |

Таблица `matches` в схеме (связка seeking+found по одному субъекту) — заготовка домена; UI-отклик идёт через **dialogs**, без обязательного второго объявления.

## Зачем диалог

Человек открыл чужое объявление о пропаже или находке (предмет, человек, животное) и откликнулся → создаётся (или открывается) диалог между:

| Участник | Кто это |
|---|---|
| `author_user_id` | автор объявления |
| `respondent_user_id` | вошедший, нажавший кнопку отклика |

В JSON у зрителя `role`: `author` или `respondent` — только сторона **в этом** чате, не роль аккаунта.

## Поток

```
ObjectCard (intent=found, не своё)
  → POST /api/entries/{id}/claim   (нужна cookie-сессия)
  → store.ClaimFound
       • проверка: published|matched, intent=found, не свой пост
       • unique (found_publication_id, claimer_user_id)
       • если диалога нет — INSERT dialogs + первое message от claimer
       • если есть — вернуть существующий
  → ChatsPanel открывает диалог
  → дальше POST /api/dialogs/{id}/messages
```

Повторный «Это моё» на ту же находку тем же человеком **не** плодит второй чат — возвращается тот же `dialogs.id`.

Без входа кнопка ведёт к форме входа/регистрации (профиль).

На **свою** находку кнопка не показывается (`user.id === entry.authorUserId`).

## Таблицы (`migrations/003_dialogs.sql`)

```
dialogs
  id
  found_publication_id → publications
  finder_user_id       → users   (автор находки)
  claimer_user_id      → users   (откликнувшийся)
  created_at
  unique (found_publication_id, claimer_user_id)
  check (finder_user_id <> claimer_user_id)

messages
  id
  dialog_id → dialogs (on delete cascade)
  sender_user_id → users
  body  (не пустой)
  created_at
```

Индексы: участники диалогов, сообщения по `dialog_id + created_at`.

Первое сообщение при создании диалога (от claimer):

> Здравствуйте! Это моё. Давайте созвонимся по поводу возврата.

## Код

| Место | Что делает |
|---|---|
| `internal/store/dialogs.go` | `ClaimFound`, `ListDialogs`, `Dialog`, `SendMessage` |
| `internal/httpapi/server.go` | маршруты claim / dialogs / messages |
| `web/src/map/ObjectCard.tsx` | кнопка «Это моё», подсказка про отсутствие роли аккаунта |
| `web/src/map/ChatsPanel.tsx` | список чатов + переписка |
| `web/src/map/api.ts` | `claimFound`, `loadDialogs`, `loadDialog`, `sendDialogMessage` |
| `web/src/map/Menu.tsx` | пункт «Чаты» |
| `web/src/map/MapPage.tsx` | связка claim → открыть `ChatsPanel` с `dialogId` |

## API (кратко)

| Метод | Путь | Кто |
|---|---|---|
| `POST` | `/api/entries/{id}/claim` | claimer, cookie |
| `GET` | `/api/dialogs` | участник своих диалогов |
| `GET` | `/api/dialogs/{id}` | finder или claimer этого диалога |
| `POST` | `/api/dialogs/{id}/messages` | body: `{ "body": "…" }` |

Ответ claim / dialog:

```json
{
  "id": "…",
  "publicationId": "…",
  "title": "Связка ключей",
  "place": "Манежная площадь",
  "peerName": "Мария Соколова",
  "peerUsername": "maria",
  "role": "claimer",
  "createdAt": "…",
  "messages": [
    { "id": "…", "senderId": "…", "mine": true, "body": "…", "createdAt": "…" }
  ]
}
```

В списке `GET /api/dialogs` дополнительно: `lastBody`, `updatedAt`, без полного массива сообщений.

Ошибки claim:

- `401` — не вошли;
- `404` — нет объявления;
- `400` — не находка / своя находка / нельзя откликнуться.

## UI

1. Левая панель → карточка находки → блок с кнопкой **«Это моё»**.
2. Правая панель **Чаты** (как профиль справа): список или открытый диалог.
3. В диалоге: заголовок предмета, peer (нашедший или откликнувшийся), лента пузырей, поле ввода.

Подпись в списке чатов:

- если вы `finder` — «Откликнулся · {peer}»;
- если вы `claimer` — «Нашедший · {peer}».

## Пример curl

```bash
# войти
curl -sS -c /tmp/find.cookies -H 'Content-Type: application/json' \
  -d '{"username":"finder","password":"secret-pass"}' \
  http://127.0.0.1:8080/api/login

# id чужой находки из ленты
ID=$(curl -sS http://127.0.0.1:8080/api/entries | python3 -c \
  'import json,sys; print(next(e["id"] for e in json.load(sys.stdin) if e["intent"]=="found" and e.get("authorUsername")=="maria"))')

# отклик
curl -sS -b /tmp/find.cookies -H 'Content-Type: application/json' -d '{}' \
  "http://127.0.0.1:8080/api/entries/$ID/claim"

# список чатов и сообщение
curl -sS -b /tmp/find.cookies http://127.0.0.1:8080/api/dialogs
curl -sS -b /tmp/find.cookies -H 'Content-Type: application/json' \
  -d '{"body":"Могу описать примету подробнее"}' \
  "http://127.0.0.1:8080/api/dialogs/<dialog-id>/messages"
```

Дальше: верификация перед перепиской и статусы — [09-chat.md](09-chat.md). Встреча — [10-meeting.md](10-meeting.md).
