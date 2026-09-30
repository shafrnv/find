package domain

import (
	"fmt"
	"strings"
	"time"
)

// Media — загруженный файл. На карточках лежит не сам файл, а ссылка PhotoRef.
type Media struct {
	ID          ID
	OwnerUserID ID
	StorageKey  string
	MimeType    string
	CreatedAt   time.Time
}

func NewMedia(id, owner ID, storageKey, mime string, now time.Time) (Media, error) {
	if id.IsZero() || owner.IsZero() {
		return Media{}, fmt.Errorf("media: нужны id и владелец файла")
	}
	storageKey = strings.TrimSpace(storageKey)
	if storageKey == "" || strings.Contains(storageKey, "..") {
		return Media{}, fmt.Errorf("media.storage_key: некорректный ключ")
	}
	switch mime {
	case "image/jpeg", "image/png", "image/webp", "image/heic":
	default:
		return Media{}, fmt.Errorf("media.mime_type: нужно изображение jpeg, png, webp или heic")
	}
	if now.IsZero() {
		return Media{}, fmt.Errorf("media.created_at: обязателен")
	}
	return Media{ID: id, OwnerUserID: owner, StorageKey: storageKey, MimeType: mime, CreatedAt: now}, nil
}

// PhotoRef привязывает снимок к субъекту или объявлению.
// Скрытый кадр приметы остаётся internal и не попадает в публичную карточку.
type PhotoRef struct {
	MediaID    ID
	Visibility Visibility
}

func NewPhotoRef(mediaID ID, visibility Visibility) (PhotoRef, error) {
	if mediaID.IsZero() {
		return PhotoRef{}, fmt.Errorf("photo.media_id: обязателен")
	}
	if err := requireVisibility("photo.visibility", visibility); err != nil {
		return PhotoRef{}, err
	}
	return PhotoRef{MediaID: mediaID, Visibility: visibility}, nil
}

func publicPhotoIDs(photos []PhotoRef) []ID {
	out := make([]ID, 0, len(photos))
	for _, photo := range photos {
		if photo.Visibility == VisibilityPublic {
			out = append(out, photo.MediaID)
		}
	}
	return out
}
