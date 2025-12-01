resource "aws_s3_bucket" "example" {
  bucket = "mytestbucket1"

  tags = {
    Name        = "My bucket"
    Environment = "Dev"
  }
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.example.id
  key    = "new_object_key"
  source = ".gitignore"

  etag = filemd5(".gitignore")
  tags = {
    Name        = "My object"
    Environment = "Dev"
  }
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.example.id

  lambda_function {
    lambda_function_arn = "http://lambda:8081/2015-03-31/functions/function/invocations"
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/"
    filter_suffix       = ".log"
  }
}
