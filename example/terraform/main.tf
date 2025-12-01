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
