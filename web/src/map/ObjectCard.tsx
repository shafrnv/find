import { useState } from "react"
import type { SessionUser } from "./api"
import { categoryColor, categoryLabel, intents, subcategoryLabel, type MapEntry } from "./entries"

interface ObjectCardProps {
  entry: MapEntry
  user: SessionUser | null
  onAuthor?: (username: string) => void
  onClaim?: () => Promise<void>
  onNeedAuth?: () => void
}

export function ObjectCard({ entry, user, onAuthor, onClaim, onNeedAuth }: ObjectCardProps) {
  const [photo, setPhoto] = useState(0)
  const [claimError, setClaimError] = useState("")
  const [claimBusy, setClaimBusy] = useState(false)
  const photos = Array.from({ length: entry.photoCount }, (_, index) => photoUri(categoryColor(entry.kind), entry.title, index))
  const current = photos[photo] ?? photos[0]
  const authorClickable = Boolean(entry.authorUsername && entry.author && onAuthor)
  const isOwn = Boolean(user && entry.authorUserId && user.id === entry.authorUserId)
  const canRespond = (entry.intent === "found" || entry.intent === "seeking") && !isOwn
  const respondLabel = respondButtonLabel(entry)
  const respondHint = respondHintText(entry)

  async function claim() {
    setClaimError("")
    if (!user) {
      onNeedAuth?.()
      return
    }
    if (!onClaim) return
    setClaimBusy(true)
    try {
      await onClaim()
    } catch (err) {
      setClaimError(err instanceof Error ? err.message : "Не получилось")
    } finally {
      setClaimBusy(false)
    }
  }

  return (
    <article className="detail">
      {current ? (
        <div className="detail-gallery">
          <img src={current} alt="" />
          {photos.length > 1 ? (
            <div className="detail-thumbs">
              {photos.map((src, index) => (
                <button key={src} type="button" aria-label={`Фото ${index + 1}`} aria-pressed={index === photo} onClick={() => setPhoto(index)}>
                  <img src={src} alt="" />
                </button>
              ))}
            </div>
          ) : null}
        </div>
      ) : null}
      <p className="detail-text">{entry.description}</p>
      <dl className="facts">
        <Fact name="Категория" value={`${categoryLabel(entry.kind)}, ${subcategoryLabel(entry.subcategoryId)}`} />
        <dt>{roleName(entry)}</dt>
        <dd>
          {authorClickable ? (
            <button type="button" className="author-link" onClick={() => onAuthor?.(entry.authorUsername!)}>
              {entry.author}
              {entry.authorUsername ? <span className="author-handle"> @{entry.authorUsername}</span> : null}
            </button>
          ) : (
            entry.author || "Скрыт"
          )}
        </dd>
        <Fact name="Объявление" value={intents[entry.intent]} />
        <Fact name={whenName(entry)} value={formatWhen(entry.at)} />
        <Fact name="Место" value={entry.place} />
        {entry.traits.map((trait) => (
          <Fact key={trait.name} name={trait.name} value={trait.value} />
        ))}
      </dl>
      {canRespond ? (
        <div className="claim-box">
          <p className="hint">{respondHint}</p>
          {claimError ? <p className="form-error">{claimError}</p> : null}
          <button type="button" className="found-button" disabled={claimBusy} onClick={() => void claim()}>
            {respondLabel}
          </button>
        </div>
      ) : null}
    </article>
  )
}

function Fact({ name, value }: { name: string; value: string }) {
  return (
    <>
      <dt>{name}</dt>
      <dd>{value}</dd>
    </>
  )
}

function roleName(entry: MapEntry): string {
  if (entry.intent === "found") return "Кто нашёл"
  if (entry.intent === "seeking") return "Кто ищет"
  return "Кто отметил"
}

function respondButtonLabel(entry: MapEntry): string {
  if (entry.kind === "person") {
    return entry.intent === "found" ? "Это мой человек" : "Я видел"
  }
  if (entry.intent === "found") return "Это моё"
  return "Я нашёл"
}

function respondHintText(entry: MapEntry): string {
  if (entry.kind === "person") {
    if (entry.intent === "found") {
      return "Отклик откроет переписку с автором ориентировки. Роли у аккаунта нет — вы можете и искать людей, и откликаться на находки."
    }
    return "Если вы видели этого человека, отклик откроет диалог с тем, кто ищет."
  }
  if (entry.intent === "found") {
    return "Скажите, что это ваше — откроется диалог с автором находки."
  }
  return "Если вы нашли то, что ищут, откроется диалог с автором объявления."
}

function whenName(entry: MapEntry): string {
  if (entry.kind === "event") return "Время проведения"
  if (entry.kind === "place") return "Отмечено"
  if (entry.intent === "found") return "Время находки"
  return "Время пропажи"
}

function formatWhen(at: number): string {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    hour: "2-digit",
    minute: "2-digit",
  }).format(at)
}

function photoUri(color: string, title: string, variant: number): string {
  const shift = 40 + variant * 70
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="640" height="400" viewBox="0 0 640 400">
    <rect width="640" height="400" fill="${color}"/>
    <rect y="250" width="640" height="150" fill="#000" opacity="0.18"/>
    <circle cx="${180 + shift}" cy="150" r="54" fill="#fff" opacity="0.9"/>
    <rect x="${250 + variant * 20}" y="190" width="180" height="110" rx="16" fill="#fff" opacity="0.82"/>
    <text x="32" y="360" fill="#fff" font-family="sans-serif" font-size="22">${escapeXml(title)}</text>
  </svg>`
  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`
}

function escapeXml(value: string): string {
  return value.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;")
}
