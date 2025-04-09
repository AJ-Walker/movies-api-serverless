provider "aws" {
  region = var.aws_region
}

locals {
  files      = fileset("${path.module}/${var.local_images_folder}", "*")
  movie_data = jsondecode(file("${path.module}/movies.json"))
}

# S3
resource "aws_s3_bucket" "movies_rest_api_bucket" {
  bucket = var.bucket_name

  tags = {
    Name        = "Movies REST API"
    Environment = "Dev"
  }
}

resource "aws_s3_object" "movies_rest_api_images" {
  for_each     = local.files
  bucket       = aws_s3_bucket.movies_rest_api_bucket.id
  key          = "${var.s3_images_prefix}/${each.value}"
  source       = "${path.module}/${var.local_images_folder}/${each.value}"
  etag         = filemd5("${path.module}/${var.local_images_folder}/${each.value}")
  content_type = "application/octet-stream"
}

resource "aws_s3_bucket_public_access_block" "movies_rest_api_bucket_public_access" {
  bucket = aws_s3_bucket.movies_rest_api_bucket.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false

}

resource "aws_s3_bucket_policy" "allow_get_images_policy" {
  bucket = aws_s3_bucket.movies_rest_api_bucket.id
  policy = data.aws_iam_policy_document.allow_get_s3_images_policy.json
}

# DynamoDB
resource "aws_dynamodb_table" "movies_db" {
  name         = "Movies"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "movieId"
  # range_key    = "releaseYear"

  attribute {
    name = "movieId"
    type = "S"
  }
  # attribute {
  #   name = "releaseYear"
  #   type = "N"
  # }

  tags = {
    "Name"        = "Movies REST API"
    "Environment" = "Dev"
  }
}

resource "aws_dynamodb_table_item" "movie_item" {
  table_name = aws_dynamodb_table.movies_db.name
  hash_key   = aws_dynamodb_table.movies_db.hash_key
  range_key  = aws_dynamodb_table.movies_db.range_key

  count = length(local.movie_data)

  item = <<ITEM
  {
  "movieId": {"S": "${local.movie_data[count.index].movieId}"},
  "title": {"S": "${local.movie_data[count.index].title}"},
  "releaseYear": {"N": "${local.movie_data[count.index].releaseYear}"},
  "genre": {"S": "${local.movie_data[count.index].genre}"},
  "coverUrl": {"S": "${local.movie_data[count.index].coverUrl}"},
  "generatedSummary": {"S": ""}
  }
  ITEM
}

# Lambda
resource "aws_lambda_function" "movies_api_lambda" {
  function_name = "movies_api_lambda"
  role          = aws_iam_role.lambda_execution_role.arn
  runtime       = "provided.al2023"
  handler       = "main"
  filename      = "${path.module}/lambda_function_payload.zip"

  timeout = 180

  source_code_hash = data.archive_file.lambda.output_base64sha256
  environment {
    variables = {
      REGION = var.aws_region
    }
  }

  tags = {
    Name        = "Movies REST API"
    Environment = "Dev"
  }
}

resource "aws_iam_role" "lambda_execution_role" {
  name               = "lambda_execution_role"
  assume_role_policy = data.aws_iam_policy_document.lambda_execution_policy.json
}

resource "aws_iam_role_policy_attachment" "lambda_cloudwatch_policy_attach" {
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_access_policy_attach" {
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.lambda_access_policy.arn
}

resource "aws_iam_policy" "lambda_access_policy" {
  name   = "lambda_dynamodb_policy"
  policy = data.aws_iam_policy_document.allow_lambda_access_policy_doc.json
}


resource "aws_lambda_permission" "apigw_lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.movies_api_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.movies_api_gateway.execution_arn}/*"
}

# API Gateway
resource "aws_api_gateway_rest_api" "movies_api_gateway" {
  name = "movies_api_gateway"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = {
    Name        = "Movies REST API"
    Environment = "Dev"
  }
}

resource "aws_api_gateway_resource" "movies_api_resource" {
  rest_api_id = aws_api_gateway_rest_api.movies_api_gateway.id
  parent_id   = aws_api_gateway_rest_api.movies_api_gateway.root_resource_id
  path_part   = "api"
}

resource "aws_api_gateway_resource" "movies_proxy_resource" {
  rest_api_id = aws_api_gateway_rest_api.movies_api_gateway.id
  parent_id   = aws_api_gateway_resource.movies_api_resource.id
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "movies_any_method" {
  rest_api_id   = aws_api_gateway_rest_api.movies_api_gateway.id
  resource_id   = aws_api_gateway_resource.movies_proxy_resource.id
  http_method   = "ANY"
  authorization = "NONE"

  request_parameters = {
    "method.request.path.proxy" = true
  }
}

resource "aws_api_gateway_integration" "movies_lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.movies_api_gateway.id
  resource_id             = aws_api_gateway_resource.movies_proxy_resource.id
  http_method             = aws_api_gateway_method.movies_any_method.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.movies_api_lambda.invoke_arn
}

resource "aws_api_gateway_deployment" "movies_api_deployment" {
  rest_api_id = aws_api_gateway_rest_api.movies_api_gateway.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.movies_api_resource.id,
      aws_api_gateway_resource.movies_proxy_resource.id,
      aws_api_gateway_method.movies_any_method.id,
      aws_api_gateway_integration.movies_lambda_integration.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "movies_api_dev_stage" {
  deployment_id = aws_api_gateway_deployment.movies_api_deployment.id
  rest_api_id   = aws_api_gateway_rest_api.movies_api_gateway.id
  stage_name    = "dev"
}
