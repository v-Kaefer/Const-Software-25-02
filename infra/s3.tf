resource "aws_s3_bucket" "grupo_l_bucket" {
  bucket = "grupo-l-terraform"

  tags = {
    Name = "grupo-l-sprint1"
  }
}