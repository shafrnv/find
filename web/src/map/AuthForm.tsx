import { useState, type FormEvent } from "react"

interface AuthFormProps {
  hint: string
  onLogin: (username: string, password: string) => Promise<void>
  onRegister: (username: string, displayName: string, password: string) => Promise<void>
}

export function AuthForm({ hint, onLogin, onRegister }: AuthFormProps) {
  const [mode, setMode] = useState<"login" | "register">("register")
  const [username, setUsername] = useState("")
  const [displayName, setDisplayName] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [busy, setBusy] = useState(false)

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError("")
    setBusy(true)
    try {
      if (mode === "register") await onRegister(username, displayName, password)
      else await onLogin(username, password)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не получилось")
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="composer" onSubmit={submit}>
      <p className="hint">{hint}</p>
      <label>
        Логин
        <input value={username} autoComplete="username" onChange={(event) => setUsername(event.target.value)} />
      </label>
      {mode === "register" ? (
        <label>
          Имя на карте
          <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} />
        </label>
      ) : null}
      <label>
        Пароль
        <input type="password" autoComplete={mode === "register" ? "new-password" : "current-password"} value={password} onChange={(event) => setPassword(event.target.value)} />
      </label>
      {error ? <p className="form-error">{error}</p> : null}
      <button type="submit" className="found-button" disabled={busy}>
        {mode === "register" ? "Зарегистрироваться" : "Войти"}
      </button>
      <button
        type="button"
        className="text-button"
        onClick={() => {
          setMode((current) => (current === "register" ? "login" : "register"))
          setError("")
        }}
      >
        {mode === "register" ? "Уже есть аккаунт" : "Создать аккаунт"}
      </button>
    </form>
  )
}
