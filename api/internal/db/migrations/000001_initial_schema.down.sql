-- Drop trigger
DROP TRIGGER IF EXISTS create_notification_jobs_on_event_insert;

-- Drop tables in reverse order of creation (respecting foreign key dependencies)
DROP TABLE IF EXISTS notification_jobs;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS object_tags;
DROP TABLE IF EXISTS object_metadata;
DROP TABLE IF EXISTS objects;
DROP TABLE IF EXISTS bucket_policies;
DROP TABLE IF EXISTS bucket_tags;
DROP TABLE IF EXISTS buckets;
