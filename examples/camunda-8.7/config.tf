terraform {
  required_version = ">= 1.0"

  # You can override the backend configuration; this is  given as an example.
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
