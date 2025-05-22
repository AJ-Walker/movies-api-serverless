# Movies Management System

A simple yet powerful movie management system built with Go, featuring a RESTful API and AWS Lambda integration. The system uses Terraform for infrastructure as code (IaC) to manage AWS resources efficiently.

## Project Structure

```
.
├── aws-infra/         # Terraform infrastructure code
│   ├── images/        # Movie poster images for S3 storage
│   └── *.tf           # Terraform configuration files (e.g., main.tf, variables.tf)
├── lambda-code/       # AWS Lambda function source code
│   ├── main.go        # Core Lambda function implementation
│   ├── bedrock.go     # AWS Bedrock integration for AI-generated summaries
│   ├── dynamoDB.go    # DynamoDB operations for data storage and retrieval
│   ├── s3.go          # S3 operations for movie posters
│   └── utils.go       # Utility functions
└── movies-api/        # Movies API testing and data loading utilities
    ├── main.go        # API implementation and data insertion logic
    └── movies.json    # Sample movie data in JSON format
```

## Features

- **RESTful API**: Manage movies via intuitive endpoints.
- **Serverless Architecture**: Powered by AWS Lambda for scalability.
- **AI-Generated Summaries**: Integrated with AWS Bedrock for dynamic movie summaries.
- **Infrastructure as Code**: AWS resources provisioned and managed with Terraform.
- **Secure Resource Management**: IAM roles and policies for secure access.
- **Data Persistence**: Movie data stored in DynamoDB.
- **Dynamic Content**: Real-time generation of movie summaries.

## Prerequisites

- **Go**: Version 1.21 or later.
- **AWS CLI**: Installed and configured with valid credentials.
- **AWS Bedrock Access**: Permissions to use Bedrock for AI features.
  - We are using claude model - **anthropic.claude-3-sonnet-20240229-v1:0**
- **Terraform**: Installed for infrastructure management.

## Setup and Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd <repository-name>
```

2. Compile Go Code for AWS Lambda
   AWS Lambda requires a Linux-compatible binary. Build it with:

```bash
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap ./lambda-code
```

- This generates a bootstrap executable (not .exe unless on Windows).
- Refer to [AWS Lambda for Go](https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html) for details.

3. Set up AWS credentials:
   Ensure your AWS CLI is set up:

```bash
aws configure
```

- Provide your AWS Access Key, Secret Key, region, and output format.

4. Deploy AWS Infrastructure with Terraform:

```bash
cd aws-infra
terraform init
terraform apply
```

- `terraform init`: Initializes the Terraform working directory.
- `terraform apply`: Provisions the AWS resources (review changes before confirming).

### Lambda Function Deployment

The Lambda function is deployed automatically via Terraform. To update and redeploy the Lambda code:

1. Navigate to Lambda Code.

```bash
cd ../lambda-code
```

2. Modify Lambda Files.

Edit the relevant files as needed:

- `main.go`: Core Lambda logic.
- `bedrock.go`: Bedrock integration for summaries.
- `dynamoDB.go`: DynamoDB interactions.
- `s3.go`: S3 interactions.
- `utils.go`: Contains some utility functions.

3. Recompile for Lambda

```bash
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap .
```

This will create a **bootstrap.exe** file. For more info [AWS Lambda for Go](https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html)

4. Redeploy with Terraform:

```bash
cd ../aws-infra
terraform apply
```

- Terraform detects changes in the bootstrap file and updates the Lambda function.

### API Gateway

After Terraform applies successfully, an API Gateway URL is outputed (example only):

```
https://ty1fryoc2g.execute-api.ap-south-1.amazonaws.com/dev
```

- Use this URL as the base for API requests

## API Endpoints

- `GET /api/movies` - Retrieve a list of all movies.
- `GET /api/movies?year={year}` - Filter movies by release year.
- `GET /api/movies?movieId={movieId}` - Get a specific movie by ID.
- `POST /api/movies` - Add a new movie (accepts multipart form data with title, releaseYear, genre, and optional coverImage).
- `PUT /api/movies?movieId={movieId}` - Update a movie's details and/or poster image (accepts multipart form data with title, releaseYear, genre, and optional coverImage).
- `DELETE /api/movies?movieId={movieId}` - Delete a movie and its associated poster from S3.
- `GET /api/movies/summary?movieId={movieId}` - Fetch an AI-generated summary for a specific movie.

## API Testing with Postman

A Postman collection has been included in this repository to help you test and interact with the API. This collection contains pre-configured requests for all available endpoints, complete with test scripts to validate responses.

### Setting Up the Postman Collection

1. **Import the Collection**:
   - Open Postman
   - Click on "Import" button in the top left
   - Select the `Movies Serverless API.postman_collection.json` file from the project root
   - The collection should appear in your Postman workspace

2. **Configure Environment Variable**:
   - Create a new environment in Postman (click on "Environments" tab)
   - Add a variable named `API_URL` with the value of your API Gateway URL (without protocol and without trailing slash)
   - Example: If your URL is `https://ty1fryoc2g.execute-api.ap-south-1.amazonaws.com/dev`, set `API_URL` to `ty1fryoc2g.execute-api.ap-south-1.amazonaws.com`
   - Save the environment and make sure to select it when using the collection

### Using the Collection

The collection contains the following requests:

1. **Get All Movies**: Retrieves the complete list of movies
2. **Get Movies By Year**: Filters movies by a specific release year
3. **Get Movie By MovieId**: Retrieves a specific movie by its ID
4. **Get Movie Summary**: Fetches the AI-generated summary for a movie
5. **Add Movie**: Creates a new movie entry with optional cover image
6. **Update Movie By MovieId**: Updates an existing movie's details
7. **Delete Movie By MovieId**: Removes a movie from the database

Each request includes:
- Appropriate HTTP method
- Required path and query parameters
- Test scripts to validate responses
- Description of the expected request/response format

For requests that require a movie ID (such as Get Movie By MovieId, Delete Movie, etc.), you'll need to:
1. First run the "Get All Movies" request
2. Copy a movie ID from the response
3. Paste it into the appropriate parameter for the subsequent request

For the Add Movie and Update Movie requests that accept file uploads, you can select any image file from your local system for testing.

## DynamoDB Schema

The movie data is stored in DynamoDB with the following structure:

- `movieId` (Primary Key): Unique identifier for each movie
- Note: Previously, `releaseYear` was used as a sort key, but it has been removed to simplify the schema and allow for more flexible querying.

## Movie Summary Feature

The system leverages AWS Bedrock to generate detailed movie summaries:

1. Retrieves movie details from DynamoDB.
2. Constructs and sends a prompt to AWS Bedrock.
3. Processes the AI response.
4. Stores the summary in DynamoDB for future use.
5. Returns the summary via the `/summary` endpoint.

## Infrastructure

Managed via Terraform, the AWS setup includes:

- **AWS Lambda**: Executes the serverless logic.
- **API Gateway**: 
  - Exposes the RESTful API.
  - Uses proxy integration for flexible routing and request handling.
  - Configured to route all requests to the Lambda function for centralized processing.
  - Handles query parameters through centralized routing for flexible request processing.
- **S3 Buckets**: Stores movie poster images with automated deletion when movies are removed.
- **IAM Roles/Policies**: Ensures secure resource access.
- **DynamoDB**: Persists movie data and summaries.

## Development Tips

1. Adhere to (Go coding standards)[https://go.dev/doc/effective_go].
2. Update infrastructure code carefully and always plan Terraform changes:

```bash
terraform plan
```

This helps avoid unintended infrastructure modifications.

## Future Changes
1. ...

## Troubleshooting

- **AWS S3 Bucket Policy Issue During** `terraform apply`:
  Sometimes, when running `terraform apply`, you may encounter an error related to S3 bucket policies due to state mismatches or permission conflicts. To resolve this:

1. Run a Terraform refresh to sync the state with the actual AWS resources:

```bash
terraform refresh
```

2.  Apply the changes again:

```bash
terraform apply
```

This ensures Terraform has the latest state and can resolve policy-related issues.

## Contributing

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/<name>`).
3. Commit your changes (`git commit -m "Add feature"`).
4. Push to the branch (`git push origin feature/<name>`).
5. Open a Pull Request.
