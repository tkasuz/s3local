#!/usr/bin/env python3
"""
S3 Event Notification Handler using AWS Lambda Powertools
Receives S3 event notifications and logs them with structured logging.
"""

from urllib.parse import unquote_plus

from aws_lambda_powertools import Logger
from aws_lambda_powertools.utilities.data_classes import S3Event, event_source

logger = Logger(service="s3-event-handler")


@event_source(data_class=S3Event)
@logger.inject_lambda_context(log_event=True)
def handler(event: S3Event, context: dict) -> dict:
    """
    Lambda handler for S3 events.

    Args:
        event: S3 event notification
        context: Lambda context

    Returns:
        Response dict
    """
    logger.info("Received S3 event notification")
    logger.info("Raw event", extra={"raw_event": event.raw_event})

    # Process each record
    logger.info(f"Processing {len(event.records)} record(s)")

    for idx, record in enumerate(event.records):
        try:
            logger.info(f"Processing record {idx + 1}/{len(event.records)}")

            # Extract event metadata
            event_name = record.event_name
            event_time = record.event_time
            event_source = record.event_source
            aws_region = record.aws_region

            # Extract S3 data
            bucket_name = record.s3.bucket.name
            bucket_arn = record.s3.bucket.arn

            # Decode URL-encoded object key
            object_key = unquote_plus(record.s3.get_object.key)
            object_size = record.s3.get_object.size
            object_etag = record.s3.get_object.e_tag
            object_version_id = record.s3.get_object.version_id
            object_sequencer = record.s3.get_object.sequencer

            # Log structured event information
            logger.info(
                "S3 Event Details",
                extra={
                    "event_name": event_name,
                    "event_time": event_time,
                    "event_source": event_source,
                    "aws_region": aws_region,
                    "bucket_name": bucket_name,
                    "bucket_arn": bucket_arn,
                    "object_key": object_key,
                    "object_size": object_size,
                    "object_etag": object_etag,
                    "object_version_id": object_version_id,
                    "object_sequencer": object_sequencer,
                }
            )

            # Process based on event type
            if "ObjectCreated" in event_name:
                logger.info(
                    "Object created event",
                    extra={
                        "bucket": bucket_name,
                        "key": object_key,
                        "size": object_size,
                        "etag": object_etag,
                    }
                )
                # Add your custom logic here for object creation
                # Example: process_created_object(bucket_name, object_key)

            elif "ObjectRemoved" in event_name:
                logger.info(
                    "Object removed event",
                    extra={
                        "bucket": bucket_name,
                        "key": object_key,
                    }
                )
                # Add your custom logic here for object deletion
                # Example: process_deleted_object(bucket_name, object_key)

            else:
                logger.info(f"Unhandled event type: {event_name}")

        except Exception as e:
            logger.exception(
                "Error processing S3 event record",
                extra={
                    "record_index": idx,
                    "error": str(e),
                }
            )
            # Continue processing other records even if one fails
            continue

    return {
        "statusCode": 200,
        "body": f"Successfully processed {len(event.records)} record(s)"
    }

