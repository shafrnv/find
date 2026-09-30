package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"find/internal/chat"
)

type DialogSummary struct {
	ID            string    `json:"id"`
	PublicationID string    `json:"publicationId"`
	Title         string    `json:"title"`
	Place         string    `json:"place"`
	Kind          string    `json:"kind"`
	Intent        string    `json:"intent"`
	PeerName      string    `json:"peerName"`
	PeerUsername  string    `json:"peerUsername"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	LastBody      string    `json:"lastBody"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type DialogDetail struct {
	ID               string    `json:"id"`
	PublicationID    string    `json:"publicationId"`
	Title            string    `json:"title"`
	Place            string    `json:"place"`
	Kind             string    `json:"kind"`
	Intent           string    `json:"intent"`
	PeerName         string    `json:"peerName"`
	PeerUsername     string    `json:"peerUsername"`
	Role             string    `json:"role"`
	Status           string    `json:"status"`
	HasChallenge     bool      `json:"hasChallenge"`
	RespondentAnswer string    `json:"respondentAnswer,omitempty"`
	CanMessage       bool      `json:"canMessage"`
	CanAnswer        bool      `json:"canAnswer"`
	CanConfirm       bool      `json:"canConfirm"`
	CreatedAt        time.Time `json:"createdAt"`
	Messages         []Message `json:"messages"`
}

type Message struct {
	ID        string    `json:"id"`
	SenderID  string    `json:"senderId"`
	Mine      bool      `json:"mine"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// ClaimFound — отклик: создаёт диалог. Если у субъекта есть скрытая примета — чат ждёт верификации.
func (s *Store) ClaimFound(ctx context.Context, publicationID, respondentID string) (DialogDetail, error) {
	var authorID, intent, status, kind, subjectID string
	var title, place string
	err := s.pool.QueryRow(ctx, `
		select pub.author_user_id::text, pub.intent, pub.status, pub.subject_kind, pub.subject_id::text,
			coalesce(nullif(it.name, ''), nullif(trim(both from pe.given_name || ' ' || pe.family_name), ''), nullif(an.nickname, ''), ''),
			coalesce(nullif(pub.place_label, ''), '')
		from publications pub
		left join items it on pub.subject_kind = 'item' and it.id = pub.subject_id
		left join persons pe on pub.subject_kind = 'person' and pe.id = pub.subject_id
		left join animals an on pub.subject_kind = 'animal' and an.id = pub.subject_id
		where pub.id = $1`, publicationID).
		Scan(&authorID, &intent, &status, &kind, &subjectID, &title, &place)
	if errors.Is(err, pgx.ErrNoRows) {
		return DialogDetail{}, ErrNotFound
	}
	if err != nil {
		return DialogDetail{}, err
	}
	if intent != "found" && intent != "seeking" {
		return DialogDetail{}, ErrBadClaim
	}
	if kind != "item" && kind != "person" && kind != "animal" {
		return DialogDetail{}, ErrBadClaim
	}
	if status != "published" && status != "matched" {
		return DialogDetail{}, ErrBadClaim
	}
	if authorID == respondentID {
		return DialogDetail{}, fmt.Errorf("нельзя откликнуться на своё объявление")
	}

	var dialogID string
	err = s.pool.QueryRow(ctx, `
		select id::text from dialogs
		where publication_id = $1 and respondent_user_id = $2`, publicationID, respondentID).Scan(&dialogID)
	if err == nil {
		return s.Dialog(ctx, dialogID, respondentID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return DialogDetail{}, err
	}

	hasSecret, err := s.hasSecretDistinctive(ctx, kind, subjectID)
	if err != nil {
		return DialogDetail{}, err
	}
	dialogStatus := string(chat.InitialStatus(hasSecret))

	dialogID, err = newID()
	if err != nil {
		return DialogDetail{}, err
	}
	now := time.Now()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return DialogDetail{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		insert into dialogs (id, publication_id, author_user_id, respondent_user_id, created_at, status)
		values ($1, $2, $3, $4, $5, $6)`, dialogID, publicationID, authorID, respondentID, now, dialogStatus); err != nil {
		return DialogDetail{}, err
	}
	if dialogStatus == string(chat.StatusOpen) {
		msgID, err := newID()
		if err != nil {
			return DialogDetail{}, err
		}
		if _, err := tx.Exec(ctx, `
			insert into messages (id, dialog_id, sender_user_id, body, created_at)
			values ($1, $2, $3, $4, $5)`,
			msgID, dialogID, respondentID, firstMessage(kind, intent), now); err != nil {
			return DialogDetail{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return DialogDetail{}, err
	}
	return s.Dialog(ctx, dialogID, respondentID)
}

func (s *Store) hasSecretDistinctive(ctx context.Context, kind, subjectID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		select exists(
			select 1 from traits
			where subject_kind = $1 and subject_id = $2
			  and key = 'distinctive' and visibility = 'internal'
		)`, kind, subjectID).Scan(&exists)
	return exists, err
}

func firstMessage(kind, intent string) string {
	if kind == "person" {
		if intent == "found" {
			return "Здравствуйте! Я знаю этого человека / это касается меня. Давайте свяжемся."
		}
		return "Здравствуйте! Возможно, я видел этого человека. Давайте свяжемся."
	}
	if intent == "found" {
		return "Здравствуйте! Это моё. Давайте созвонимся по поводу возврата."
	}
	return "Здравствуйте! Кажется, я нашёл то, что вы ищете. Давайте свяжемся."
}

func (s *Store) ListDialogs(ctx context.Context, userID string) ([]DialogSummary, error) {
	rows, err := s.pool.Query(ctx, `
		select
			d.id::text,
			d.publication_id::text,
			coalesce(nullif(it.name, ''), nullif(trim(both from pe.given_name || ' ' || pe.family_name), ''), nullif(an.nickname, ''), 'Объявление'),
			coalesce(nullif(pub.place_label, ''), ''),
			pub.subject_kind,
			pub.intent,
			case when d.author_user_id = $1 then respondent.display_name else author.display_name end,
			case when d.author_user_id = $1 then respondent.username else author.username end,
			case when d.author_user_id = $1 then 'author' else 'respondent' end,
			d.status,
			coalesce((
				select m.body from messages m
				where m.dialog_id = d.id
				order by m.created_at desc limit 1
			), case
				when d.status = 'pending_answer' then 'Нужно назвать отличительную особенность'
				when d.status = 'pending_confirm' then 'Ждём подтверждения приметы'
				when d.status = 'rejected' then 'Отклик отклонён'
				else ''
			end),
			coalesce((
				select m.created_at from messages m
				where m.dialog_id = d.id
				order by m.created_at desc limit 1
			), d.created_at)
		from dialogs d
		join publications pub on pub.id = d.publication_id
		join users author on author.id = d.author_user_id
		join users respondent on respondent.id = d.respondent_user_id
		left join items it on pub.subject_kind = 'item' and it.id = pub.subject_id
		left join persons pe on pub.subject_kind = 'person' and pe.id = pub.subject_id
		left join animals an on pub.subject_kind = 'animal' and an.id = pub.subject_id
		where d.author_user_id = $1 or d.respondent_user_id = $1
		order by 12 desc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DialogSummary
	for rows.Next() {
		var item DialogSummary
		if err := rows.Scan(
			&item.ID, &item.PublicationID, &item.Title, &item.Place, &item.Kind, &item.Intent,
			&item.PeerName, &item.PeerUsername, &item.Role, &item.Status, &item.LastBody, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	if list == nil {
		list = []DialogSummary{}
	}
	return list, rows.Err()
}

func (s *Store) Dialog(ctx context.Context, dialogID, userID string) (DialogDetail, error) {
	var detail DialogDetail
	var authorID, respondentID, subjectKind, subjectID string
	err := s.pool.QueryRow(ctx, `
		select
			d.id::text,
			d.publication_id::text,
			d.author_user_id::text,
			d.respondent_user_id::text,
			d.created_at,
			d.status,
			d.respondent_answer,
			coalesce(nullif(it.name, ''), nullif(trim(both from pe.given_name || ' ' || pe.family_name), ''), nullif(an.nickname, ''), 'Объявление'),
			coalesce(nullif(pub.place_label, ''), ''),
			pub.subject_kind,
			pub.intent,
			pub.subject_id::text
		from dialogs d
		join publications pub on pub.id = d.publication_id
		left join items it on pub.subject_kind = 'item' and it.id = pub.subject_id
		left join persons pe on pub.subject_kind = 'person' and pe.id = pub.subject_id
		left join animals an on pub.subject_kind = 'animal' and an.id = pub.subject_id
		where d.id = $1`, dialogID).
		Scan(&detail.ID, &detail.PublicationID, &authorID, &respondentID, &detail.CreatedAt, &detail.Status, &detail.RespondentAnswer,
			&detail.Title, &detail.Place, &subjectKind, &detail.Intent, &subjectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return DialogDetail{}, ErrNotFound
	}
	if err != nil {
		return DialogDetail{}, err
	}
	if userID != authorID && userID != respondentID {
		return DialogDetail{}, ErrForbidden
	}
	detail.Kind = subjectKind
	hasSecret, err := s.hasSecretDistinctive(ctx, subjectKind, subjectID)
	if err != nil {
		return DialogDetail{}, err
	}
	detail.HasChallenge = hasSecret
	st := chat.Status(detail.Status)
	detail.CanMessage = chat.CanSendMessage(st)
	detail.Messages = []Message{}
	if userID == authorID {
		detail.Role = "author"
		detail.CanConfirm = st == chat.StatusPendingConfirm
		if err := s.pool.QueryRow(ctx, `select display_name, username from users where id = $1`, respondentID).
			Scan(&detail.PeerName, &detail.PeerUsername); err != nil {
			return DialogDetail{}, err
		}
	} else {
		detail.Role = "respondent"
		detail.CanAnswer = st == chat.StatusPendingAnswer
		// Ответ на примету видит только автор (для подтверждения).
		detail.RespondentAnswer = ""
		if err := s.pool.QueryRow(ctx, `select display_name, username from users where id = $1`, authorID).
			Scan(&detail.PeerName, &detail.PeerUsername); err != nil {
			return DialogDetail{}, err
		}
	}

	if detail.CanMessage {
		rows, err := s.pool.Query(ctx, `
			select id::text, sender_user_id::text, body, created_at
			from messages where dialog_id = $1
			order by created_at`, dialogID)
		if err != nil {
			return DialogDetail{}, err
		}
		defer rows.Close()
		for rows.Next() {
			var msg Message
			if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Body, &msg.CreatedAt); err != nil {
				return DialogDetail{}, err
			}
			msg.Mine = msg.SenderID == userID
			detail.Messages = append(detail.Messages, msg)
		}
		if err := rows.Err(); err != nil {
			return DialogDetail{}, err
		}
	}
	return detail, nil
}

func (s *Store) SubmitDialogAnswer(ctx context.Context, dialogID, userID, answer string) (DialogDetail, error) {
	answer = strings.TrimSpace(answer)
	if answer == "" || len([]rune(answer)) > 2000 {
		return DialogDetail{}, fmt.Errorf("ответ: от 1 до 2000 символов")
	}
	var authorID, respondentID, status string
	err := s.pool.QueryRow(ctx, `
		select author_user_id::text, respondent_user_id::text, status from dialogs where id = $1`, dialogID).
		Scan(&authorID, &respondentID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return DialogDetail{}, ErrNotFound
	}
	if err != nil {
		return DialogDetail{}, err
	}
	if userID != respondentID {
		return DialogDetail{}, ErrForbidden
	}
	if status != string(chat.StatusPendingAnswer) {
		return DialogDetail{}, fmt.Errorf("сейчас нельзя отправить ответ на примету")
	}
	_, err = s.pool.Exec(ctx, `
		update dialogs set status = $2, respondent_answer = $3 where id = $1`,
		dialogID, string(chat.StatusPendingConfirm), answer)
	if err != nil {
		return DialogDetail{}, err
	}
	return s.Dialog(ctx, dialogID, userID)
}

func (s *Store) ConfirmDialogAnswer(ctx context.Context, dialogID, userID string, accept bool) (DialogDetail, error) {
	var authorID, status string
	err := s.pool.QueryRow(ctx, `
		select author_user_id::text, status from dialogs where id = $1`, dialogID).
		Scan(&authorID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return DialogDetail{}, ErrNotFound
	}
	if err != nil {
		return DialogDetail{}, err
	}
	if userID != authorID {
		return DialogDetail{}, ErrForbidden
	}
	if status != string(chat.StatusPendingConfirm) {
		return DialogDetail{}, fmt.Errorf("сейчас нельзя подтвердить примету")
	}
	next := chat.StatusRejected
	var verified any
	if accept {
		next = chat.StatusOpen
		verified = time.Now()
	}
	_, err = s.pool.Exec(ctx, `
		update dialogs set status = $2, verified_at = $3 where id = $1`, dialogID, string(next), verified)
	if err != nil {
		return DialogDetail{}, err
	}
	if accept {
		msgID, err := newID()
		if err != nil {
			return DialogDetail{}, err
		}
		if _, err := s.pool.Exec(ctx, `
			insert into messages (id, dialog_id, sender_user_id, body, created_at)
			values ($1, $2, $3, $4, $5)`,
			msgID, dialogID, userID, "Примета подтверждена. Можно продолжать переписку.", time.Now()); err != nil {
			return DialogDetail{}, err
		}
	}
	return s.Dialog(ctx, dialogID, userID)
}

func (s *Store) SendMessage(ctx context.Context, dialogID, userID, body string) (Message, error) {
	body = strings.TrimSpace(body)
	if body == "" || len([]rune(body)) > 4000 {
		return Message{}, fmt.Errorf("сообщение: от 1 до 4000 символов")
	}
	var authorID, respondentID, status string
	err := s.pool.QueryRow(ctx, `
		select author_user_id::text, respondent_user_id::text, status from dialogs where id = $1`, dialogID).
		Scan(&authorID, &respondentID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Message{}, ErrNotFound
	}
	if err != nil {
		return Message{}, err
	}
	if userID != authorID && userID != respondentID {
		return Message{}, ErrForbidden
	}
	if !chat.CanSendMessage(chat.Status(status)) {
		return Message{}, fmt.Errorf("переписка откроется после подтверждения отличительной особенности")
	}
	id, err := newID()
	if err != nil {
		return Message{}, err
	}
	now := time.Now()
	if _, err := s.pool.Exec(ctx, `
		insert into messages (id, dialog_id, sender_user_id, body, created_at)
		values ($1, $2, $3, $4, $5)`, id, dialogID, userID, body, now); err != nil {
		return Message{}, err
	}
	return Message{ID: id, SenderID: userID, Mine: true, Body: body, CreatedAt: now}, nil
}
