import type { MapEntry, TraitLine } from "./entries"

export interface SessionUser {
  id: string
  username: string
  displayName: string
  bio: string
  identityStatus: string
}

export interface Profile {
  id: string
  username: string
  displayName: string
  bio: string
  identityStatus: string
  createdAt: string
  entryCount: number
  mine: boolean
}

export interface FoundDraft {
  kind: string
  subcategory: string
  title: string
  description: string
  place: string
  at: string
  sex: string
  trait: string
  secretTrait: string
  lat: number
  lon: number
}

interface EntryDTO extends Omit<MapEntry, "at" | "traits"> {
  at: string
  traits: TraitLine[] | null
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
  })
  const text = await response.text()
  const body = text ? (JSON.parse(text) as T & { error?: string }) : ({} as T & { error?: string })
  if (!response.ok) throw new Error(body.error || "Не получилось")
  return body
}

function toEntry(dto: EntryDTO): MapEntry {
  return {
    ...dto,
    authorUserId: dto.authorUserId ?? null,
    authorUsername: dto.authorUsername ?? null,
    at: new Date(dto.at).getTime(),
    traits: dto.traits ?? [],
  }
}

export function loadEntries(): Promise<MapEntry[]> {
  return request<EntryDTO[]>("/api/entries").then((list) => list.map(toEntry))
}

export async function loadMe(): Promise<SessionUser | null> {
  const response = await fetch("/api/me")
  if (response.status === 401) return null
  const body = (await response.json()) as SessionUser & { error?: string }
  if (!response.ok) throw new Error(body.error || "Не получилось")
  return body
}

export function register(username: string, displayName: string, password: string): Promise<SessionUser> {
  return request("/api/register", { method: "POST", body: JSON.stringify({ username, displayName, password }) })
}

export function login(username: string, password: string): Promise<SessionUser> {
  return request("/api/login", { method: "POST", body: JSON.stringify({ username, password }) })
}

export function logout(): Promise<void> {
  return request("/api/logout", { method: "POST", body: "{}" })
}

export function updateProfile(displayName: string, bio: string): Promise<SessionUser> {
  return request("/api/me", { method: "PATCH", body: JSON.stringify({ displayName, bio }) })
}

export function loadProfile(username: string): Promise<Profile> {
  return request(`/api/users/${encodeURIComponent(username)}`)
}

export function loadUserEntries(username: string): Promise<MapEntry[]> {
  return request<EntryDTO[]>(`/api/users/${encodeURIComponent(username)}/entries`).then((list) => list.map(toEntry))
}

export function createFound(draft: FoundDraft): Promise<MapEntry> {
  return request<EntryDTO>("/api/found", { method: "POST", body: JSON.stringify(draft) }).then(toEntry)
}
