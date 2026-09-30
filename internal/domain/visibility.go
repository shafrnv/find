package domain

import "fmt"

// Visibility отделяет внешние поля карточки от внутренних.
// Внешнее видно в поиске. Внутреннее видно сторонам дела и системе:
// владелец вещи, скрытая примета для проверки, автор объявления «ищу».
type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityInternal Visibility = "internal"
)

func (v Visibility) valid() bool {
	return v == VisibilityPublic || v == VisibilityInternal
}

func requireVisibility(field string, v Visibility) error {
	if !v.valid() {
		return fmt.Errorf("%s: видимость public или internal", field)
	}
	return nil
}
