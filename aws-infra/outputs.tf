output "movies_api_dev_url" {
  description = "The dev api url of the movies api gateway"
  value       = aws_api_gateway_stage.movies_api_dev_stage.invoke_url
}
