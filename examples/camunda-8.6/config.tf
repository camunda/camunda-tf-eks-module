terraform {
  required_version = ">= 1.0"

  backend "s3" {
    bucket  = "my-eks-tf-state"
    key     = "camunda-terraform/terraform-std.tfstate"
    encrypt = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.69"
    }
  }
}

provider "aws" {}
