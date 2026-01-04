-- +goose Up
create table if not exists posts (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade,
  title text not null,
  body text not null,
  created_at timestamptz not null default now()
);

create table if not exists post_votes (
  post_id uuid not null references posts(id) on delete cascade,
  user_id uuid not null references users(id) on delete cascade,
  value smallint not null check (value in (-1, 1)),
  created_at timestamptz not null default now(),
  primary key (post_id, user_id)
);

-- +goose Down
drop table if exists post_votes;
drop table if exists posts;
