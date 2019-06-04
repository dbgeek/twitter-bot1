terraform {
  backend "s3" {
    bucket                 = "whitepoplar-north-1"
    key                    = "tf/eu-north-1/vpc-twitter-bot.tfstate"
    region                 = "eu-north-1"
    encrypt                = "true"
    skip_region_validation = "true"
  }
  required_version = ">= 0.12"
}

provider "aws" {
  version = "2.11.0"
}