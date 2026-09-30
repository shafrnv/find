import { useState, type FormEvent } from "react"
import { AuthForm } from "./AuthForm"
import type { FoundDraft, SessionUser } from "./api"
import { categories, lossKinds, subcategories } from "./entries"

interface FoundFormProps {
  user: SessionUser | null
  onLogin: (username: string, password: string) => Promise<void>
  onRegister: (username: string, displayName: string, password: string) => Promise<void>
  onLogout: () => Promise<void>
  onCreate: (draft: Omit<FoundDraft, "lat" | "lon">) => Promise<void>
}

export function FoundForm({ user, onLogin, onRegister, onLogout, onCreate }: FoundFormProps) {
  if (!user) {
    return (
      <AuthForm
        hint="Чтобы опубликовать находку, войдите. Это не роль аккаунта — тот же человек может и искать, и находить, и отмечать места."
        onLogin={onLogin}
        onRegister={onRegister}
      />
    )
  }
  return <CreateForm user={user} onLogout={onLogout} onCreate={onCreate} />
}

function CreateForm({
  user,
  onLogout,
  onCreate,
}: {
  user: SessionUser
  onLogout: FoundFormProps["onLogout"]
  onCreate: FoundFormProps["onCreate"]
}) {
  const [subcategory, setSubcategory] = useState("keys")
  const [title, setTitle] = useState("")
  const [description, setDescription] = useState("")
  const [place, setPlace] = useState("")
  const [at, setAt] = useState(localInputValue)
  const [sex, setSex] = useState("unknown")
  const [trait, setTrait] = useState("")
  const [secretTrait, setSecretTrait] = useState("")
  const [error, setError] = useState("")
  const [busy, setBusy] = useState(false)
  const kind = subcategories.find((item) => item.id === subcategory)?.categoryId ?? "item"

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError("")
    setBusy(true)
    try {
      await onCreate({
        kind,
        subcategory,
        title,
        description,
        place,
        at,
        sex: kind === "item" ? "unknown" : sex,
        trait,
        secretTrait,
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не получилось")
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="composer" onSubmit={submit}>
      <p className="hint">{`Автор этой находки — ${user.displayName}. Точка — центр карты: сдвиньте её к месту.`}</p>
      <div className="field">
        <span className="field-label">Категория</span>
        <ul className="select-menu">
          {categories.filter((category) => lossKinds.has(category.id)).map((category) => (
            <li key={category.id}>
              <span className="tree-label">{category.chip}</span>
              <ul className="subcats">
                {subcategories
                  .filter((item) => item.categoryId === category.id)
                  .map((item) => (
                    <li key={item.id}>
                      <label>
                        <input type="radio" name="subcategory" checked={subcategory === item.id} onChange={() => setSubcategory(item.id)} />
                        {item.label}
                      </label>
                    </li>
                  ))}
              </ul>
            </li>
          ))}
        </ul>
      </div>
      <label>
        Название
        <input value={title} onChange={(event) => setTitle(event.target.value)} />
      </label>
      <label>
        Описание
        <textarea value={description} onChange={(event) => setDescription(event.target.value)} />
      </label>
      <label>
        Место
        <input value={place} onChange={(event) => setPlace(event.target.value)} />
      </label>
      <label>
        Время находки
        <input type="datetime-local" value={at} onChange={(event) => setAt(event.target.value)} />
      </label>
      {kind === "item" ? null : (
        <label>
          Пол
          <select value={sex} onChange={(event) => setSex(event.target.value)}>
            <option value="unknown">Не указан</option>
            {kind === "person" ? (
              <>
                <option value="female">Женщина</option>
                <option value="male">Мужчина</option>
              </>
            ) : (
              <>
                <option value="female">Самка</option>
                <option value="male">Самец</option>
              </>
            )}
          </select>
        </label>
      )}
      <label>
        Примета (видна всем)
        <input value={trait} onChange={(event) => setTrait(event.target.value)} placeholder="Необязательно" />
      </label>
      <label>
        Отличительная особенность (для проверки при чате)
        <input
          value={secretTrait}
          onChange={(event) => setSecretTrait(event.target.value)}
          placeholder="Не видно в карточке; откликнувшийся должен назвать"
        />
      </label>
      <p className="hint">Если задать особенность — чат откроется только после ответа откликнувшегося и вашего подтверждения.</p>
      {error ? <p className="form-error">{error}</p> : null}
      <button type="submit" className="found-button" disabled={busy}>
        Опубликовать
      </button>
      <button type="button" className="text-button" onClick={() => void onLogout()}>
        Выйти
      </button>
    </form>
  )
}

function localInputValue(): string {
  const date = new Date()
  const pad = (value: number) => String(value).padStart(2, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}
