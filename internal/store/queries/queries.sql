-- name: GetTenantByID :one
select * from tenants
where id = $1;

-- name: ListTenants :many
select * from tenants
order by id;

-- name: CreateTenant :one
insert into tenants (id, repo_url, path, target_revision, values)
values ($1, $2, $3, $4, $5)
returning *;

-- name: UpdateTenant :one
update tenants
set repo_url = $2, path = $3, target_revision = $4, values = $5
where id = $1
returning *;

-- name: DeleteTenant :one
delete from tenants
where id = $1
returning *;