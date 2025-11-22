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