terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.35.0"
    }
    rhcs = {
      version = "= 1.6.2"
      source  = "terraform-redhat/rhcs"
    }
  }

  backend "s3" {}
}

provider "rhcs" {
  token = var.offline_access_token
  url   = var.url
}

data "aws_caller_identity" "current" {}
