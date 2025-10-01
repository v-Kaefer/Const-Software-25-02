resource "aws_s3_bucket" "localstack-bucket" {
  bucket = "my-localstack-bucket"

  tags = {
    Name = "grupo-l-sprint1"
  }
}