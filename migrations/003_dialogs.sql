-- Диалог по отклику «Это моё» на находку.
-- Роли пользователя нет: один человек может и находить, и терять, и отмечать.

create table dialogs (
    id uuid primary key,
    found_publication_id uuid not null references publications (id),
    finder_user_id uuid not null references users (id),
    claimer_user_id uuid not null references users (id),
    created_at timestamptz not null,
    unique (found_publication_id, claimer_user_id),
    check (finder_user_id <> claimer_user_id)
);

create index dialogs_participant_idx on dialogs (finder_user_id, created_at desc);
create index dialogs_claimer_idx on dialogs (claimer_user_id, created_at desc);

create table messages (
    id uuid primary key,
    dialog_id uuid not null references dialogs (id) on delete cascade,
    sender_user_id uuid not null references users (id),
    body text not null,
    created_at timestamptz not null,
    check (char_length(body) > 0)
);

create index messages_dialog_idx on messages (dialog_id, created_at);
