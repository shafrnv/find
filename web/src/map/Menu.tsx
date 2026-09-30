import type { SessionUser } from "./api"

interface MenuProps {
  user: SessionUser | null
  onProfile: () => void
  onChats: () => void
}

export function Menu({ user, onProfile, onChats }: MenuProps) {
  return (
    <nav className="menu" aria-label="Меню">
      <button type="button" onClick={onProfile}>
        {user ? "Профиль" : "Войти"}
      </button>
      <button type="button" onClick={onChats}>
        Чаты
      </button>
      <button type="button">Подписки</button>
      <button type="button" className="avatar" aria-label={user ? `Профиль ${user.displayName}` : "Войти"} onClick={onProfile}>
        <svg viewBox="0 0 24 24" width="18" height="18">
          <circle cx="12" cy="8" r="3.2" fill="none" stroke="currentColor" strokeWidth="1.8" />
          <path
            d="M5 19.2c1.4-3 3.8-4.5 7-4.5s5.6 1.5 7 4.5"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            strokeLinecap="round"
          />
        </svg>
      </button>
    </nav>
  )
}
