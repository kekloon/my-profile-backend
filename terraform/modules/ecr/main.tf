resource "aws_ecr_repository" "lambda_repository" {
  name = "${var.project_name}-lambda-repo"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name = "${var.project_name}-lambda-repo"
  }
}

resource "aws_ecr_lifecycle_policy" "lambda_repository_policy" {
  repository = aws_ecr_repository.lambda_repository.name

  policy = file("${path.root}/policies/ecr_lifecycle_policy.json")

  depends_on = [aws_ecr_repository.lambda_repository]
}