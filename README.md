# Movies Management System

A comprehensive movie management system built with Go, featuring a RESTful API and AWS Lambda integration for image processing. The system includes infrastructure as code using Terraform for AWS resource management.

## Project Structure

```
.
├── aws-infra/         # Terraform infrastructure code
│   ├── images/        # Movie poster images
│   └── *.tf           # Terraform configuration files
├── lambda-code/       # AWS Lambda function code
│   └── main.go        # Lambda function implementation
└── movies-api/        # Movies REST API
    ├── main.go        # API implementation
    └── movies.json    # Movie data
```

## Features

- RESTful API for movie management
- AWS Lambda integration for image processing
- Infrastructure as Code with Terraform
- Secure AWS resource management
- Movie data storage and retrieval

## Prerequisites

- Go 1.21 or later
- AWS CLI configured with appropriate credentials
- Terraform installed
- Docker (optional, for local development)

## Setup and Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd <repository-name>
```

2. Set up AWS credentials:
```bash
aws configure
```

3. Initialize and apply Terraform infrastructure:
```bash
cd aws-infra
terraform init
terraform apply
```

4. Install API dependencies:
```bash
cd ../movies-api
go mod download
```

## Running the Application

### Movies API

1. Navigate to the movies-api directory:
```bash
cd movies-api
```

2. Run the API:
```bash
go run .
```

The API will be available at `http://localhost:8080`

### Lambda Function

The Lambda function is automatically deployed through Terraform. It processes images when triggered through the configured AWS events.

## API Endpoints

- `GET /movies` - List all movies
- `GET /movies/{id}` - Get movie by ID
- Other endpoints as implemented in the API

## Infrastructure

The AWS infrastructure is managed using Terraform and includes:
- Lambda functions for image processing
- S3 buckets for storage
- IAM roles and policies
- Other AWS resources as defined in terraform files

## Development

1. Make sure to run tests before submitting changes:
```bash
go test ./...
```

2. Follow Go coding standards and conventions

3. Update infrastructure code carefully and always plan Terraform changes:
```bash
terraform plan
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

[Add your license information here]

