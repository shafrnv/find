package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// Category — узел справочника. Scope совпадает с видом субъекта.
// Подкатегория — дочерний узел (у вещи — тип, у животного — порода,
// у места и события — свой справочник).
type Category struct {
	ID       ID
	Scope    SubjectKind
	ParentID *ID
	Name     string
	Slug     string
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func NewCategory(id ID, scope SubjectKind, parentID *ID, name, slug string) (Category, error) {
	if id.IsZero() {
		return Category{}, fmt.Errorf("category.id: обязателен")
	}
	if !scope.valid() {
		return Category{}, fmt.Errorf("category.scope: неизвестный вид субъекта")
	}
	if parentID != nil && parentID.IsZero() {
		return Category{}, fmt.Errorf("category.parent_id: пустой")
	}
	if parentID != nil && *parentID == id {
		return Category{}, fmt.Errorf("category.parent_id: категория не может быть родителем самой себя")
	}
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 80 {
		return Category{}, fmt.Errorf("category.name: от 1 до 80 символов")
	}
	slug = strings.TrimSpace(strings.ToLower(slug))
	if !slugPattern.MatchString(slug) || len(slug) > 80 {
		return Category{}, fmt.Errorf("category.slug: латиница, цифры и дефис")
	}
	return Category{ID: id, Scope: scope, ParentID: parentID, Name: name, Slug: slug}, nil
}

// ValidChildOf проверяет, что подкатегория лежит в том же справочнике
// и ссылается на родителя.
func (c Category) ValidChildOf(parent Category) error {
	if c.ParentID == nil || *c.ParentID != parent.ID {
		return fmt.Errorf("category.parent_id: не совпадает с родителем")
	}
	if c.Scope != parent.Scope {
		return fmt.Errorf("category.scope: подкатегория должна быть в том же справочнике")
	}
	if parent.ParentID != nil {
		return fmt.Errorf("category.parent_id: вложенность пока только на один уровень")
	}
	return nil
}
