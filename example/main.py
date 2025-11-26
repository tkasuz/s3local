import boto3
from botocore.client import Config
from botocore.exceptions import ClientError


def create_s3_client():
    """Create a boto3 S3 client configured for local endpoint."""
    return boto3.client(
        's3',
        endpoint_url='http://localhost:8080',
        aws_access_key_id='test',
        aws_secret_access_key='test',
        config=Config(signature_version='s3v4'),
        region_name='us-east-1'
    )


def test_create_bucket(s3_client, bucket_name):
    """Test bucket creation."""
    print(f"\n=== Testing CreateBucket: {bucket_name} ===")
    try:
        response = s3_client.create_bucket(Bucket=bucket_name)
        print("✓ Bucket created successfully")
        print(f"  Location: {response.get('Location')}")
        return True
    except ClientError as e:
        print(f"✗ Failed to create bucket: {e}")
        return False


def test_head_bucket(s3_client, bucket_name):
    """Test bucket head operation."""
    print(f"\n=== Testing HeadBucket: {bucket_name} ===")
    try:
        response = s3_client.head_bucket(Bucket=bucket_name)
        print("✓ Bucket exists")
        print(f"  Region: {response.get('BucketRegion', 'N/A')}")
        return True
    except ClientError as e:
        print(f"✗ Failed to head bucket: {e}")
        return False


def test_list_buckets(s3_client):
    """Test listing buckets."""
    print("\n=== Testing ListBuckets ===")
    try:
        response = s3_client.list_buckets()
        buckets = response.get('Buckets', [])
        print(f"✓ Found {len(buckets)} bucket(s)")
        for bucket in buckets:
            print(f"  - {bucket['Name']} (Created: {bucket['CreationDate']})")
        return True
    except ClientError as e:
        print(f"✗ Failed to list buckets: {e}")
        return False


def test_put_object(s3_client, bucket_name, key, content):
    """Test object upload."""
    print(f"\n=== Testing PutObject: {bucket_name}/{key} ===")
    try:
        response = s3_client.put_object(
            Bucket=bucket_name,
            Key=key,
            Body=content.encode('utf-8')
        )
        print("✓ Object uploaded successfully")
        print(f"  ETag: {response.get('ETag')}")
        return True
    except ClientError as e:
        print(f"✗ Failed to put object: {e}")
        return False


def test_get_object(s3_client, bucket_name, key):
    """Test object retrieval."""
    print(f"\n=== Testing GetObject: {bucket_name}/{key} ===")
    try:
        response = s3_client.get_object(Bucket=bucket_name, Key=key)
        content = response['Body'].read().decode('utf-8')
        print("✓ Object retrieved successfully")
        print(f"  Content: {content}")
        print(f"  ContentType: {response.get('ContentType', 'N/A')}")
        print(f"  ContentLength: {response.get('ContentLength', 0)}")
        return True
    except ClientError as e:
        print(f"✗ Failed to get object: {e}")
        return False


def test_list_objects_v2(s3_client, bucket_name):
    """Test listing objects."""
    print(f"\n=== Testing ListObjectsV2: {bucket_name} ===")
    try:
        response = s3_client.list_objects_v2(Bucket=bucket_name)
        contents = response.get('Contents', [])
        print(f"✓ Found {len(contents)} object(s)")
        for obj in contents:
            print(f"  - {obj['Key']} ({obj['Size']} bytes)")
        return True
    except ClientError as e:
        print(f"✗ Failed to list objects: {e}")
        return False


def test_delete_object(s3_client, bucket_name, key):
    """Test object deletion."""
    print(f"\n=== Testing DeleteObject: {bucket_name}/{key} ===")
    try:
        s3_client.delete_object(Bucket=bucket_name, Key=key)
        print("✓ Object deleted successfully")
        return True
    except ClientError as e:
        print(f"✗ Failed to delete object: {e}")
        return False


def test_delete_bucket(s3_client, bucket_name):
    """Test bucket deletion."""
    print(f"\n=== Testing DeleteBucket: {bucket_name} ===")
    try:
        s3_client.delete_bucket(Bucket=bucket_name)
        print("✓ Bucket deleted successfully")
        return True
    except ClientError as e:
        print(f"✗ Failed to delete bucket: {e}")
        return False


def main():
    """Run all S3 operation tests."""
    print("=" * 60)
    print("S3Local Test Suite")
    print("=" * 60)
    
    s3_client = create_s3_client()
    bucket_name = "test-bucket"
    object_key = "test-file.txt"
    object_content = "Hello from S3Local!"
    
    results = []
    
    # Test bucket operations
    results.append(("CreateBucket", test_create_bucket(s3_client, bucket_name)))
    results.append(("HeadBucket", test_head_bucket(s3_client, bucket_name)))
    results.append(("ListBuckets", test_list_buckets(s3_client)))
    
    # Test object operations
    results.append(("PutObject", test_put_object(s3_client, bucket_name, object_key, object_content)))
    results.append(("GetObject", test_get_object(s3_client, bucket_name, object_key)))
    results.append(("ListObjectsV2", test_list_objects_v2(s3_client, bucket_name)))
    results.append(("DeleteObject", test_delete_object(s3_client, bucket_name, object_key)))
    
    # Test bucket deletion
    results.append(("DeleteBucket", test_delete_bucket(s3_client, bucket_name)))
    
    # Summary
    print("\n" + "=" * 60)
    print("Test Summary")
    print("=" * 60)
    passed = sum(1 for _, result in results if result)
    total = len(results)
    print(f"Passed: {passed}/{total}")
    for test_name, result in results:
        status = "✓" if result else "✗"
        print(f"  {status} {test_name}")
    
    if passed == total:
        print("\n🎉 All tests passed!")
    else:
        print(f"\n⚠️  {total - passed} test(s) failed")


if __name__ == "__main__":
    main()