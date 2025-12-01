-- name: CreateNotification :one
INSERT INTO notifications (bucket_name, event_type, destination_type, destination_arn, filter_prefix, filter_suffix, enabled)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetNotification :one
SELECT id, bucket_name, event_type, destination_type, destination_arn, filter_prefix, filter_suffix, enabled, created_at, updated_at
FROM notifications
WHERE id = ?;

-- name: ListNotificationsByBucket :many
SELECT id, bucket_name, event_type, destination_type, destination_arn, filter_prefix, filter_suffix, enabled, created_at, updated_at
FROM notifications
WHERE bucket_name = ?
ORDER BY created_at DESC;

-- name: ListEnabledNotificationsByBucket :many
SELECT id, bucket_name, event_type, destination_type, destination_arn, filter_prefix, filter_suffix, enabled, created_at, updated_at
FROM notifications
WHERE bucket_name = ? AND enabled = 1
ORDER BY created_at DESC;

-- name: ListNotificationsByEventType :many
SELECT id, bucket_name, event_type, destination_type, destination_arn, filter_prefix, filter_suffix, enabled, created_at, updated_at
FROM notifications
WHERE bucket_name = ? AND event_type = ? AND enabled = 1;

-- name: UpdateNotification :exec
UPDATE notifications
SET event_type = ?, destination_type = ?, destination_arn = ?, filter_prefix = ?, filter_suffix = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateNotificationEnabled :exec
UPDATE notifications
SET enabled = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = ?;

-- name: ListPendingNotificationJobs :many
SELECT
    sqlc.embed(notification_jobs),
    sqlc.embed(events),
    sqlc.embed(notifications)
FROM notification_jobs
JOIN events ON notification_jobs.event_id = events.id
JOIN notifications ON notification_jobs.notification_id = notifications.id
WHERE notification_jobs.status = 'pending'
ORDER BY notification_jobs.created_at ASC;

-- name: UpdateNotificationJobStatus :exec
UPDATE notification_jobs
SET status = ?,
    attempts = ?,
    error_message = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
