import { useEffect, useRef, useState, type FormEvent } from "react"
import { AuthForm } from "../map/AuthForm"
import {
  confirmDialogAnswer,
  loadDialog,
  loadDialogs,
  sendDialogMessage,
  submitDialogAnswer,
  type DialogDetail,
  type DialogStatus,
  type DialogSummary,
  type SessionUser,
} from "./api"

interface ChatsPanelProps {
  open: boolean
  dialogId: string | null
  user: SessionUser | null
  onClose: () => void
  onSelectDialog: (id: string | null) => void
  onLogin: (username: string, password: string) => Promise<void>
  onRegister: (username: string, displayName: string, password: string) => Promise<void>
}

/** Самостоятельный UI-блок чата: список, верификация по примете, переписка. */
export function ChatsPanel({ open, dialogId, user, onClose, onSelectDialog, onLogin, onRegister }: ChatsPanelProps) {
  if (!open) return null

  return (
    <aside className="profile-panel chats-panel" aria-label="Чаты">
      <header className="sheet-head">
        <div className="sheet-title-row">
          {dialogId ? (
            <button type="button" className="filter-toggle" aria-label="К списку чатов" onClick={() => onSelectDialog(null)}>
              <BackIcon />
            </button>
          ) : (
            <button type="button" className="filter-toggle" aria-label="Закрыть чаты" onClick={onClose}>
              <CloseIcon />
            </button>
          )}
          <h1>{dialogId ? "Диалог" : "Чаты"}</h1>
        </div>
      </header>
      {!user ? (
        <AuthForm
          hint="Чаты появляются после отклика на чужое объявление. Если автор задал отличительную особенность — сначала проверка, потом переписка."
          onLogin={onLogin}
          onRegister={onRegister}
        />
      ) : dialogId ? (
        <ChatView dialogId={dialogId} user={user} />
      ) : (
        <DialogList onOpen={onSelectDialog} />
      )}
    </aside>
  )
}

function DialogList({ onOpen }: { onOpen: (id: string) => void }) {
  const [list, setList] = useState<DialogSummary[]>([])
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancel = false
    setLoading(true)
    loadDialogs()
      .then((items) => {
        if (!cancel) {
          setList(items)
          setLoading(false)
        }
      })
      .catch((err: unknown) => {
        if (!cancel) {
          setError(err instanceof Error ? err.message : "Не получилось")
          setLoading(false)
        }
      })
    return () => {
      cancel = true
    }
  }, [])

  return (
    <div className="list profile-list">
      {loading ? <p className="empty">Открываем чаты</p> : null}
      {error ? <p className="form-error">{error}</p> : null}
      {!loading && !error && list.length === 0 ? (
        <p className="empty">Пока пусто. Откройте чужое объявление и откликнитесь.</p>
      ) : null}
      {list.map((item) => (
        <button key={item.id} type="button" className="row" onClick={() => onOpen(item.id)}>
          <span className="dot" style={{ background: item.role === "author" ? "#2f6fed" : "#0f766e" }} />
          <span>
            <strong>{item.title}</strong>
            <span className="meta">
              {statusLabel(item.status)} · {item.role === "author" ? "Откликнулся" : "Автор объявления"} · {item.peerName}
            </span>
            <span className="place">{item.lastBody || item.place}</span>
          </span>
        </button>
      ))}
    </div>
  )
}

function ChatView({ dialogId, user }: { dialogId: string; user: SessionUser }) {
  const [dialog, setDialog] = useState<DialogDetail | null>(null)
  const [error, setError] = useState("")
  const [text, setText] = useState("")
  const [answer, setAnswer] = useState("")
  const [busy, setBusy] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    let cancel = false
    loadDialog(dialogId)
      .then((detail) => {
        if (!cancel) setDialog(detail)
      })
      .catch((err: unknown) => {
        if (!cancel) setError(err instanceof Error ? err.message : "Не получилось")
      })
    return () => {
      cancel = true
    }
  }, [dialogId])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [dialog?.messages.length])

  async function submitMessage(event: FormEvent) {
    event.preventDefault()
    if (!text.trim()) return
    setBusy(true)
    setError("")
    try {
      const msg = await sendDialogMessage(dialogId, text)
      setDialog((current) => (current ? { ...current, messages: [...current.messages, msg] } : current))
      setText("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не получилось")
    } finally {
      setBusy(false)
    }
  }

  async function submitAnswer(event: FormEvent) {
    event.preventDefault()
    if (!answer.trim()) return
    setBusy(true)
    setError("")
    try {
      const detail = await submitDialogAnswer(dialogId, answer)
      setDialog(detail)
      setAnswer("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не получилось")
    } finally {
      setBusy(false)
    }
  }

  async function confirm(accept: boolean) {
    setBusy(true)
    setError("")
    try {
      const detail = await confirmDialogAnswer(dialogId, accept)
      setDialog(detail)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не получилось")
    } finally {
      setBusy(false)
    }
  }

  if (!dialog && !error) return <p className="empty">Открываем диалог</p>
  if (!dialog) return <p className="form-error">{error}</p>

  return (
    <div className="chat">
      <div className="chat-about">
        <strong>{dialog.title}</strong>
        <span>
          {dialog.role === "author" ? "Откликнулся" : "Автор объявления"}: {dialog.peerName} @{dialog.peerUsername}
        </span>
        {dialog.place ? <span>{dialog.place}</span> : null}
        <span className="meta">{statusLabel(dialog.status)}</span>
      </div>

      {dialog.canAnswer ? (
        <form className="composer chat-gate" onSubmit={submitAnswer}>
          <p className="hint">Автор задал отличительную особенность. Назовите её, чтобы открыть переписку.</p>
          <label>
            Отличительная особенность
            <input value={answer} onChange={(event) => setAnswer(event.target.value)} placeholder="Например: гравировка на кольце" />
          </label>
          {error ? <p className="form-error">{error}</p> : null}
          <button type="submit" className="found-button" disabled={busy}>
            Отправить ответ
          </button>
        </form>
      ) : null}

      {dialog.canConfirm ? (
        <div className="composer chat-gate">
          <p className="hint">Откликнувшийся назвал отличительную особенность. Подтвердите, если это верно.</p>
          <p>
            <strong>Ответ:</strong> {dialog.respondentAnswer || "—"}
          </p>
          {error ? <p className="form-error">{error}</p> : null}
          <div className="chat-gate-actions">
            <button type="button" className="found-button" disabled={busy} onClick={() => void confirm(true)}>
              Да, это та
            </button>
            <button type="button" className="text-button" disabled={busy} onClick={() => void confirm(false)}>
              Нет, отклонить
            </button>
          </div>
        </div>
      ) : null}

      {dialog.status === "pending_confirm" && dialog.role === "respondent" ? (
        <p className="hint">Ответ отправлен. Ждём подтверждения автора объявления.</p>
      ) : null}

      {dialog.status === "rejected" ? <p className="form-error">Отклик отклонён: примета не подтверждена.</p> : null}

      {dialog.canMessage ? (
        <>
          <div className="chat-messages">
            {dialog.messages.map((msg) => (
              <div key={msg.id} className={msg.mine || msg.senderId === user.id ? "bubble mine" : "bubble"}>
                <p>{msg.body}</p>
                <time>{formatWhen(msg.createdAt)}</time>
              </div>
            ))}
            <div ref={bottomRef} />
          </div>
          {error ? <p className="form-error">{error}</p> : null}
          <form className="chat-compose" onSubmit={submitMessage}>
            <input value={text} placeholder="Сообщение" onChange={(event) => setText(event.target.value)} />
            <button type="submit" className="found-button" disabled={busy}>
              Отправить
            </button>
          </form>
        </>
      ) : null}
    </div>
  )
}

function statusLabel(status: DialogStatus): string {
  switch (status) {
    case "pending_answer":
      return "Ждём примету"
    case "pending_confirm":
      return "Подтверждение приметы"
    case "open":
      return "Открыт"
    case "rejected":
      return "Отклонён"
    case "closed":
      return "Закрыт"
    default:
      return status
  }
}

function formatWhen(value: string): string {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value))
}

function CloseIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
      <path d="M7 7l10 10M17 7 7 17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  )
}

function BackIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
      <path d="M14.5 6.5 9 12l5.5 5.5" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}
