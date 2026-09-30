package domain

import (
	"strings"
	"testing"
	"time"
)

func TestPublicItemHidesOwnerAndSecretTrait(t *testing.T) {
	now := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	owner := uid("11111111-1111-1111-1111-111111111111")
	photo, err := NewPhotoRef(uid("22222222-2222-2222-2222-222222222222"), VisibilityPublic)
	if err != nil {
		t.Fatal(err)
	}
	secretPhoto, err := NewPhotoRef(uid("33333333-3333-3333-3333-333333333333"), VisibilityInternal)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := SecretDistinctive("царапина внутри чехла")
	if err != nil {
		t.Fatal(err)
	}
	color, err := NewTrait(TraitColor, "белый", VisibilityPublic)
	if err != nil {
		t.Fatal(err)
	}
	item, err := NewItem(
		uid("44444444-4444-4444-4444-444444444444"),
		"AirPods Pro",
		uid("55555555-5555-5555-5555-555555555555"),
		nil,
		"кейс без гравировки",
		[]PhotoRef{photo, secretPhoto},
		[]Trait{color, secret},
		&owner,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	view := item.Public()
	if view.Name != "AirPods Pro" {
		t.Fatalf("name: %s", view.Name)
	}
	if len(view.PhotoIDs) != 1 || view.PhotoIDs[0] != photo.MediaID {
		t.Fatalf("photos: %+v", view.PhotoIDs)
	}
	if len(view.Traits) != 1 || view.Traits[0].Key != TraitColor {
		t.Fatalf("traits: %+v", view.Traits)
	}
}

func TestFoundPostShowsFinderAndSeekingPostHidesAuthor(t *testing.T) {
	now := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	when := now.Add(-time.Hour)
	point, err := NewGeoPoint(59.93, 30.31)
	if err != nil {
		t.Fatal(err)
	}
	itemID := uid("44444444-4444-4444-4444-444444444444")
	finder := uid("66666666-6666-6666-6666-666666666666")
	seeker := uid("77777777-7777-7777-7777-777777777777")

	found, err := NewPublication(
		uid("88888888-8888-8888-8888-888888888888"),
		finder,
		IntentFound,
		SubjectItem,
		itemID,
		"лежит на скамейке",
		&when,
		&point,
		"у метро",
		nil,
		nil,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	foundView := found.Public()
	if foundView.AuthorUserID == nil || *foundView.AuthorUserID != finder {
		t.Fatalf("finder must be public, got %+v", foundView.AuthorUserID)
	}

	seeking, err := NewPublication(
		uid("99999999-9999-9999-9999-999999999999"),
		seeker,
		IntentSeeking,
		SubjectItem,
		itemID,
		"выпали в метро",
		&when,
		nil,
		"Невский проспект",
		nil,
		nil,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	if seeking.Public().AuthorUserID != nil {
		t.Fatal("seeker is internal until they reveal themselves")
	}
	revealed, err := seeking.RevealAuthor()
	if err != nil {
		t.Fatal(err)
	}
	if revealed.Public().AuthorUserID == nil || *revealed.Public().AuthorUserID != seeker {
		t.Fatal("revealed seeker must be public")
	}

	match, err := NewMatch(uid("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), revealed, found, now)
	if err != nil {
		t.Fatal(err)
	}
	if match.Status != MatchProposed {
		t.Fatalf("status: %s", match.Status)
	}
}

func TestPlaceAndEventAreMarkedNotLost(t *testing.T) {
	now := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	author := uid("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	place, err := NewPlace(
		uid("cccccccc-cccc-cccc-cccc-cccccccccccc"),
		uid("dddddddd-dddd-dddd-dddd-dddddddddddd"),
		nil,
		"Двор на Рубинштейна",
		"тихий вход со двора",
		"Рубинштейна, 15",
		nil,
		[]string{"Тихий-двор", "тихий-двор"},
		nil,
		nil,
		author,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(place.Tags) != 1 || place.Tags[0] != "тихий-двор" {
		t.Fatalf("tags: %+v", place.Tags)
	}

	_, err = NewPublication(
		uid("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"),
		author,
		IntentSeeking,
		SubjectPlace,
		place.ID,
		"",
		nil,
		nil,
		"",
		nil,
		nil,
		now,
	)
	if err == nil || !strings.Contains(err.Error(), "отмечают") {
		t.Fatalf("seeking a place must fail, got %v", err)
	}

	mark, err := NewPublication(
		uid("ffffffff-ffff-ffff-ffff-ffffffffffff"),
		author,
		IntentMarking,
		SubjectPlace,
		place.ID,
		"здесь нормальный кофе после десяти",
		nil,
		nil,
		"",
		&place.ID,
		nil,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	if mark.Public().AuthorUserID == nil {
		t.Fatal("mark author is public")
	}
}

func TestFollowAndEventBounds(t *testing.T) {
	now := time.Date(2026, 9, 29, 18, 0, 0, 0, time.UTC)
	user := uid("11111111-1111-1111-1111-111111111111")
	if _, err := NewFollow(user, FollowUser, user, now); err == nil {
		t.Fatal("self follow must fail")
	}
	ends := now.Add(-time.Hour)
	_, err := NewEvent(
		uid("12121212-1212-1212-1212-121212121212"),
		uid("13131313-1313-1313-1313-131313131313"),
		nil,
		"Сбор находок",
		"",
		nil,
		"парк",
		nil,
		now,
		&ends,
		nil,
		nil,
		user,
		now,
	)
	if err == nil {
		t.Fatal("event ending before start must fail")
	}
}

func TestPersonRequiresAName(t *testing.T) {
	_, err := NewPerson(uid("14141414-1414-1414-1414-141414141414"), " ", "", "", SexUnknown, nil, nil, time.Now())
	if err == nil {
		t.Fatal("empty person name must fail")
	}
	person, err := NewPerson(
		uid("15151515-1515-1515-1515-151515151515"),
		"Иван",
		"Петров",
		"Сергеевич",
		SexMale,
		nil,
		[]Trait{{Key: TraitHeight, Value: "180", Visibility: VisibilityPublic}},
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if person.Public().FamilyName != "Петров" {
		t.Fatalf("family: %s", person.Public().FamilyName)
	}
}

func uid(raw string) ID {
	id, err := ParseID(raw)
	if err != nil {
		panic(err)
	}
	return id
}
