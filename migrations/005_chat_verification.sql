-- Верификация перед открытием чата: ответ откликнувшегося и подтверждение автора.
-- Встреча (meeting) в эту миграцию не входит — отдельный модуль.

alter table dialogs
    add column status text not null default 'open'
        check (status in ('pending_answer', 'pending_confirm', 'open', 'rejected', 'closed')),
    add column respondent_answer text not null default '',
    add column verified_at timestamptz;

-- Старые диалоги без проверки остаются открытыми (default open).
