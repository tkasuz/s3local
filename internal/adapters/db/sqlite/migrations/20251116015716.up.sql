-- create "buckets" table
CREATE TABLE `buckets` (
  `name` text NOT NULL,
  `region` text NOT NULL DEFAULT 'us-east-1',
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`name`)
);
-- create index "idx_buckets_created_at" to table: "buckets"
CREATE INDEX `idx_buckets_created_at` ON `buckets` (`created_at`);
-- create "objects" table
CREATE TABLE `objects` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `bucket_name` text NOT NULL,
  `key` text NOT NULL,
  `data` blob NOT NULL,
  `size` integer NOT NULL,
  `etag` text NOT NULL,
  `content_type` text NOT NULL DEFAULT 'application/octet-stream',
  `content_encoding` text NULL,
  `content_disposition` text NULL,
  `cache_control` text NULL,
  `expires` datetime NULL,
  `storage_class` text NOT NULL DEFAULT 'STANDARD',
  `server_side_encryption` text NULL,
  `version_id` text NULL,
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  CONSTRAINT `0` FOREIGN KEY (`bucket_name`) REFERENCES `buckets` (`name`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "objects_bucket_name_key" to table: "objects"
CREATE UNIQUE INDEX `objects_bucket_name_key` ON `objects` (`bucket_name`, `key`);
-- create index "idx_objects_bucket_name" to table: "objects"
CREATE INDEX `idx_objects_bucket_name` ON `objects` (`bucket_name`);
-- create index "idx_objects_key" to table: "objects"
CREATE INDEX `idx_objects_key` ON `objects` (`key`);
-- create index "idx_objects_bucket_key" to table: "objects"
CREATE INDEX `idx_objects_bucket_key` ON `objects` (`bucket_name`, `key`);
-- create index "idx_objects_updated_at" to table: "objects"
CREATE INDEX `idx_objects_updated_at` ON `objects` (`updated_at`);
-- create "object_metadata" table
CREATE TABLE `object_metadata` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `object_id` integer NOT NULL,
  `key` text NOT NULL,
  `value` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`object_id`) REFERENCES `objects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "object_metadata_object_id_key" to table: "object_metadata"
CREATE UNIQUE INDEX `object_metadata_object_id_key` ON `object_metadata` (`object_id`, `key`);
-- create index "idx_object_metadata_object_id" to table: "object_metadata"
CREATE INDEX `idx_object_metadata_object_id` ON `object_metadata` (`object_id`);
-- create "object_tags" table
CREATE TABLE `object_tags` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `object_id` integer NOT NULL,
  `key` text NOT NULL,
  `value` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`object_id`) REFERENCES `objects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "object_tags_object_id_key" to table: "object_tags"
CREATE UNIQUE INDEX `object_tags_object_id_key` ON `object_tags` (`object_id`, `key`);
-- create index "idx_object_tags_object_id" to table: "object_tags"
CREATE INDEX `idx_object_tags_object_id` ON `object_tags` (`object_id`);
