package domain

import (
	"fmt"
	"time"
)

// FollowTarget — за кем или за чем следят, как за каналом.
type FollowTarget string

const (
	FollowUser  FollowTarget = "user"
	FollowPlace FollowTarget = "place"
	FollowEvent FollowTarget = "event"
)

func (t FollowTarget) valid() bool {
	return t == FollowUser || t == FollowPlace || t == FollowEvent
}

// Follow — подписка. Лента автора собирается из его объявлений.
// Подписка на место или событие собирает отметки разных авторов об этом субъекте.
type Follow struct {
	FollowerUserID ID
	TargetKind     FollowTarget
	TargetID       ID
	CreatedAt      time.Time
}

func NewFollow(follower ID, target FollowTarget, targetID ID, now time.Time) (Follow, error) {
	if follower.IsZero() || targetID.IsZero() {
		return Follow{}, fmt.Errorf("follow: нужны подписчик и цель")
	}
	if !target.valid() {
		return Follow{}, fmt.Errorf("follow.target_kind: user, place или event")
	}
	if target == FollowUser && follower == targetID {
		return Follow{}, fmt.Errorf("follow: на себя не подписываются")
	}
	if now.IsZero() {
		return Follow{}, fmt.Errorf("follow.created_at: обязателен")
	}
	return Follow{FollowerUserID: follower, TargetKind: target, TargetID: targetID, CreatedAt: now}, nil
}
