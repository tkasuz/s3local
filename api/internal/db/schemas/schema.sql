-- Buckets table
CREATE TABLE IF NOT EXISTS buckets (
    name TEXT PRIMARY KEY NOT NULL,
    region TEXT NOT NULL DEFAULT 'us-east-1',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_buckets_created_at ON buckets(created_at);

-- Objects table
CREATE TABLE IF NOT EXISTS objects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bucket_name TEXT NOT NULL,
    key TEXT NOT NULL,
    data BLOB NOT NULL,
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
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_events_event_time ON events(event_time);
