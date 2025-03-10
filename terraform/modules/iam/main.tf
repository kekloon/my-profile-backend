locals {
  lambda_policy_json = templatefile("${path.root}/policies/lambda_policy.json.tmpl", {
    bucket_name = var.bucket_name
  })
  lambda_assume_role_json = file("${path.root}/policies/lambda_assume_role.json")
}

resource "aws_iam_role" "lambda_role" {
  name = "${var.project_name}-lambda-role"

  assume_role_policy = local.lambda_assume_role_json
  
  inline_policy {
    name = "${var.project_name}-lambda-policy"
    policy = local.lambda_policy_json
  }
}

resource "aws_iam_role_policy_attachment" "lambda_policy_attachment" {
  role = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}