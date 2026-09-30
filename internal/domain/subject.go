package domain

// SubjectKind — что именно ищут или отмечают.
type SubjectKind string

const (
	SubjectItem   SubjectKind = "item"
	SubjectPerson SubjectKind = "person"
	SubjectAnimal SubjectKind = "animal"
	SubjectPlace  SubjectKind = "place"
	SubjectEvent  SubjectKind = "event"
)

func (k SubjectKind) valid() bool {
	switch k {
	case SubjectItem, SubjectPerson, SubjectAnimal, SubjectPlace, SubjectEvent:
		return true
	default:
		return false
	}
}

// LostFound сообщает, бывает ли у этого субъекта пара «ищущий / нашедший».
// Место и событие отмечают, а не теряют.
func (k SubjectKind) LostFound() bool {
	return k == SubjectItem || k == SubjectPerson || k == SubjectAnimal
}

// Subject — общий контракт пяти сущностей поиска.
type Subject interface {
	Kind() SubjectKind
	GetID() ID
}
