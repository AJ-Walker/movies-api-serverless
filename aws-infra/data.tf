data "aws_iam_policy_document" "allow_get_s3_images_policy" {
  statement {
    sid = "1"

    actions = ["s3:GetObject"]

    resources = ["arn:aws:s3:::${var.bucket_name}/${var.s3_images_prefix}/*"]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }
}

data "aws_iam_policy_document" "lambda_execution_policy" {
  statement {
    sid    = "1"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "allow_lambda_access_policy_doc" {
  statement {
    sid    = "1"
    effect = "Allow"

    actions   = ["dynamodb:Scan", "dynamodb:Query", "dynamodb:UpdateItem", "dynamodb:GetItem"]
    resources = [aws_dynamodb_table.movies_db.arn]
  }
  statement {
    sid    = "2"
    effect = "Allow"

    actions   = ["bedrock:InvokeModel"]
    resources = ["arn:aws:bedrock:ap-south-1::foundation-model/anthropic.claude-3-sonnet-20240229-v1:0"]
  }
}

data "archive_file" "lambda" {
  type        = "zip"
  source_file = "${path.module}/../lambda-code/bootstrap"
  output_path = "lambda_function_payload.zip"
}
