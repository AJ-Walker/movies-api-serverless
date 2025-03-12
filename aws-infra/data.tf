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

data "archive_file" "lambda" {
  type        = "zip"
  source_file = "${path.module}/../lambda-code/bootstrap"
  output_path = "lambda_function_payload.zip"
}
