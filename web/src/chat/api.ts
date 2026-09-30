import type { SessionUser } from "../map/api"

export type DialogStatus = "pending_answer" | "pending_confirm" | "open" | "rejected" | "closed"

export interface DialogSummary {
  id: string
  publicationId: string
  title: string
  place: string
  kind?: string
  intent?: string
  peerName: string
  peerUsername: string
  role: "author" | "respondent"
  status: DialogStatus
  lastBody: string
  updatedAt: string
}

export interface ChatMessage {
  id: string
  senderId: string
  mine: boolean
  body: string
  createdAt: string
}

export interface DialogDetail {
  id: string
  publicationId: string
  title: string
  place: string
  kind?: string
  intent?: string
  peerName: string
  peerUsername: string
  role: "author" | "respondent"
  status: DialogStatus
  hasChallenge: boolean
  respondentAnswer?: string
  canMessage: boolean
  canAnswer: boolean
  canConfirm: boolean
  createdAt: string
  messages: ChatMessage[]
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

export function claimFound(publicationId: string): Promise<DialogDetail> {
  return request(`/api/entries/${encodeURIComponent(publicationId)}/claim`, { method: "POST", body: "{}" })
}

export function loadDialogs(): Promise<DialogSummary[]> {
  return request("/api/dialogs")
}

export function loadDialog(id: string): Promise<DialogDetail> {
  return request(`/api/dialogs/${encodeURIComponent(id)}`)
}

export function sendDialogMessage(id: string, body: string): Promise<ChatMessage> {
  return request(`/api/dialogs/${encodeURIComponent(id)}/messages`, { method: "POST", body: JSON.stringify({ body }) })
}

export function submitDialogAnswer(id: string, answer: string): Promise<DialogDetail> {
  return request(`/api/dialogs/${encodeURIComponent(id)}/answer`, { method: "POST", body: JSON.stringify({ answer }) })
}

export function confirmDialogAnswer(id: string, accept: boolean): Promise<DialogDetail> {
  return request(`/api/dialogs/${encodeURIComponent(id)}/confirm`, { method: "POST", body: JSON.stringify({ accept }) })
}

export type { SessionUser }
