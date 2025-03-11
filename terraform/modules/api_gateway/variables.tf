variable "project_name" {
  type = string
}

variable "account_id" {
  type = string
}

variable "region" {
  type = string
}

variable "lambda_function" {
  type = object({
    function_name = string
    invoke_arn = string
    image_uri = string
  })
}

variable "lambda_role_name" {
  type = string
}
