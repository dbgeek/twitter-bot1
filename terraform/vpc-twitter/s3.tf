resource "aws_s3_bucket" "twitter_lambda" {
  bucket = "ba78-twitter-lambda"
  acl    = "private"

  tags = {
    Name = "ba78-twitter-lambda"
  }
}

output "twitter-lambda-arn" {
  value = aws_s3_bucket.twitter_lambda.arn
}