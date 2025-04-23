-- name: CreateVerificationCode :one
INSERT INTO VerificationCodes (id, user_id, code, expires_at, created_at)
VALUES (uuid_generate_v4(), $1, $2, $3, now())
RETURNING *;

-- name: IsVerificationCodeValid :one
SELECT EXISTS (
  SELECT 1
  FROM VerificationCodes
  WHERE user_id = $1
  AND code = $2
  AND expires_at > now()
) AS valid_code_exists;

-- name: DeleteVerificationCodes :exec
DELETE FROM VerificationCodes
WHERE user_id = $1;
