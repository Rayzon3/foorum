-- +goose Up
alter table users
  add column if not exists username text;

create unique index if not exists users_username_unique
  on users (username);

-- +goose Down
drop index if exists users_username_unique;
alter table users drop column if exists username;
