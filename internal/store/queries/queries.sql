-- name: GetTenantByID :one
select * from tenants
where id = $1;

-- name: ListTenants :many
select * from tenants;

-- name: CreateTenant :one
insert into tenants (id, repo_url, path, values)
values ($1, $2, $3, $4)
returning *;

-- name: UpdateTenant :one
update tenants
set repo_url = $2, path = $3, values = $4
where id = $1
returning *;

-- name: DeleteTenant :one
delete from tenants
where id = $1
returning *;