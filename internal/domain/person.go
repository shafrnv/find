package domain

import (
	"fmt"
	"strings"
	"time"
)

// Sex — пол человека или животного.
type Sex string

const (
	SexFemale  Sex = "female"
	SexMale    Sex = "male"
	SexUnknown Sex = "unknown"
)

func (s Sex) valid() bool {
	return s == SexFemale || s == SexMale || s == SexUnknown
}

// Person — пропавший или найденный человек. Это не аккаунт User.
// Имя, пол, фото и внешние характеристики публичны.
// Кто подал ориентировку и как с ним связаться — автор объявления,
// и для намерения «ищу» он по умолчанию скрыт.
type Person struct {
	ID         ID
	GivenName  string
	FamilyName string
	Patronymic string
	Sex        Sex
	Photos     []PhotoRef
	Traits     []Trait
	CreatedAt  time.Time
}

func NewPerson(id ID, given, family, patronymic string, sex Sex, photos []PhotoRef, traits []Trait, now time.Time) (Person, error) {
	if id.IsZero() {
		return Person{}, fmt.Errorf("person.id: обязателен")
	}
	given = strings.TrimSpace(given)
	family = strings.TrimSpace(family)
	patronymic = strings.TrimSpace(patronymic)
	if given == "" && family == "" {
		return Person{}, fmt.Errorf("person.name: нужно имя или фамилия")
	}
	if tooLong(given, 80) || tooLong(family, 80) || tooLong(patronymic, 80) {
		return Person{}, fmt.Errorf("person.name: часть имени длиннее 80 символов")
	}
	if !sex.valid() {
		return Person{}, fmt.Errorf("person.sex: female, male или unknown")
	}
	if now.IsZero() {
		return Person{}, fmt.Errorf("person.created_at: обязателен")
	}
	return Person{
		ID:         id,
		GivenName:  given,
		FamilyName: family,
		Patronymic: patronymic,
		Sex:        sex,
		Photos:     photos,
		Traits:     traits,
		CreatedAt:  now,
	}, nil
}

func (p Person) Kind() SubjectKind { return SubjectPerson }
func (p Person) GetID() ID         { return p.ID }

type PersonView struct {
	ID         ID
	GivenName  string
	FamilyName string
	Patronymic string
	Sex        Sex
	PhotoIDs   []ID
	Traits     []Trait
}

func (p Person) Public() PersonView {
	return PersonView{
		ID:         p.ID,
		GivenName:  p.GivenName,
		FamilyName: p.FamilyName,
		Patronymic: p.Patronymic,
		Sex:        p.Sex,
		PhotoIDs:   publicPhotoIDs(p.Photos),
		Traits:     publicTraits(p.Traits),
	}
}

func tooLong(value string, max int) bool {
	return len([]rune(value)) > max
}
