import { useCallback, useEffect, useMemo, useState } from "react"
import type { SubjectKind } from "../domain"
import { ChatsPanel, claimFound } from "../chat"
import { createFound, loadEntries, loadMe, login, logout, register, type FoundDraft, type SessionUser } from "./api"
import { Menu } from "./Menu"
import { ProfilePanel } from "./ProfilePanel"
import { Sheet } from "./Sheet"
import { categories, categoryPreset, categoryPresetSelected, isInside, matchesTime, subcategories, timeCaption, type MapEntry } from "./entries"
import { useYandexMap } from "./useYandexMap"
import { toMapPoint } from "./ymaps"

function toggle<T>(current: ReadonlySet<T>, value: T): Set<T> {
  const next = new Set(current)
  if (next.has(value)) next.delete(value)
  else next.add(value)
  return next
}

function parseLocal(value: string): number | null {
  if (!value) return null
  const time = new Date(value).getTime()
  return Number.isNaN(time) ? null : time
}

export function MapPage() {
  const [categoriesOn, setCategoriesOn] = useState<ReadonlySet<SubjectKind>>(() => new Set(categories.map((category) => category.id)))
  const [subcategoriesOn, setSubcategoriesOn] = useState<ReadonlySet<string>>(() => new Set(subcategories.map((item) => item.id)))
  const [from, setFrom] = useState("")
  const [to, setTo] = useState("")
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [openedId, setOpenedId] = useState<string | null>(null)
  const [entries, setEntries] = useState<MapEntry[]>([])
  const [catalogReady, setCatalogReady] = useState(false)
  const [catalogError, setCatalogError] = useState<string | null>(null)
  const [user, setUser] = useState<SessionUser | null>(null)
  const [foundOpen, setFoundOpen] = useState(false)
  const [profileOpen, setProfileOpen] = useState(false)
  const [profileUsername, setProfileUsername] = useState<string | null>(null)
  const [chatsOpen, setChatsOpen] = useState(false)
  const [dialogId, setDialogId] = useState<string | null>(null)

  useEffect(() => {
    let cancel = false
    Promise.all([loadEntries(), loadMe()])
      .then(([list, account]) => {
        if (cancel) return
        setEntries(list)
        setUser(account)
        setCatalogReady(true)
      })
      .catch((err: unknown) => {
        if (cancel) return
        setCatalogError(err instanceof Error ? err.message : "Не удалось открыть объявления")
        setCatalogReady(true)
      })
    return () => {
      cancel = true
    }
  }, [])

  const select = useCallback((id: string) => {
    setSelectedId(id)
    setOpenedId(id)
  }, [])
  const { containerRef, mapRef, markersRef, zoomRef, onSelectRef, bounds, ready, error } = useYandexMap(select)

  const inView = useMemo(() => (bounds ? entries.filter((entry) => isInside(entry, bounds)) : []), [bounds, entries])
  const visible = useMemo(() => {
    const range = { from: parseLocal(from), to: parseLocal(to) }
    const caption = timeCaption(categoriesOn)
    return inView.filter(
      (entry) => categoriesOn.has(entry.kind) && subcategoriesOn.has(entry.subcategoryId) && matchesTime(entry, range, caption),
    )
  }, [inView, categoriesOn, subcategoriesOn, from, to])

  useEffect(() => {
    const map = mapRef.current
    const ymaps = window.ymaps
    if (!ready || !map || !ymaps) return

    const ids = new Set(visible.map((entry) => entry.id))
    for (const [id, placemark] of markersRef.current) {
      if (!ids.has(id)) {
        map.geoObjects.remove(placemark)
        markersRef.current.delete(id)
      }
    }

    for (const entry of visible) {
      let placemark = markersRef.current.get(entry.id)
      if (!placemark) {
        placemark = new ymaps.Placemark(
          toMapPoint(entry.coordinates),
          { hintContent: entry.title },
          { preset: categoryPreset[entry.kind] },
        )
        placemark.events.add("click", () => onSelectRef.current(entry.id))
        map.geoObjects.add(placemark)
        markersRef.current.set(entry.id, placemark)
      }
      placemark.options.set("preset", entry.id === selectedId ? categoryPresetSelected[entry.kind] : categoryPreset[entry.kind])
    }
  }, [ready, visible, selectedId, mapRef, markersRef, onSelectRef])

  useEffect(() => {
    if (!openedId) return
    const entry = entries.find((item) => item.id === openedId)
    if (!entry) return
    mapRef.current?.setCenter(toMapPoint(entry.coordinates), Math.max(zoomRef.current, 14), { duration: 250 })
  }, [openedId, entries, mapRef, zoomRef])

  const opened = entries.find((item) => item.id === openedId) ?? null

  function openOwnProfile() {
    setProfileUsername(user?.username ?? null)
    setProfileOpen(true)
    setFoundOpen(false)
    setChatsOpen(false)
  }

  function openAuthor(username: string) {
    setProfileUsername(username)
    setProfileOpen(true)
    setFoundOpen(false)
    setChatsOpen(false)
  }

  function openChats(id: string | null = null) {
    setDialogId(id)
    setChatsOpen(true)
    setProfileOpen(false)
    setFoundOpen(false)
  }

  async function handleLogin(username: string, password: string) {
    const account = await login(username, password)
    setUser(account)
    setProfileUsername(account.username)
  }

  async function handleRegister(username: string, displayName: string, password: string) {
    const account = await register(username, displayName, password)
    setUser(account)
    setProfileUsername(account.username)
  }

  async function handleLogout() {
    await logout()
    setUser(null)
    setProfileUsername(null)
    setProfileOpen(false)
  }

  async function publish(draft: Omit<FoundDraft, "lat" | "lon">) {
    if (!bounds) throw new Error("Карта ещё не готова")
    const [[west, south], [east, north]] = bounds
    const entry = await createFound({
      ...draft,
      lon: (Math.min(west, east) + Math.max(west, east)) / 2,
      lat: (Math.min(south, north) + Math.max(south, north)) / 2,
    })
    setEntries((current) => [entry, ...current])
    setFoundOpen(false)
    select(entry.id)
  }

  return (
    <>
      <div ref={containerRef} className="map-canvas" />
      <Menu user={user} onProfile={openOwnProfile} onChats={() => openChats()} />
      <Sheet
        categoriesOn={categoriesOn}
        subcategoriesOn={subcategoriesOn}
        from={from}
        to={to}
        visible={visible}
        inView={inView.length}
        boundsReady={bounds != null}
        catalogReady={catalogReady}
        catalogError={catalogError}
        error={error}
        selectedId={selectedId}
        opened={opened}
        onClose={() => setOpenedId(null)}
        onToggleCategory={(kind) => {
          setCategoriesOn((current) => toggle(current, kind))
          setSubcategoriesOn((current) => {
            const next = new Set(current)
            const related = subcategories.filter((item) => item.categoryId === kind).map((item) => item.id)
            if (categoriesOn.has(kind)) related.forEach((id) => next.delete(id))
            else related.forEach((id) => next.add(id))
            return next
          })
        }}
        onToggleSubcategory={(id) => setSubcategoriesOn((current) => toggle(current, id))}
        onFrom={setFrom}
        onTo={setTo}
        onFocus={(entry) => select(entry.id)}
        foundOpen={foundOpen}
        user={user}
        onFound={() => {
          setFoundOpen(true)
          setProfileOpen(false)
          setChatsOpen(false)
        }}
        onCloseFound={() => setFoundOpen(false)}
        onLogin={handleLogin}
        onRegister={handleRegister}
        onLogout={handleLogout}
        onCreate={(draft) => publish(draft)}
        onAuthor={openAuthor}
        onClaim={async () => {
          if (!opened) return
          const dialog = await claimFound(opened.id)
          openChats(dialog.id)
        }}
        onNeedAuth={() => {
          setProfileUsername(null)
          setProfileOpen(true)
          setChatsOpen(false)
        }}
      />
      <ProfilePanel
        open={profileOpen}
        username={profileUsername}
        user={user}
        onClose={() => {
          setProfileOpen(false)
          setProfileUsername(null)
        }}
        onLogin={handleLogin}
        onRegister={handleRegister}
        onLogout={handleLogout}
        onUserChange={setUser}
        onOpenEntry={(entry) => {
          if (!entries.some((item) => item.id === entry.id)) {
            setEntries((current) => [entry, ...current])
          }
          select(entry.id)
          setProfileOpen(false)
        }}
      />
      <ChatsPanel
        open={chatsOpen}
        dialogId={dialogId}
        user={user}
        onClose={() => {
          setChatsOpen(false)
          setDialogId(null)
        }}
        onSelectDialog={setDialogId}
        onLogin={handleLogin}
        onRegister={handleRegister}
      />
    </>
  )
}
