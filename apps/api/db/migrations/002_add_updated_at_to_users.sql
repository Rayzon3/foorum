-- +goose Up
alter table users
  add column if not exists updated_at timestamptz not null default now();

-- +goose StatementBegin
create or replace function set_updated_at()
returns trigger as $$
begin
  NEW.updated_at = now();
  return NEW;
end;
$$ language plpgsql;
-- +goose StatementEnd

drop trigger if exists users_set_updated_at on users;
create trigger users_set_updated_at
before update on users
for each row
execute function set_updated_at();

-- +goose Down
drop trigger if exists users_set_updated_at on users;
-- +goose StatementBegin
drop function if exists set_updated_at();
-- +goose StatementEnd
alter table users drop column if exists updated_at;
