package domain

import (
	"fmt"
	"strings"
	"time"
)

// Animal — животное. Категория задаёт вид, подкатегория или BreedText — породу.
// Хозяин, как и владелец вещи, во внешнюю карточку не входит.
type Animal struct {
	ID              ID
	CategoryID      ID
	BreedCategoryID *ID
	BreedText       string
	Sex             Sex
	Nickname        string
	Photos          []PhotoRef
	Traits          []Trait
	OwnerUserID     *ID
	CreatedAt       time.Time
}

func NewAnimal(id, categoryID ID, breedCategoryID *ID, breedText string, sex Sex, nickname string, photos []PhotoRef, traits []Trait, owner *ID, now time.Time) (Animal, error) {
	if id.IsZero() || categoryID.IsZero() {
		return Animal{}, fmt.Errorf("animal: нужны id и категория вида")
	}
	if breedCategoryID != nil && (breedCategoryID.IsZero() || *breedCategoryID == categoryID) {
		return Animal{}, fmt.Errorf("animal.breed_category_id: некорректная порода")
	}
	breedText = strings.TrimSpace(breedText)
	if tooLong(breedText, 80) {
		return Animal{}, fmt.Errorf("animal.breed_text: длиннее 80 символов")
	}
	if !sex.valid() {
		return Animal{}, fmt.Errorf("animal.sex: female, male или unknown")
	}
	nickname = strings.TrimSpace(nickname)
	if tooLong(nickname, 80) {
		return Animal{}, fmt.Errorf("animal.nickname: длиннее 80 символов")
	}
	if owner != nil && owner.IsZero() {
		return Animal{}, fmt.Errorf("animal.owner_user_id: пустой")
	}
	if now.IsZero() {
		return Animal{}, fmt.Errorf("animal.created_at: обязателен")
	}
	return Animal{
		ID:              id,
		CategoryID:      categoryID,
		BreedCategoryID: breedCategoryID,
		BreedText:       breedText,
		Sex:             sex,
		Nickname:        nickname,
		Photos:          photos,
		Traits:          traits,
		OwnerUserID:     owner,
		CreatedAt:       now,
	}, nil
}

func (a Animal) Kind() SubjectKind { return SubjectAnimal }
func (a Animal) GetID() ID         { return a.ID }

type AnimalView struct {
	ID              ID
	CategoryID      ID
	BreedCategoryID *ID
	BreedText       string
	Sex             Sex
	Nickname        string
	PhotoIDs        []ID
	Traits          []Trait
}

func (a Animal) Public() AnimalView {
	return AnimalView{
		ID:              a.ID,
		CategoryID:      a.CategoryID,
		BreedCategoryID: a.BreedCategoryID,
		BreedText:       a.BreedText,
		Sex:             a.Sex,
		Nickname:        a.Nickname,
		PhotoIDs:        publicPhotoIDs(a.Photos),
		Traits:          publicTraits(a.Traits),
	}
}
