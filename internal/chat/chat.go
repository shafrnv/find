// Package chat — самостоятельный блок переписки по отклику на объявление.
//
// Границы блока:
//   - создание диалога после отклика (claim / respond);
//   - условная верификация по скрытой «отличительной особенности»;
//   - сообщения только после статуса open.
//
// Не входит сюда:
//   - карта, фильтры, профили (карта только вызывает claim);
//   - встреча / handshake / QR — пакет meeting (следующий блок).
//
// Статусы диалога:
//   pending_answer  — откликнувшийся ещё не назвал примету;
//   pending_confirm — ждём подтверждения автора объявления;
//   open            — можно писать;
//   rejected        — автор отклонил ответ;
//   closed          — закрыт вручную (зарезервировано).
package chat

// Status — жизнь диалога до и после верификации.
type Status string

const (
	StatusPendingAnswer  Status = "pending_answer"
	StatusPendingConfirm Status = "pending_confirm"
	StatusOpen           Status = "open"
	StatusRejected       Status = "rejected"
	StatusClosed         Status = "closed"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPendingAnswer, StatusPendingConfirm, StatusOpen, StatusRejected, StatusClosed:
		return true
	default:
		return false
	}
}

// CanSendMessage — сообщения разрешены только в открытом чате.
func CanSendMessage(status Status) bool {
	return status == StatusOpen
}

// InitialStatus — если у субъекта есть скрытая примета, чат начинается с проверки.
func InitialStatus(hasSecretTrait bool) Status {
	if hasSecretTrait {
		return StatusPendingAnswer
	}
	return StatusOpen
}
