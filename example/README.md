# S3Local Example with ElasticMQ and Lambda

This example demonstrates how to set up s3local with ElasticMQ (local SQS) and a Python Lambda function to process S3 event notifications.

## Services

- **s3local** (port 8080): S3-compatible storage service
- **elasticmq** (port 9324): Local SQS service for message queuing
- **lambda** (port 8081): Python Lambda function to receive and process S3 events

## Prerequisites

- Docker and Docker Compose installed
- S3Local API built (the Dockerfile in `../api`)

## Quick Start

1. Start all services:
   ```bash
   docker-compose up -d
   ```

2. Check service health:
   ```bash
   docker-compose ps
   ```

3. View logs:
   ```bash
   # All services
   docker-compose logs -f

   # Specific service
   docker-compose logs -f lambda
   ```

## Testing S3 Event Notifications

### 1. Create a bucket
```bash
aws s3 mb s3://test-bucket --endpoint-url http://localhost:8080
```

### 2. Configure S3 notification to Lambda
You can configure s3local to send notifications to the Lambda function by inserting a notification configuration:

```bash
# Using sqlite3 or your preferred method to insert into the notifications table
sqlite3 /path/to/s3local.db <<EOF
INSERT INTO notifications (bucket_name, event_type, destination_type, destination_arn, enabled)
VALUES ('test-bucket', 's3:ObjectCreated:Put', 'lambda', 'http://lambda:8080/2015-03-31/functions/function/invocations', 1);
EOF
```

### 3. Upload an object to trigger the notification
```bash
echo "Hello World" > test.txt
aws s3 cp test.txt s3://test-bucket/ --endpoint-url http://localhost:8080
```

### 4. Check Lambda logs to see the event
```bash
docker-compose logs lambda
```

You should see structured logs from AWS Lambda Powertools showing:
- Event metadata (version, source, name, time)
- Bucket and object details (name, key, size, etag)
- Event processing information

## Lambda Function

The Lambda function uses AWS Lambda Powertools for structured logging and event handling. It:
- Receives S3 event notifications via HTTP POST
- Parses the event using Lambda Powertools Logger
- Logs structured event information with context
- Returns success/failure status

The function exposes two endpoints:
- `POST /2015-03-31/functions/function/invocations` - Lambda invocation endpoint
- `GET /health` - Health check endpoint

## ElasticMQ Configuration

ElasticMQ is configured with a queue named `s3-notifications`. You can:

- Access the SQS API at `http://localhost:9324`
- View stats at `http://localhost:9325`
- Modify the configuration in `elasticmq.conf`

## Customization

### Lambda Handler

Edit `lambda/handler.py` to customize event processing:

```python
# Process based on event type
if 'ObjectCreated' in event_name:
    logger.info("Object created", extra={
        "bucket": bucket_name,
        "key": object_key,
        "size": object_size
    })
    # Add your custom logic here
```

### Notification Configuration

Supported event types:
- `s3:ObjectCreated:Put`
- `s3:ObjectCreated:Post`
- `s3:ObjectCreated:Copy`
- `s3:ObjectRemoved:Delete`

Destination types:
- `lambda` - HTTP endpoint (Lambda function)
- `sqs` - SQS queue URL
- `sns` - SNS topic URL

## Cleanup

Stop and remove all containers:
```bash
docker-compose down
```

Remove volumes (deletes all data):
```bash
docker-compose down -v
```

## Troubleshooting

### Lambda not receiving events
1. Check that the notification is enabled in the database
2. Verify the `destination_arn` points to `http://lambda:8080/2015-03-31/functions/function/invocations`
3. Check Lambda logs: `docker-compose logs lambda`
4. Check s3local logs: `docker-compose logs s3local`

### Connection issues
Ensure all services are healthy:
```bash
docker-compose ps
```

All services should show "healthy" status.
