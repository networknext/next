
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "us-east-1"
  region                   = "us-east-1"
}
