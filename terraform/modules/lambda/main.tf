resource "aws_lambda_function" "lambda_function" {
  function_name = "${var.project_name}-lambda-function"
  image_uri = "${var.ecr_repository_url}:latest"
  package_type = "Image"
  role = var.lambda_role_arn
  timeout = 30
  memory_size = 1024

  environment {
    variables = {
      BUCKET_NAME = var.bucket_name
      OPENAI_API_KEY = var.openai_api_key
    }
  }
}

