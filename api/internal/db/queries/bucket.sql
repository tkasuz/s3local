-- name: CreateBucket :exec
INSERT INTO buckets (name, region)
VALUES (?, ?);

-- name: GetBucket :one
SELECT name, region, created_at
FROM buckets
WHERE name = ?;

-- name: ListBuckets :many
SELECT name, region, created_at
FROM buckets
ORDER BY created_at ASC;

-- name: ListBucketsFiltered :many
SELECT name, region, created_at
FROM buckets
WHERE (sqlc.narg('region') IS NULL OR region = sqlc.narg('region'))
  AND (sqlc.narg('prefix') IS NULL OR name LIKE sqlc.narg('prefix') || '%')
ORDER BY created_at ASC
LIMIT sqlc.arg('limit');

-- name: DeleteBucket :exec
DELETE FROM buckets
WHERE name = ?;

-- name: BucketExists :one
SELECT COUNT(*) > 0 as bucket_exists
FROM buckets
WHERE name = ?;

-- name: CountObjectsInBucket :one
SELECT COUNT(*) as count
FROM objects
WHERE bucket_name = ?;

-- name: CreateBucketTag :exec
INSERT INTO bucket_tags (bucket_name, key, value)
VALUES (?, ?, ?)
ON CONFLICT(bucket_name, key) DO UPDATE SET value = excluded.value;

-- name: GetBucketTags :many
SELECT key, value
FROM bucket_tags
WHERE bucket_name = ?
ORDER BY key ASC;

-- name: DeleteBucketTags :exec
DELETE FROM bucket_tags
WHERE bucket_name = ?;

-- name: DeleteBucketTag :exec
DELETE FROM bucket_tags
WHERE bucket_name = ? AND key = ?;

-- name: PutBucketPolicy :exec
INSERT INTO bucket_policies (bucket_name, policy)
VALUES (?, ?)
ON CONFLICT(bucket_name) DO UPDATE SET
    policy = excluded.policy,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetBucketPolicy :one
SELECT policy, created_at, updated_at
FROM bucket_policies
WHERE bucket_name = ?;

-- name: DeleteBucketPolicy :exec
DELETE FROM bucket_policies
WHERE bucket_name = ?;

-- name: BucketPolicyExists :one
SELECT COUNT(*) > 0 as policy_exists
FROM bucket_policies
WHERE bucket_name = ?;