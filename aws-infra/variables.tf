variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-south-1"
}

variable "bucket_name" {
  description = "AWS s3 bucket name"
  type        = string
  default     = "movies-api-data"
}

variable "s3_images_prefix" {
  description = "AWS s3 images folder for the image"
  type        = string
  default     = "images"
}

variable "local_images_folder" {
  description = "Local folder name of the images"
  type        = string
  default     = "images"
}
