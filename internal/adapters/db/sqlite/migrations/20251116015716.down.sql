-- reverse: create index "idx_object_tags_object_id" to table: "object_tags"
DROP INDEX `idx_object_tags_object_id`;
-- reverse: create index "object_tags_object_id_key" to table: "object_tags"
DROP INDEX `object_tags_object_id_key`;
-- reverse: create "object_tags" table
DROP TABLE `object_tags`;
-- reverse: create index "idx_object_metadata_object_id" to table: "object_metadata"
DROP INDEX `idx_object_metadata_object_id`;
-- reverse: create index "object_metadata_object_id_key" to table: "object_metadata"
DROP INDEX `object_metadata_object_id_key`;
-- reverse: create "object_metadata" table
DROP TABLE `object_metadata`;
-- reverse: create index "idx_objects_updated_at" to table: "objects"
DROP INDEX `idx_objects_updated_at`;
-- reverse: create index "idx_objects_bucket_key" to table: "objects"
DROP INDEX `idx_objects_bucket_key`;
-- reverse: create index "idx_objects_key" to table: "objects"
DROP INDEX `idx_objects_key`;
-- reverse: create index "idx_objects_bucket_name" to table: "objects"
DROP INDEX `idx_objects_bucket_name`;
-- reverse: create index "objects_bucket_name_key" to table: "objects"
DROP INDEX `objects_bucket_name_key`;
-- reverse: create "objects" table
DROP TABLE `objects`;
-- reverse: create index "idx_buckets_created_at" to table: "buckets"
DROP INDEX `idx_buckets_created_at`;
-- reverse: create "buckets" table
DROP TABLE `buckets`;
