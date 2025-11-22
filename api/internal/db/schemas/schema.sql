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
