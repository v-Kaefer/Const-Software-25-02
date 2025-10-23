provider "aws" {

  access_key = "test"
  secret_key = "test"
  region     = "us-east-1"

  # only required for non virtual hosted-style endpoint use case.
  # https://registry.terraform.io/providers/hashicorp/aws/latest/docs#s3_use_path_style
  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    apigateway      = "http://localhost:4566"
    apigatewayv2    = "http://localhost:4566"
    cloudformation  = "http://localhost:4566"
    cloudwatch      = "http://localhost:4566"
    cognitoidentity = "http://localhost:4566" # ← Identity Pool
    cognitoidp      = "http://localhost:4566" # ← User Pool
    dynamodb        = "http://localhost:4566"
    ec2             = "http://localhost:4566"
    es              = "http://localhost:4566"
    elasticache     = "http://localhost:4566"
    firehose        = "http://localhost:4566"
    iam             = "http://localhost:4566"
    kinesis         = "http://localhost:4566"
    lambda          = "http://localhost:4566"
    rds             = "http://localhost:4566"
    redshift        = "http://localhost:4566"
    route53         = "http://localhost:4566"
    s3              = "http://s3.localhost.localstack.cloud:4566"
    secretsmanager  = "http://localhost:4566"
    ses             = "http://localhost:4566"
    sns             = "http://localhost:4566"
    sqs             = "http://localhost:4566"
    ssm             = "http://localhost:4566"
    stepfunctions   = "http://localhost:4566"
    sts             = "http://localhost:4566"
  }
}

data "aws_caller_identity" "current" {}
output "is_localstack" {
  value = data.aws_caller_identity.current.id == "000000000000"
}