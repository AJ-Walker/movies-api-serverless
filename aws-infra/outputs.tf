output "movies_api_dev_url" {
  description = "The dev api url of the movies api gateway"
  value       = aws_api_gateway_stage.movies_api_dev_stage.invoke_url
}

output "dynamodb_table_arn" {
  description = "The ARN of the dynamodb table"
  value       = aws_dynamodb_table.movies_db.arn
}
