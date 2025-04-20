-- name: FindSiteByID :one
SELECT * FROM sites WHERE id = $1;

-- name: CreateSite :one
INSERT INTO sites (id, user_id, site_url, created_at, updated_at)
VALUES (uuid_generate_v4(), $1, $2, now(), now())
RETURNING *;

-- name: FindSiteByUserIDAndURL :one
SELECT * FROM sites
WHERE user_id = $1 AND site_url = $2;

-- name: ListSitesByUserID :many
SELECT * FROM sites
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateSiteURL :one
UPDATE sites
SET site_url = $1, updated_at = now()
WHERE id = $2
RETURNING *;

-- name: DeleteSite :exec
DELETE FROM sites
WHERE id = $1 AND user_id = $2;

-- name: GetSiteCount :one
SELECT COUNT(*) FROM sites
WHERE user_id = $1;