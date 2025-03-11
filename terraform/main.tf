provider "aws" {
  region = "${var.region}"

  assume_role {
    role_arn = "arn:aws:iam::${var.account_id}:role/TerraformExecutionRole"
  }
}

terraform {
  required_version = ">= 1.10.5"
}

locals{
  project_name = "${var.project_name}"
  region = "${var.region}"
}

module "s3" {
  source = "./modules/s3"
  bucket_name = "${var.project_name}-wsf5"
}

module "ecr"{
  source = "./modules/ecr"
  project_name = "${var.project_name}"
}

module "iam" {
  source = "./modules/iam"
  project_name = "${var.project_name}"
  bucket_name = "${module.s3.s3_bucket.bucket}"
}

module "lambda" {
  source = "./modules/lambda"
  project_name = "${var.project_name}"
  openai_api_key = "${var.openai_api_key}"
  bucket_name = "${module.s3.s3_bucket.bucket}"
  ecr_repository_url = "${module.ecr.ecr_repository_url}"
  lambda_role_arn = "${module.iam.lambda_role_arn}"

  depends_on = [module.ecr, module.s3, module.iam]
}

module "api_gateway" {
  source = "./modules/api_gateway"
  account_id = "${var.account_id}"
  project_name = "${var.project_name}"
  region = "${var.region}"
  lambda_function = "${module.lambda.lambda_function}"
  lambda_role_name = "${module.iam.lambda_role_name}"
  depends_on = [module.lambda]
}