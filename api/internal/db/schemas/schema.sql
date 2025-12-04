-- Buckets table
CREATE TABLE IF NOT EXISTS buckets (
    name TEXT PRIMARY KEY NOT NULL,
    region TEXT NOT NULL DEFAULT 'us-east-1',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_buckets_created_at ON buckets(created_at);

-- Bucket tags table
CREATE TABLE IF NOT EXISTS bucket_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bucket_name TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY (bucket_name) REFERENCES buckets(name) ON DELETE CASCADE,
    UNIQUE(bucket_name, key)
);

CREATE INDEX IF NOT EXISTS idx_bucket_tags_bucket_name ON bucket_tags(bucket_name);

-- Bucket policies table
CREATE TABLE IF NOT EXISTS bucket_policies (
    bucket_name TEXT PRIMARY KEY NOT NULL,
    policy TEXT NOT NULL, -- JSON policy document stored as TEXT
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bucket_name) REFERENCES buckets(name) ON DELETE CASCADE
);

-- Objects table
CREATE TABLE IF NOT EXISTS objects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bucket_name TEXT NOT NULL,
    key TEXT NOT NULL,
    data BLOB,
    size INTEGER NOT NULL,
    etag TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    content_encoding TEXT,
    content_disposition TEXT,
    cache_control TEXT,
    expires DATETIME,
    storage_class TEXT NOT NULL DEFAULT 'STANDARD',
    server_side_encryption TEXT,
    version_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(bucket_name, key),
    FOREIGN KEY (bucket_name) REFERENCES buckets(name) ON DELETE CASCADE
);

-- Indexes for objects table
CREATE INDEX IF NOT EXISTS idx_objects_bucket_name ON objects(bucket_name);
CREATE INDEX IF NOT EXISTS idx_objects_key ON objects(key);
CREATE INDEX IF NOT EXISTS idx_objects_bucket_key ON objects(bucket_name, key);
CREATE INDEX IF NOT EXISTS idx_objects_updated_at ON objects(updated_at);

-- Object metadata table (for custom user metadata)
CREATE TABLE IF NOT EXISTS object_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    object_id INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY (object_id) REFERENCES objects(id) ON DELETE CASCADE,
    UNIQUE(object_id, key)
);

CREATE INDEX IF NOT EXISTS idx_object_metadata_object_id ON object_metadata(object_id);

-- Object tags table
CREATE TABLE IF NOT EXISTS object_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    object_id INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY (object_id) REFERENCES objects(id) ON DELETE CASCADE,
    UNIQUE(object_id, key)
);

CREATE INDEX IF NOT EXISTS idx_object_tags_object_id ON object_tags(object_id);

-- S3 notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bucket_name TEXT NOT NULL,
    event_type TEXT NOT NULL, -- e.g., 's3:ObjectCreated:Put', 's3:ObjectRemoved:Delete'
    destination_type TEXT NOT NULL, -- 'sns', 'sqs', 'lambda'
    destination_arn TEXT NOT NULL,
    filter_prefix TEXT,
    filter_suffix TEXT,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bucket_name) REFERENCES buckets(name) ON DELETE CASCADE
);

-- Indexes for notifications table
CREATE INDEX IF NOT EXISTS idx_notifications_bucket_name ON notifications(bucket_name);
CREATE INDEX IF NOT EXISTS idx_notifications_event_type ON notifications(event_type);
CREATE INDEX IF NOT EXISTS idx_notifications_enabled ON notifications(enabled);

-- Notification events log table (for tracking sent notifications)
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bucket_name TEXT NOT NULL,
    object_id INTEGER NOT NULL,
    event_type TEXT NOT NULL,
    event_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bucket_name) REFERENCES buckets(name) ON DELETE CASCADE,
    FOREIGN KEY (object_id) REFERENCES objects(id) ON DELETE CASCADE
);

-- Indexes for notification events table
CREATE INDEX IF NOT EXISTS idx_events_bucket_name ON events(bucket_name);
CREATE INDEX IF NOT EXISTS idx_events_object_id ON events(object_id);
CREATE INDEX IF NOT EXISTS idx_events_event_time ON events(event_time);

-- Notification jobs table (for queuing notifications to be sent)
CREATE TABLE IF NOT EXISTS notification_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL,
    notification_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'completed', 'failed'
    attempts INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
);

-- Indexes for notification jobs table
CREATE INDEX IF NOT EXISTS idx_notification_jobs_event_id ON notification_jobs(event_id);
CREATE INDEX IF NOT EXISTS idx_notification_jobs_notification_id ON notification_jobs(notification_id);
CREATE INDEX IF NOT EXISTS idx_notification_jobs_status ON notification_jobs(status);
CREATE INDEX IF NOT EXISTS idx_notification_jobs_created_at ON notification_jobs(created_at);

-- Trigger to automatically create notification jobs when an event is inserted
CREATE TRIGGER IF NOT EXISTS create_notification_jobs_on_event_insert
AFTER INSERT ON events
FOR EACH ROW
BEGIN
    INSERT INTO notification_jobs (event_id, notification_id)
    SELECT
        NEW.id,
        n.id
    FROM notifications n
    WHERE n.bucket_name = NEW.bucket_name
      AND n.event_type = NEW.event_type
      AND n.enabled = 1;
END;
