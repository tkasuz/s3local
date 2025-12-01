-- name: CreateObject :one
INSERT INTO objects (
    bucket_name,
    key,
    data,
    size,
    etag,
    content_type,
    content_encoding,
    content_disposition,
    cache_control,
    expires,
    storage_class,
    server_side_encryption,
    version_id
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, bucket_name, key, size, etag, content_type, content_encoding,
          content_disposition, cache_control, expires, storage_class,
          server_side_encryption, version_id, created_at, updated_at;

-- name: GetObject :one
SELECT id, bucket_name, key, data, size, etag, content_type, content_encoding,
       content_disposition, cache_control, expires, storage_class,
       server_side_encryption, version_id, created_at, updated_at
FROM objects
WHERE bucket_name = ? AND key = ?;

-- name: GetObjectByID :one
SELECT sqlc.embed(objects)
FROM objects
WHERE id = ?;

-- name: GetObjectMetadata :one
SELECT id, bucket_name, key, size, etag, content_type, content_encoding,
       content_disposition, cache_control, expires, storage_class,
       server_side_encryption, version_id, created_at, updated_at
FROM objects
WHERE bucket_name = ? AND key = ?;

-- name: UpdateObject :exec
UPDATE objects
SET data = ?,
    size = ?,
    etag = ?,
    content_type = ?,
    content_encoding = ?,
    content_disposition = ?,
    cache_control = ?,
    expires = ?,
    storage_class = ?,
    server_side_encryption = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE bucket_name = ? AND key = ?;

-- name: DeleteObject :exec
DELETE FROM objects
WHERE bucket_name = ? AND key = ?;

-- name: ObjectExists :one
SELECT COUNT(*) > 0 as object_exists
FROM objects
WHERE bucket_name = ? AND key = ?;

-- name: ListObjects :many
SELECT id, bucket_name, key, size, etag, content_type, content_encoding,
       content_disposition, cache_control, expires, storage_class,
       server_side_encryption, version_id, created_at, updated_at
FROM objects
WHERE bucket_name = ?
  AND (? = '' OR key >= ?)
  AND (? = '' OR key LIKE ? || '%')
ORDER BY key ASC
LIMIT ?;

-- name: ListObjectsWithDelimiter :many
SELECT id, bucket_name, key, size, etag, content_type, content_encoding,
       content_disposition, cache_control, expires, storage_class,
       server_side_encryption, version_id, created_at, updated_at
FROM objects
WHERE bucket_name = ?
  AND (? = '' OR key >= ?)
  AND (? = '' OR key LIKE ? || '%')
ORDER BY key ASC;

-- name: CopyObject :one
INSERT INTO objects (
    bucket_name,
    key,
    data,
    size,
    etag,
    content_type,
    content_encoding,
    content_disposition,
    cache_control,
    expires,
    storage_class,
    server_side_encryption
)
SELECT ?, ?, o.data, o.size, o.etag, o.content_type, o.content_encoding,
       o.content_disposition, o.cache_control, o.expires, ?, ?
FROM objects o
WHERE o.bucket_name = ? AND o.key = ?
RETURNING id, bucket_name, key, size, etag, content_type, content_encoding,
          content_disposition, cache_control, expires, storage_class,
          server_side_encryption, version_id, created_at, updated_at;

-- name: GetObjectID :one
SELECT id
FROM objects
WHERE bucket_name = ? AND key = ?;

-- Object Metadata queries
-- name: CreateObjectMetadata :exec
INSERT INTO object_metadata (object_id, key, value)
VALUES (?, ?, ?);

-- name: GetObjectMetadataByObjectID :many
SELECT key, value
FROM object_metadata
WHERE object_id = ?;

-- name: DeleteObjectMetadata :exec
DELETE FROM object_metadata
WHERE object_id = ?;

-- Object Tags queries
-- name: CreateObjectTag :exec
INSERT INTO object_tags (object_id, key, value)
VALUES (?, ?, ?);

-- name: GetObjectTags :many
SELECT key, value
FROM object_tags
WHERE object_id = ?;

-- name: DeleteObjectTags :exec
DELETE FROM object_tags
WHERE object_id = ?;

-- name: DeleteAllObjectTags :exec
DELETE FROM object_tags
WHERE object_id = ?;
