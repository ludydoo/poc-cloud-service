create table tenants
(
    id       text primary key,
    repo_url text  not null,
    path     text  not null,
    values   jsonb not null
)