variable "aws_region" {
  type        = string
  description = "AWS region for all resources."
}

variable "project" {
  type        = string
  description = "Stable project identifier used in resource names."
}

variable "environment" {
  type        = string
  description = "Deployment environment such as staging or production."
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Additional resource tags."
}
