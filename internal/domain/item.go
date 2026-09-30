package domain

import (
	"fmt"
	"strings"
	"time"
)

// Item — предмет. Потеря и находка описываются объявлениями,
// а не полями самой вещи: место и время бывают разными у ищущего и нашедшего.
// OwnerUserID внутренний: в поиске владельца не показывают.
type Item struct {
	ID            ID
	Name          string
	CategoryID    ID
	SubcategoryID *ID
	Description   string
	Photos        []PhotoRef
	Traits        []Trait
	OwnerUserID   *ID
	CreatedAt     time.Time
}

func NewItem(id ID, name string, categoryID ID, subcategoryID *ID, description string, photos []PhotoRef, traits []Trait, owner *ID, now time.Time) (Item, error) {
	if id.IsZero() || categoryID.IsZero() {
		return Item{}, fmt.Errorf("item: нужны id и категория")
	}
	if subcategoryID != nil && (subcategoryID.IsZero() || *subcategoryID == categoryID) {
		return Item{}, fmt.Errorf("item.subcategory_id: некорректная подкатегория")
	}
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 160 {
		return Item{}, fmt.Errorf("item.name: от 1 до 160 символов")
	}
	description = strings.TrimSpace(description)
	if len([]rune(description)) > 8000 {
		return Item{}, fmt.Errorf("item.description: длиннее 8000 символов")
	}
	if owner != nil && owner.IsZero() {
		return Item{}, fmt.Errorf("item.owner_user_id: пустой")
	}
	if now.IsZero() {
		return Item{}, fmt.Errorf("item.created_at: обязателен")
	}
	return Item{
		ID:            id,
		Name:          name,
		CategoryID:    categoryID,
		SubcategoryID: subcategoryID,
		Description:   description,
		Photos:        photos,
		Traits:        traits,
		OwnerUserID:   owner,
		CreatedAt:     now,
	}, nil
}

func (i Item) Kind() SubjectKind { return SubjectItem }
func (i Item) GetID() ID         { return i.ID }

// ItemView — внешняя карточка вещи.
type ItemView struct {
	ID            ID
	Name          string
	CategoryID    ID
	SubcategoryID *ID
	Description   string
	PhotoIDs      []ID
	Traits        []Trait
}

func (i Item) Public() ItemView {
	return ItemView{
		ID:            i.ID,
		Name:          i.Name,
		CategoryID:    i.CategoryID,
		SubcategoryID: i.SubcategoryID,
		Description:   i.Description,
		PhotoIDs:      publicPhotoIDs(i.Photos),
		Traits:        publicTraits(i.Traits),
	}
}
