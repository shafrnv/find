# Домен (`internal/domain`)

Пакет без внешних зависимостей кроме стандартной библиотеки. Здесь живут **инварианты**: что можно создать, что видно снаружи, что считается пропажей.

Описание пакета — в `doc.go`. Зеркало публичных типов для TypeScript: `web/src/domain.ts` (пока не все поля UI использует напрямую; карта работает через `MapEntry`).

## Идея: субъект ≠ объявление

| Понятие | Смысл |
|---|---|
| **Subject** | То, о чём говорят: вещь, человек, животное, место, событие |
| **Publication** | Авторское высказывание о субъекте (ищу / нашёл / отмечаю) |
| **User** | Аккаунт. Пропавший человек — это `Person`, не `User` |

Место и время потери/находки лежат на **объявлении**, не на предмете: у одной вещи могут быть «потерял у метро» и «нашли у парка».

## Файлы и сущности

| Файл | Содержание |
|---|---|
| `id.go` | `ID` как UUID-строка, `ParseID`, `IsZero` |
| `visibility.go` | `public` / `internal` |
| `geo.go` | `GeoPoint` lat/lon с проверкой диапазонов |
| `user.go` | `User`, `IdentityVerification`, `PublicUser`, `NewUser` |
| `taxonomy.go` | `Category` (scope = вид субъекта, один уровень parent→child) |
| `media.go` | `Media`, `PhotoRef` |
| `trait.go` | Признаки; `SecretDistinctive` → internal |
| `subject.go` | `SubjectKind`, `LostFound()`, интерфейс `Subject` |
| `item.go` / `person.go` / `animal.go` / `place.go` / `event.go` | Субъекты + `Public()` views |
| `publication.go` | Intent, статус, `RevealAuthor`, `Public()` |
| `follow.go` | Подписка на user / place / event |
| `match.go` | Связка seeking + found |
| `domain_test.go` | Юнит-тесты инвариантов |

## SubjectKind

```
item | person | animal | place | event
```

`LostFound()` истинно только для `item`, `person`, `animal`. Для place/event допустим только intent `marking` при создании публикации.

## Intent объявления

| Intent | Смысл на **этом** посте | Видимость автора по умолчанию |
|---|---|---|
| `seeking` | Автор ищет / потерял | `internal` (скрыт, пока не `RevealAuthor`) |
| `found` | Автор нашёл / видел | `public` |
| `marking` | Автор отметил место или событие | `public` |

У **пользователя** постоянной роли нет: один аккаунт может публиковать все три вида объявлений. Отклик на чужую находку («Это моё») и чат — [08-dialogs.md](08-dialogs.md).

Статусы публикации: `draft` → `published` | `matched` | `closed` | `hidden`.

## Visibility

- **public** — в поиске и публичной карточке.
- **internal** — владелец вещи, скрытая примета, скрытый автор «ищу».

Методы `Public()` / `PublicUser` отрезают внутреннее. Например `Item.Public()` не отдаёт `OwnerUserID`; `Publication.Public()` не отдаёт `AuthorUserID`, если автор internal.

## Категории

`Category.Scope` совпадает с `SubjectKind`. Подкатегория — узел с `ParentID`. `ValidChildOf` требует тот же scope и родителя без собственного parent (вложенность один уровень).

В UI сейчас slug’и: `keys`, `electronics`, `bags`, `adults`, `children`, `cats`, `dogs`, `yards`, `viewpoints`, `gatherings`, `meetings`. В БД у корней slug = scope (`item`, `person`, …).

## Traits (канонические ключи)

`height_cm`, `weight_kg`, `nationality`, `appearance`, `distinctive`, `color`, `age`. Неизвестный ключ допустим; константы фиксируют смысл. `SecretDistinctive` создаёт `distinctive` с visibility internal.

## Как домен связан с остальным

- `store.Register` → `domain.NewUser`.
- `store.CreateFound` → `NewItem` / `NewPerson` / `NewAnimal`, `NewTrait`, `NewGeoPoint`, `NewPublication`, `WithStatus(published)`.
- HTTP и SQL **не** вызываются из domain.
- Тесты: `go test ./internal/domain`.
