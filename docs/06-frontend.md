# Фронтенд (`web/`)

## Стек

- React 19 + TypeScript
- Vite 8, порт **5173**, proxy `/api` → `http://127.0.0.1:8080`
- Яндекс.Карты **JavaScript API 2.1** (ключ в `index.html`; API 3.0 с этим ключом не работает)

## Точки входа

```
index.html
  → script api-maps.yandex.ru/2.1/?apikey=…&lang=ru_RU
  → /src/main.tsx
       → App.tsx → MapPage
```

Стили: один файл `src/map/MapPage.css` (карта, меню, sheet, фильтры, composer, карточка).

## Модули карты

| Файл | Зачем |
|---|---|
| `MapPage.tsx` | Состояние фильтров, entries из API, сессия, foundOpen; связка карты и Sheet |
| `Sheet.tsx` | Левая панель: список / фильтры / карточка / форма «Я нашел» |
| `FoundForm.tsx` | Auth + создание находки |
| `ObjectCard.tsx` | Полная карточка: описание, факты, SVG-фото |
| `Menu.tsx` | Профиль / Чаты / Подписки (без роутинга) |
| `entries.ts` | Типы `MapEntry`, справочники категорий/подкатегорий, пресеты маркеров, `matchesTime` / `isInside` |
| `api.ts` | Клиент `/api/*` |
| `useYandexMap.ts` | Создание карты, bounds, zoom control, refs маркеров |
| `ymaps.ts` | Типы ymaps 2.1, `toMapPoint`, `boundsFromMap` |
| `domain.ts` | Зеркало домена (справочно) |

## Координаты

- В данных приложения и API: **`[lng, lat]`**.
- В Яндекс API 2.1: **`[lat, lng]`**.
- Конвертация только на границе `toMapPoint` / разборе bounds.

## MapPage — состояние

- `categoriesOn`, `subcategoriesOn` — множества (по умолчанию всё включено)
- `from` / `to` — строки `datetime-local`
- `selectedId` / `openedId` — маркер и открытая карточка
- `entries` — массив с сервера (не захардкоженный список)
- `user` — сессия или `null`
- `foundOpen` — режим формы находки
- `catalogReady` / `catalogError` — загрузка `/api/entries`

При монтировании:

```ts
Promise.all([loadEntries(), loadMe()])
```

`publish(draft)` берёт центр из `bounds`, вызывает `createFound`, prepend в `entries`, закрывает форму, `select(entry.id)`.

Фильтрация видимого:

```
inView = entries ∩ bounds
visible = inView ∩ category ∩ subcategory ∩ matchesTime
```

`matchesTime`: если нет time caption или kind=`place` → всегда true.

## Sheet — режимы панели

Взаимоисключающие режимы контента:

1. **Список** + футер «Я нашел»
2. **Фильтры** (иконка у заголовка)
3. **ObjectCard** (клик по строке/маркеру)
4. **FoundForm** (кнопка «Я нашел»)

Заголовок: «В области» | название объекта | «Я нашел». Назад сбрасывает opened или foundOpen.

Дерево категорий в фильтрах: checkbox категории; если checked — вложенный `ul.subcats` с checkbox подкатегорий. Снятие категории в `MapPage.onToggleCategory` также удаляет связанные subcategory id из множества.

Подпись времени: `timeCaption(categoriesOn)` → «Время пропажи» / «Время проведения» / «Время пропажи/проведения» / `null`.

## FoundForm

Без `user`: регистрация (логин, имя на карте, пароль) или вход.  
С `user`: дерево radio подкатегорий только для lossKinds (item/person/animal), поля названия, описания, места, времени, опционально пол и примета. Подсказка: точка = центр карты.

## ObjectCard

Фото не качаются с сервера: `photoCount` раз рисуется data-URI SVG с цветом категории. Факты: категория+подкатегория, «Кто нашёл/ищет/отметил» (автор **этого** объявления, не роль аккаунта), тип объявления, время, место, traits.

На `intent === found`, если пост не ваш — кнопка **«Это моё»** → claim → панель чатов.

## useYandexMap

1. Ждёт `window.ymaps` до ~8 с (poll 50 ms).
2. `ymaps.ready` → `new ymaps.Map` (центр `[55.7558, 37.6173]`, zoom 13, `controls: []`, `suppressMapOpenBlock: true`).
3. ZoomControl слева от панели (`left: 392` на широком экране).
4. `boundschange` батчится через `requestAnimationFrame` → React state `bounds`.
5. Cleanup при unmount (Strict Mode safe): destroy map, снять resize listener.

CSS скрываетcopyright-полосу: `.map-canvas [class*="copyrights-pane"] { display: none !important; }`.

Маркеры создаёт **MapPage** (не хук): добавляет/убирает Placemark по `visible`, клик → `onSelectRef.current(id)`, preset меняется при selection.

## api.ts

`fetch` с `Content-Type: application/json`. Cookie уходит автоматически (same-origin через proxy).  
`loadMe`: `401` → `null`, не throw. Остальные ошибки: throw `Error(body.error)`.

`at` с сервера (ISO) → `Date.getTime()` для фильтров и форматирования.

На карте в Entry есть поля авторства: `author`, `authorUserId`, `authorUsername`.
Если автор скрыт (seeking с internal), эти поля пустые / null, в списке пишется «Автор скрыт».
Клик по имени нашедшего/автора на карточке открывает правую панель профиля.

### Профиль (`ProfilePanel`)

Меню «Профиль» / «Войти» справа сверху:

- без сессии — регистрация или вход;
- со своим аккаунтом — имя, `@username`, статус личности, bio, редактирование, выход;
- лента объявлений этого пользователя с **тем же деревом категорий/подкатегорий**, что и на карте;
- чужой профиль открывается по клику на автора в `ObjectCard`.

API: `GET /api/users/{username}`, `GET /api/users/{username}/entries`.
Владелец в своей ленте видит и скрытые «ищу»; чужие — только с публичным авторством.

### Чаты (`web/src/chat`)

Отдельный UI-блок (не часть карты). Меню «Чаты» или переход после отклика:

- без сессии — вход/регистрация;
- список диалогов (+ статус верификации);
- до `open` — ответ на отличительную особенность / подтверждение автором;
- после `open` — сообщения.

Роль `author` / `respondent` — только сторона **в этом** чате. Подробности: [09-chat.md](09-chat.md).
