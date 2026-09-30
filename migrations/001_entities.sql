-- Сущности поиска. Идентификаторы выдаёт приложение.
-- Паспорт и платёжные реквизиты в этой схеме отсутствуют.

create table users (
    id uuid primary key,
    username text not null unique,
    display_name text not null,
    bio text not null default '',
    avatar_media_id uuid,
    created_at timestamptz not null
);

create table identity_verifications (
    user_id uuid primary key references users (id),
    status text not null check (status in ('none', 'pending', 'verified')),
    verified_at timestamptz,
    check (
        (status = 'verified' and verified_at is not null)
        or (status <> 'verified' and verified_at is null)
    )
);

create table categories (
    id uuid primary key,
    scope text not null check (scope in ('item', 'person', 'animal', 'place', 'event')),
    parent_id uuid references categories (id),
    name text not null,
    slug text not null,
    check (parent_id is null or parent_id <> id),
    unique (scope, slug)
);

create table media (
    id uuid primary key,
    owner_user_id uuid not null references users (id),
    storage_key text not null unique,
    mime_type text not null check (mime_type in ('image/jpeg', 'image/png', 'image/webp', 'image/heic')),
    created_at timestamptz not null
);

alter table users
    add constraint users_avatar_media_fk
    foreign key (avatar_media_id) references media (id);

create table items (
    id uuid primary key,
    name text not null,
    category_id uuid not null references categories (id),
    subcategory_id uuid references categories (id),
    description text not null default '',
    owner_user_id uuid references users (id),
    created_at timestamptz not null,
    check (subcategory_id is null or subcategory_id <> category_id)
);

create table persons (
    id uuid primary key,
    given_name text not null default '',
    family_name text not null default '',
    patronymic text not null default '',
    sex text not null check (sex in ('female', 'male', 'unknown')),
    created_at timestamptz not null,
    check (given_name <> '' or family_name <> '')
);

create table animals (
    id uuid primary key,
    category_id uuid not null references categories (id),
    breed_category_id uuid references categories (id),
    breed_text text not null default '',
    sex text not null check (sex in ('female', 'male', 'unknown')),
    nickname text not null default '',
    owner_user_id uuid references users (id),
    created_at timestamptz not null,
    check (breed_category_id is null or breed_category_id <> category_id)
);

create table places (
    id uuid primary key,
    category_id uuid not null references categories (id),
    subcategory_id uuid references categories (id),
    name text not null,
    description text not null default '',
    address text not null default '',
    lat double precision,
    lon double precision,
    created_by uuid not null references users (id),
    created_at timestamptz not null,
    check (subcategory_id is null or subcategory_id <> category_id),
    check ((lat is null) = (lon is null)),
    check (address <> '' or lat is not null)
);

create table place_tags (
    place_id uuid not null references places (id) on delete cascade,
    tag text not null,
    primary key (place_id, tag)
);

create table place_external_refs (
    id uuid primary key,
    place_id uuid not null references places (id) on delete cascade,
    source text not null check (source in ('yandex', '2gis', 'google', 'osm')),
    external_id text not null,
    url text not null default '',
    snapshot jsonb,
    unique (place_id, source, external_id)
);

create table events (
    id uuid primary key,
    category_id uuid not null references categories (id),
    subcategory_id uuid references categories (id),
    name text not null,
    description text not null default '',
    place_id uuid references places (id),
    address text not null default '',
    lat double precision,
    lon double precision,
    starts_at timestamptz not null,
    ends_at timestamptz,
    created_by uuid not null references users (id),
    created_at timestamptz not null,
    check (subcategory_id is null or subcategory_id <> category_id),
    check ((lat is null) = (lon is null)),
    check (place_id is not null or address <> '' or lat is not null),
    check (ends_at is null or ends_at > starts_at)
);

create table event_tags (
    event_id uuid not null references events (id) on delete cascade,
    tag text not null,
    primary key (event_id, tag)
);

-- Объявление автора о субъекте. subject_id не имеет внешнего ключа:
-- он указывает в одну из пяти таблиц согласно subject_kind.
create table publications (
    id uuid primary key,
    author_user_id uuid not null references users (id),
    author_visibility text not null check (author_visibility in ('public', 'internal')),
    intent text not null check (intent in ('seeking', 'found', 'marking')),
    subject_kind text not null check (subject_kind in ('item', 'person', 'animal', 'place', 'event')),
    subject_id uuid not null,
    body text not null default '',
    occurred_at timestamptz,
    lat double precision,
    lon double precision,
    place_label text not null default '',
    place_id uuid references places (id),
    status text not null check (status in ('draft', 'published', 'matched', 'closed', 'hidden')),
    created_at timestamptz not null,
    updated_at timestamptz not null,
    check ((lat is null) = (lon is null)),
    check (
        intent = 'marking'
        or subject_kind in ('item', 'person', 'animal')
    )
);

create index publications_author_idx on publications (author_user_id, created_at desc);
create index publications_subject_idx on publications (subject_kind, subject_id);

create table traits (
    id uuid primary key,
    subject_kind text not null check (subject_kind in ('item', 'person', 'animal')),
    subject_id uuid not null,
    key text not null,
    value text not null,
    visibility text not null check (visibility in ('public', 'internal')),
    sort_order integer not null default 0
);

create index traits_subject_idx on traits (subject_kind, subject_id);

create table attachments (
    parent_kind text not null check (parent_kind in ('item', 'person', 'animal', 'place', 'event', 'publication')),
    parent_id uuid not null,
    media_id uuid not null references media (id),
    visibility text not null check (visibility in ('public', 'internal')),
    sort_order integer not null default 0,
    primary key (parent_kind, parent_id, media_id)
);

create table follows (
    follower_user_id uuid not null references users (id),
    target_kind text not null check (target_kind in ('user', 'place', 'event')),
    target_id uuid not null,
    created_at timestamptz not null,
    primary key (follower_user_id, target_kind, target_id),
    check (target_kind <> 'user' or follower_user_id <> target_id)
);

create index follows_target_idx on follows (target_kind, target_id);

create table matches (
    id uuid primary key,
    seeking_publication_id uuid not null references publications (id),
    found_publication_id uuid not null references publications (id),
    status text not null check (status in ('proposed', 'confirmed', 'rejected')),
    created_at timestamptz not null,
    check (seeking_publication_id <> found_publication_id),
    unique (seeking_publication_id, found_publication_id)
);
