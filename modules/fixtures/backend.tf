# this file is used to declare a backend used during the tests

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.35.0"
    }
  }

  backend "s3" {}
}
