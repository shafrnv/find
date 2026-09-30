import { useEffect, useMemo, useState, type FormEvent } from "react"
import type { SubjectKind } from "../domain"
import { AuthForm } from "./AuthForm"
import { loadProfile, loadUserEntries, updateProfile, type Profile, type SessionUser } from "./api"
import { categories, categoryColor, categoryLabel, intents, subcategories, type MapEntry } from "./entries"

interface ProfilePanelProps {
  open: boolean
  username: string | null
  user: SessionUser | null
  onClose: () => void
  onLogin: (username: string, password: string) => Promise<void>
  onRegister: (username: string, displayName: string, password: string) => Promise<void>
  onLogout: () => Promise<void>
  onUserChange: (user: SessionUser) => void
  onOpenEntry: (entry: MapEntry) => void
}

export function ProfilePanel({
  open,
  username,
  user,
  onClose,
  onLogin,
  onRegister,
  onLogout,
  onUserChange,
  onOpenEntry,
}: ProfilePanelProps) {
  const target = username ?? user?.username ?? null
  const needAuth = open && !target

  if (!open) return null

  return (
    <aside className="profile-panel" aria-label="Профиль">
      <header className="sheet-head">
        <div className="sheet-title-row">
          <button type="button" className="filter-toggle" aria-label="Закрыть профиль" onClick={onClose}>
            <CloseIcon />
          </button>
          <h1>Профиль</h1>
        </div>
      </header>
      {needAuth ? (
        <AuthForm
          hint="Войдите или зарегистрируйтесь — после этого откроется ваш профиль и лента объявлений."
          onLogin={onLogin}
          onRegister={onRegister}
        />
      ) : target ? (
        <ProfileBody
          username={target}
          user={user}
          onLogout={onLogout}
          onUserChange={onUserChange}
          onOpenEntry={onOpenEntry}
        />
      ) : null}
    </aside>
  )
}

function ProfileBody({
  username,
  user,
  onLogout,
  onUserChange,
  onOpenEntry,
}: {
  username: string
  user: SessionUser | null
  onLogout: () => Promise<void>
  onUserChange: (user: SessionUser) => void
  onOpenEntry: (entry: MapEntry) => void
}) {
  const [profile, setProfile] = useState<Profile | null>(null)
  const [entries, setEntries] = useState<MapEntry[]>([])
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(true)
  const [categoriesOn, setCategoriesOn] = useState<ReadonlySet<SubjectKind>>(() => new Set(categories.map((c) => c.id)))
  const [subcategoriesOn, setSubcategoriesOn] = useState<ReadonlySet<string>>(() => new Set(subcategories.map((s) => s.id)))
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [editing, setEditing] = useState(false)

  useEffect(() => {
    let cancel = false
    setLoading(true)
    setError("")
    Promise.all([loadProfile(username), loadUserEntries(username)])
      .then(([nextProfile, list]) => {
        if (cancel) return
        setProfile(nextProfile)
        setEntries(list)
        setLoading(false)
      })
      .catch((err: unknown) => {
        if (cancel) return
        setError(err instanceof Error ? err.message : "Не получилось")
        setLoading(false)
      })
    return () => {
      cancel = true
    }
  }, [username, user?.id])

  const visible = useMemo(
    () => entries.filter((entry) => categoriesOn.has(entry.kind) && subcategoriesOn.has(entry.subcategoryId)),
    [entries, categoriesOn, subcategoriesOn],
  )

  function toggleCategory(kind: SubjectKind) {
    setCategoriesOn((current) => {
      const next = new Set(current)
      if (next.has(kind)) next.delete(kind)
      else next.add(kind)
      return next
    })
    setSubcategoriesOn((current) => {
      const next = new Set(current)
      const related = subcategories.filter((item) => item.categoryId === kind).map((item) => item.id)
      if (categoriesOn.has(kind)) related.forEach((id) => next.delete(id))
      else related.forEach((id) => next.add(id))
      return next
    })
  }

  return (
    <div className="profile-body">
      {loading ? <p className="empty">Открываем профиль</p> : null}
      {error ? <p className="form-error">{error}</p> : null}
      {profile ? (
        <>
          <div className="profile-card">
            <strong>{profile.displayName}</strong>
            <span className="profile-handle">@{profile.username}</span>
            <span className="profile-meta">
              {identityLabel(profile.identityStatus)} · {profile.entryCount}{" "}
              {plural(profile.entryCount, "объявление", "объявления", "объявлений")}
            </span>
            {profile.bio ? <p className="profile-bio">{profile.bio}</p> : <p className="profile-bio is-empty">Пока без описания</p>}
            {profile.mine ? (
              <div className="profile-actions">
                <button type="button" className="text-button" onClick={() => setEditing((v) => !v)}>
                  {editing ? "Скрыть редактирование" : "Редактировать"}
                </button>
                <button type="button" className="text-button" onClick={() => void onLogout()}>
                  Выйти
                </button>
              </div>
            ) : null}
          </div>
          {profile.mine && editing ? (
            <EditProfile
              displayName={profile.displayName}
              bio={profile.bio}
              onSave={async (displayName, bio) => {
                const updated = await updateProfile(displayName, bio)
                onUserChange(updated)
                setProfile({ ...profile, displayName: updated.displayName, bio: updated.bio })
                setEditing(false)
              }}
            />
          ) : null}
          <div className="profile-feed-head">
            <h2>Лента</h2>
            <button type="button" className="filter-toggle" aria-label="Фильтры ленты" aria-expanded={filtersOpen} onClick={() => setFiltersOpen((v) => !v)}>
              <FilterIcon />
            </button>
            <span>
              {visible.length} из {entries.length}
            </span>
          </div>
          {filtersOpen ? (
            <div className="filters profile-filters">
              <div className="field">
                <span className="field-label">Категория</span>
                <ul className="select-menu">
                  {categories.map((category) => {
                    const children = subcategories.filter((item) => item.categoryId === category.id)
                    const checked = categoriesOn.has(category.id)
                    return (
                      <li key={category.id}>
                        <label>
                          <input type="checkbox" checked={checked} onChange={() => toggleCategory(category.id)} />
                          {category.chip}
                        </label>
                        {checked ? (
                          <ul className="subcats">
                            {children.map((item) => (
                              <li key={item.id}>
                                <label>
                                  <input
                                    type="checkbox"
                                    checked={subcategoriesOn.has(item.id)}
                                    onChange={() =>
                                      setSubcategoriesOn((current) => {
                                        const next = new Set(current)
                                        if (next.has(item.id)) next.delete(item.id)
                                        else next.add(item.id)
                                        return next
                                      })
                                    }
                                  />
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
              </div>
            </div>
          ) : null}
          <div className="list profile-list">
            {visible.length === 0 ? <p className="empty">В этой ленте пока пусто по выбранным категориям</p> : null}
            {visible.map((entry) => (
              <button key={entry.id} type="button" className="row" onClick={() => onOpenEntry(entry)}>
                <span className="dot" style={{ background: categoryColor(entry.kind) }} />
                <span>
                  <strong>{entry.title}</strong>
                  <span className="meta">
                    {intents[entry.intent]} · {categoryLabel(entry.kind)} · {formatWhen(entry.at)}
                  </span>
                  <span className="place">{entry.place}</span>
                </span>
              </button>
            ))}
          </div>
        </>
      ) : null}
    </div>
  )
}

function EditProfile({
  displayName,
  bio,
  onSave,
}: {
  displayName: string
  bio: string
  onSave: (displayName: string, bio: string) => Promise<void>
}) {
  const [name, setName] = useState(displayName)
  const [about, setAbout] = useState(bio)
  const [error, setError] = useState("")
  const [busy, setBusy] = useState(false)

  async function submit(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError("")
    try {
      await onSave(name, about)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не получилось")
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="composer profile-edit" onSubmit={submit}>
      <label>
        Имя на карте
        <input value={name} onChange={(event) => setName(event.target.value)} />
      </label>
      <label>
        О себе
        <textarea value={about} onChange={(event) => setAbout(event.target.value)} />
      </label>
      {error ? <p className="form-error">{error}</p> : null}
      <button type="submit" className="found-button" disabled={busy}>
        Сохранить
      </button>
    </form>
  )
}

function identityLabel(status: string): string {
  if (status === "verified") return "Личность подтверждена"
  if (status === "pending") return "Проверка личности"
  return "Без подтверждения"
}

function plural(n: number, one: string, few: string, many: string): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return few
  return many
}

function formatWhen(at: number): string {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(at)
}

function CloseIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
      <path d="M7 7l10 10M17 7 7 17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
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
