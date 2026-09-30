-- Вход в аккаунт и подкатегория человека.
-- Паспорт и платёжные реквизиты по-прежнему не хранятся.

alter table users
    add column password_hash text not null;

alter table persons
    add column subcategory_id uuid references categories (id);

create table sessions (
    token_hash text primary key,
    user_id uuid not null references users (id) on delete cascade,
    created_at timestamptz not null,
    expires_at timestamptz not null
);

create index sessions_user_idx on sessions (user_id);
