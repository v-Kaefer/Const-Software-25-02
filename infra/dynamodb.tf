resource "aws_dynamodb_table" "grupo-l-dynamodb-table" {
  name         = "GrupoLConstSoftSprint1DynamoDB"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "name"
  range_key    = "date"

  attribute {
    name = "name"
    type = "S"
  }

  attribute {
    name = "date"
    type = "N"
  }

  tags = {
    Name = "grupo-l-sprint1"
  }
}