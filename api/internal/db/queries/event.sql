-- name: CreateEvent :one
INSERT INTO events (bucket_name, object_id, event_type)
VALUES (?, ?, ?)
RETURNING *;


-- name: ListEventsByBucket :many
SELECT id, bucket_name, object_id, event_type, event_time
FROM events
WHERE bucket_name = ?
ORDER BY event_time DESC
LIMIT ? OFFSET ?;
