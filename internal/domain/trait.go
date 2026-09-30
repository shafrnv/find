package domain

import (
	"fmt"
	"strings"
)

// Канонические ключи внешних характеристик.
// Список открытый: неизвестный ключ допустим, эти константы фиксируют смысл.
const (
	TraitHeight      = "height_cm"
	TraitWeight      = "weight_kg"
	TraitNationality = "nationality"
	TraitAppearance  = "appearance"
	TraitDistinctive = "distinctive"
	TraitColor       = "color"
	TraitAge         = "age"
)

// Trait — одна характеристика субъекта.
// Отличительная черта для проверки владельца создаётся как internal.
type Trait struct {
	Key        string
	Value      string
	Visibility Visibility
}

func NewTrait(key, value string, visibility Visibility) (Trait, error) {
	key = strings.TrimSpace(strings.ToLower(key))
	if key == "" || len(key) > 40 || strings.ContainsAny(key, " \n\t") {
		return Trait{}, fmt.Errorf("trait.key: нужен короткий ключ без пробелов")
	}
	value = strings.TrimSpace(value)
	if value == "" || len([]rune(value)) > 2000 {
		return Trait{}, fmt.Errorf("trait.value: от 1 до 2000 символов")
	}
	if err := requireVisibility("trait.visibility", visibility); err != nil {
		return Trait{}, err
	}
	return Trait{Key: key, Value: value, Visibility: visibility}, nil
}

// SecretDistinctive — примета, которую должен назвать владелец.
// В публичную карточку она не попадает.
func SecretDistinctive(value string) (Trait, error) {
	return NewTrait(TraitDistinctive, value, VisibilityInternal)
}

func publicTraits(traits []Trait) []Trait {
	out := make([]Trait, 0, len(traits))
	for _, trait := range traits {
		if trait.Visibility == VisibilityPublic {
			out = append(out, trait)
		}
	}
	return out
}
