import { useEffect, useState } from "react"
import type { SubjectKind } from "../domain"
import type { SessionUser } from "./api"
import { categories, categoryColor, categoryLabel, intents, subcategories, timeCaption, type MapEntry } from "./entries"
import { FoundForm } from "./FoundForm"
import { ObjectCard } from "./ObjectCard"

interface SheetProps {
  categoriesOn: ReadonlySet<SubjectKind>
  subcategoriesOn: ReadonlySet<string>
  from: string
  to: string
  visible: MapEntry[]
  inView: number
  boundsReady: boolean
  catalogReady: boolean
  catalogError: string | null
  error: string | null
  selectedId: string | null
  opened: MapEntry | null
  onClose: () => void
  onToggleCategory: (kind: SubjectKind) => void
  onToggleSubcategory: (id: string) => void
  onFrom: (value: string) => void
  onTo: (value: string) => void
  onFocus: (entry: MapEntry) => void
  foundOpen: boolean
  user: SessionUser | null
  onFound: () => void
  onCloseFound: () => void
  onLogin: (username: string, password: string) => Promise<void>
  onRegister: (username: string, displayName: string, password: string) => Promise<void>
  onLogout: () => Promise<void>
  onCreate: (draft: {
    kind: string
    subcategory: string
    title: string
    description: string
    place: string
    at: string
    sex: string
    trait: string
    secretTrait: string
  }) => Promise<void>
  onAuthor: (username: string) => void
  onClaim: () => Promise<void>
  onNeedAuth: () => void
}

export function Sheet({
  categoriesOn,
  subcategoriesOn,
  from,
  to,
  visible,
  inView,
  boundsReady,
  catalogReady,
  catalogError,
  error,
  selectedId,
  opened,
  onClose,
  onToggleCategory,
  onToggleSubcategory,
  onFrom,
  onTo,
  onFocus,
  foundOpen,
  user,
  onFound,
  onCloseFound,
  onLogin,
  onRegister,
  onLogout,
  onCreate,
  onAuthor,
  onClaim,
  onNeedAuth,
}: SheetProps) {
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [categoryOpen, setCategoryOpen] = useState(false)

  useEffect(() => {
    if (opened || foundOpen) setFiltersOpen(false)
  }, [opened, foundOpen])
  const count = error ? "Карта не открылась" : boundsReady ? `${visible.length} из ${inView}` : "Ждём карту"
  const caption = timeCaption(categoriesOn)
  const pickedCategories = categories.filter((category) => categoriesOn.has(category.id))
  const categorySummary = pickedCategories.length === 0 ? "Не выбрано" : pickedCategories.length === categories.length ? "Все" : pickedCategories.map((category) => category.chip).join(", ")

  return (
    <aside className="sheet" aria-label="Что есть на карте в этой области">
      <header className="sheet-head">
        {opened || foundOpen ? (
          <div className="sheet-title-row">
            <button type="button" className="filter-toggle" aria-label="К списку" onClick={foundOpen ? onCloseFound : onClose}>
              <BackIcon />
            </button>
            <h1>{foundOpen ? "Я нашел" : opened?.title}</h1>
          </div>
        ) : (
          <>
            <div className="sheet-title-row">
              <h1>В области</h1>
              <button
                type="button"
                className="filter-toggle"
                aria-label="Фильтры"
                aria-expanded={filtersOpen}
                onClick={() => {
                  setFiltersOpen((open) => !open)
                  setCategoryOpen(false)
                }}
              >
                <FilterIcon />
              </button>
            </div>
            <p>{count}</p>
          </>
        )}
      </header>

      {opened && !foundOpen ? (
        <ObjectCard entry={opened} user={user} onAuthor={onAuthor} onClaim={onClaim} onNeedAuth={onNeedAuth} />
      ) : null}

      {!opened && !foundOpen && filtersOpen ? (
        <div className="filters">
          <div className="field">
            <span className="field-label">Категория</span>
            <button type="button" className="select-button" aria-expanded={categoryOpen} onClick={() => setCategoryOpen((open) => !open)}>
              <span>{categorySummary}</span>
            </button>
            {categoryOpen ? (
              <ul className="select-menu">
                {categories.map((category) => {
                  const children = subcategories.filter((item) => item.categoryId === category.id)
                  const checked = categoriesOn.has(category.id)
                  return (
                    <li key={category.id}>
                      <label>
                        <input type="checkbox" checked={checked} onChange={() => onToggleCategory(category.id)} />
                        {category.chip}
                      </label>
                      {checked ? (
                        <ul className="subcats">
                          {children.map((item) => (
                            <li key={item.id}>
                              <label>
                                <input type="checkbox" checked={subcategoriesOn.has(item.id)} onChange={() => onToggleSubcategory(item.id)} />
                                {item.label}
                              </label>
                            </li>
                          ))}
                        </ul>
                      ) : null}
                    </li>
                  )
                })}
              </ul>
            ) : null}
          </div>
          {caption ? (
            <fieldset className="time-range">
              <legend>{caption}</legend>
              <label>
                От
                <input type="datetime-local" value={from} onChange={(event) => onFrom(event.target.value)} />
              </label>
              <label>
                До
                <input type="datetime-local" value={to} onChange={(event) => onTo(event.target.value)} />
              </label>
            </fieldset>
          ) : null}
        </div>
      ) : null}

      {foundOpen ? (
        <FoundForm user={user} onLogin={onLogin} onRegister={onRegister} onLogout={onLogout} onCreate={onCreate} />
      ) : null}

      {opened || filtersOpen || foundOpen ? null : (
        <div className="list">
          {error ? <p className="empty">{error}</p> : null}
          {catalogError ? <p className="empty">{catalogError}</p> : null}
          {!error && !boundsReady ? <p className="empty">Открываем карту</p> : null}
          {!error && boundsReady && !catalogReady ? <p className="empty">Открываем объявления</p> : null}
          {!error && !catalogError && boundsReady && catalogReady && visible.length === 0 ? (
            <p className="empty">В этой области ничего нет по выбранным фильтрам</p>
          ) : null}
          {!error &&
            visible.map((entry) => (
              <button
                key={entry.id}
                type="button"
                className={entry.id === selectedId ? "row is-selected" : "row"}
                onClick={() => onFocus(entry)}
              >
                <span className="dot" style={{ background: categoryColor(entry.kind) }} />
                <span>
                  <strong>{entry.title}</strong>
                  <span className="meta">
                    {intents[entry.intent]} · {categoryLabel(entry.kind)} · {entry.author || "Автор скрыт"} · {formatWhen(entry.at)}
                  </span>
                  <span className="place">{entry.place}</span>
                </span>
              </button>
            ))}
        </div>
      )}

      {opened || filtersOpen || foundOpen ? null : (
        <footer className="sheet-foot">
          <button type="button" className="found-button" onClick={onFound}>
            Я нашел
          </button>
        </footer>
      )}
    </aside>
  )
}

function BackIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
      <path d="M14.5 6.5 9 12l5.5 5.5" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function FilterIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
      <path d="M4 6h16M7 12h10M10 18h4" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  )
}

function formatWhen(at: number): string {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(at)
}
