
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-south-2"
  region                   = "ap-south-2"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-south-1"
  region                   = "ap-south-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "eu-south-1"
  region                   = "eu-south-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "eu-south-2"
  region                   = "eu-south-2"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "me-central-1"
  region                   = "me-central-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "il-central-1"
  region                   = "il-central-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ca-central-1"
  region                   = "ca-central-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-east-2"
  region                   = "ap-east-2"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "mx-central-1"
  region                   = "mx-central-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "eu-central-1"
  region                   = "eu-central-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "eu-central-2"
  region                   = "eu-central-2"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "us-west-1"
  region                   = "us-west-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "us-west-2"
  region                   = "us-west-2"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "af-south-1"
  region                   = "af-south-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "eu-north-1"
  region                   = "eu-north-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "eu-west-3"
  region                   = "eu-west-3"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "eu-west-2"
  region                   = "eu-west-2"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "eu-west-1"
  region                   = "eu-west-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-northeast-3"
  region                   = "ap-northeast-3"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-northeast-2"
  region                   = "ap-northeast-2"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "me-south-1"
  region                   = "me-south-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-northeast-1"
  region                   = "ap-northeast-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "sa-east-1"
  region                   = "sa-east-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-east-1"
  region                   = "ap-east-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ca-west-1"
  region                   = "ca-west-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-southeast-1"
  region                   = "ap-southeast-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-southeast-2"
  region                   = "ap-southeast-2"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-southeast-3"
  region                   = "ap-southeast-3"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-southeast-4"
  region                   = "ap-southeast-4"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "us-east-1"
  region                   = "us-east-1"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-southeast-5"
  region                   = "ap-southeast-5"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "us-east-2"
  region                   = "us-east-2"
}

provider "aws" { 
  shared_config_files      = var.config
  shared_credentials_files = var.credentials
  profile                  = var.profile
  alias                    = "ap-southeast-7"
  region                   = "ap-southeast-7"
}

module "region_ap_south_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-south-2
  }
}

module "region_ap_south_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-south-1
  }
}

module "region_eu_south_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.eu-south-1
  }
}

module "region_eu_south_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.eu-south-2
  }
}

module "region_me_central_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.me-central-1
  }
}

module "region_il_central_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.il-central-1
  }
}

module "region_ca_central_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ca-central-1
  }
}

module "region_ap_east_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-east-2
  }
}

module "region_mx_central_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.mx-central-1
  }
}

module "region_eu_central_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.eu-central-1
  }
}

module "region_eu_central_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.eu-central-2
  }
}

module "region_us_west_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.us-west-1
  }
}

module "region_us_west_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.us-west-2
  }
}

module "region_af_south_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.af-south-1
  }
}

module "region_eu_north_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.eu-north-1
  }
}

module "region_eu_west_3" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.eu-west-3
  }
}

module "region_eu_west_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.eu-west-2
  }
}

module "region_eu_west_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.eu-west-1
  }
}

module "region_ap_northeast_3" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-northeast-3
  }
}

module "region_ap_northeast_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-northeast-2
  }
}

module "region_me_south_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.me-south-1
  }
}

module "region_ap_northeast_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-northeast-1
  }
}

module "region_sa_east_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.sa-east-1
  }
}

module "region_ap_east_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-east-1
  }
}

module "region_ca_west_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ca-west-1
  }
}

module "region_ap_southeast_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-southeast-1
  }
}

module "region_ap_southeast_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-southeast-2
  }
}

module "region_ap_southeast_3" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-southeast-3
  }
}

module "region_ap_southeast_4" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-southeast-4
  }
}

module "region_us_east_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.us-east-1
  }
}

module "region_ap_southeast_5" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-southeast-5
  }
}

module "region_us_east_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.us-east-2
  }
}

module "region_ap_southeast_7" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ap-southeast-7
  }
}

locals {

  datacenter_map = {

    "amazon.johannesburg.1" = {
      azid        = "afs1-az1"
      zone        = "af-south-1a"
      region      = "af-south-1"
      native_name = "afs1-az1 (af-south-1a)"
      latitude    = -33.92
      longitude   = 18.42
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.johannesburg.2" = {
      azid        = "afs1-az2"
      zone        = "af-south-1b"
      region      = "af-south-1"
      native_name = "afs1-az2 (af-south-1b)"
      latitude    = -33.92
      longitude   = 18.42
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.johannesburg.3" = {
      azid        = "afs1-az3"
      zone        = "af-south-1c"
      region      = "af-south-1"
      native_name = "afs1-az3 (af-south-1c)"
      latitude    = -33.92
      longitude   = 18.42
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.nigeria.1" = {
      azid        = "afs1-los1-az1"
      zone        = "af-south-1-los-1a"
      region      = "af-south-1"
      native_name = "afs1-los1-az1 (af-south-1-los-1a)"
      latitude    = 6.52
      longitude   = 3.38
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.hongkong.1" = {
      azid        = "ape1-az1"
      zone        = "ap-east-1a"
      region      = "ap-east-1"
      native_name = "ape1-az1 (ap-east-1a)"
      latitude    = 22.32
      longitude   = 114.17
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.hongkong.2" = {
      azid        = "ape1-az2"
      zone        = "ap-east-1b"
      region      = "ap-east-1"
      native_name = "ape1-az2 (ap-east-1b)"
      latitude    = 22.32
      longitude   = 114.17
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.hongkong.3" = {
      azid        = "ape1-az3"
      zone        = "ap-east-1c"
      region      = "ap-east-1"
      native_name = "ape1-az3 (ap-east-1c)"
      latitude    = 22.32
      longitude   = 114.17
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.tokyo.1" = {
      azid        = "apne1-az1"
      zone        = "ap-northeast-1c"
      region      = "ap-northeast-1"
      native_name = "apne1-az1 (ap-northeast-1c)"
      latitude    = 35.68
      longitude   = 139.65
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.tokyo.2" = {
      azid        = "apne1-az2"
      zone        = "ap-northeast-1d"
      region      = "ap-northeast-1"
      native_name = "apne1-az2 (ap-northeast-1d)"
      latitude    = 35.68
      longitude   = 139.65
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.tokyo.4" = {
      azid        = "apne1-az4"
      zone        = "ap-northeast-1a"
      region      = "ap-northeast-1"
      native_name = "apne1-az4 (ap-northeast-1a)"
      latitude    = 35.68
      longitude   = 139.65
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.taipei.1" = {
      azid        = "apne1-tpe1-az1"
      zone        = "ap-northeast-1-tpe-1a"
      region      = "ap-northeast-1"
      native_name = "apne1-tpe1-az1 (ap-northeast-1-tpe-1a)"
      latitude    = 25.03
      longitude   = 121.57
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.seoul.1" = {
      azid        = "apne2-az1"
      zone        = "ap-northeast-2a"
      region      = "ap-northeast-2"
      native_name = "apne2-az1 (ap-northeast-2a)"
      latitude    = 37.57
      longitude   = 126.98
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.seoul.2" = {
      azid        = "apne2-az2"
      zone        = "ap-northeast-2b"
      region      = "ap-northeast-2"
      native_name = "apne2-az2 (ap-northeast-2b)"
      latitude    = 37.57
      longitude   = 126.98
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.seoul.3" = {
      azid        = "apne2-az3"
      zone        = "ap-northeast-2c"
      region      = "ap-northeast-2"
      native_name = "apne2-az3 (ap-northeast-2c)"
      latitude    = 37.57
      longitude   = 126.98
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.seoul.4" = {
      azid        = "apne2-az4"
      zone        = "ap-northeast-2d"
      region      = "ap-northeast-2"
      native_name = "apne2-az4 (ap-northeast-2d)"
      latitude    = 37.57
      longitude   = 126.98
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.osaka.1" = {
      azid        = "apne3-az1"
      zone        = "ap-northeast-3b"
      region      = "ap-northeast-3"
      native_name = "apne3-az1 (ap-northeast-3b)"
      latitude    = 34.69
      longitude   = 135.50
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.osaka.2" = {
      azid        = "apne3-az2"
      zone        = "ap-northeast-3c"
      region      = "ap-northeast-3"
      native_name = "apne3-az2 (ap-northeast-3c)"
      latitude    = 34.69
      longitude   = 135.50
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.osaka.3" = {
      azid        = "apne3-az3"
      zone        = "ap-northeast-3a"
      region      = "ap-northeast-3"
      native_name = "apne3-az3 (ap-northeast-3a)"
      latitude    = 34.69
      longitude   = 135.50
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.mumbai.1" = {
      azid        = "aps1-az1"
      zone        = "ap-south-1a"
      region      = "ap-south-1"
      native_name = "aps1-az1 (ap-south-1a)"
      latitude    = 19.08
      longitude   = 72.88
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.mumbai.2" = {
      azid        = "aps1-az2"
      zone        = "ap-south-1c"
      region      = "ap-south-1"
      native_name = "aps1-az2 (ap-south-1c)"
      latitude    = 19.08
      longitude   = 72.88
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.mumbai.3" = {
      azid        = "aps1-az3"
      zone        = "ap-south-1b"
      region      = "ap-south-1"
      native_name = "aps1-az3 (ap-south-1b)"
      latitude    = 19.08
      longitude   = 72.88
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.kolkata.1" = {
      azid        = "aps1-ccu1-az1"
      zone        = "ap-south-1-ccu-1a"
      region      = "ap-south-1"
      native_name = "aps1-ccu1-az1 (ap-south-1-ccu-1a)"
      latitude    = 22.57
      longitude   = 88.36
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.delhi.1" = {
      azid        = "aps1-del1-az1"
      zone        = "ap-south-1-del-1a"
      region      = "ap-south-1"
      native_name = "aps1-del1-az1 (ap-south-1-del-1a)"
      latitude    = 28.70
      longitude   = 77.10
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.hyderabad.1" = {
      azid        = "aps2-az1"
      zone        = "ap-south-2a"
      region      = "ap-south-2"
      native_name = "aps2-az1 (ap-south-2a)"
      latitude    = 17.39
      longitude   = 78.49
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.hyderabad.2" = {
      azid        = "aps2-az2"
      zone        = "ap-south-2b"
      region      = "ap-south-2"
      native_name = "aps2-az2 (ap-south-2b)"
      latitude    = 17.39
      longitude   = 78.49
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.hyderabad.3" = {
      azid        = "aps2-az3"
      zone        = "ap-south-2c"
      region      = "ap-south-2"
      native_name = "aps2-az3 (ap-south-2c)"
      latitude    = 17.39
      longitude   = 78.49
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.singapore.1" = {
      azid        = "apse1-az1"
      zone        = "ap-southeast-1a"
      region      = "ap-southeast-1"
      native_name = "apse1-az1 (ap-southeast-1a)"
      latitude    = 1.35
      longitude   = 103.82
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.singapore.2" = {
      azid        = "apse1-az2"
      zone        = "ap-southeast-1b"
      region      = "ap-southeast-1"
      native_name = "apse1-az2 (ap-southeast-1b)"
      latitude    = 1.35
      longitude   = 103.82
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.singapore.3" = {
      azid        = "apse1-az3"
      zone        = "ap-southeast-1c"
      region      = "ap-southeast-1"
      native_name = "apse1-az3 (ap-southeast-1c)"
      latitude    = 1.35
      longitude   = 103.82
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.sydney.1" = {
      azid        = "apse2-az1"
      zone        = "ap-southeast-2a"
      region      = "ap-southeast-2"
      native_name = "apse2-az1 (ap-southeast-2a)"
      latitude    = -33.87
      longitude   = 151.21
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.sydney.2" = {
      azid        = "apse2-az2"
      zone        = "ap-southeast-2c"
      region      = "ap-southeast-2"
      native_name = "apse2-az2 (ap-southeast-2c)"
      latitude    = -33.87
      longitude   = 151.21
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.sydney.3" = {
      azid        = "apse2-az3"
      zone        = "ap-southeast-2b"
      region      = "ap-southeast-2"
      native_name = "apse2-az3 (ap-southeast-2b)"
      latitude    = -33.87
      longitude   = 151.21
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.perth.1" = {
      azid        = "apse2-per1-az1"
      zone        = "ap-southeast-2-per-1a"
      region      = "ap-southeast-2"
      native_name = "apse2-per1-az1 (ap-southeast-2-per-1a)"
      latitude    = -31.95
      longitude   = 115.86
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.jakarta.1" = {
      azid        = "apse3-az1"
      zone        = "ap-southeast-3a"
      region      = "ap-southeast-3"
      native_name = "apse3-az1 (ap-southeast-3a)"
      latitude    = -6.21
      longitude   = 106.85
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.jakarta.2" = {
      azid        = "apse3-az2"
      zone        = "ap-southeast-3b"
      region      = "ap-southeast-3"
      native_name = "apse3-az2 (ap-southeast-3b)"
      latitude    = -6.21
      longitude   = 106.85
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.jakarta.3" = {
      azid        = "apse3-az3"
      zone        = "ap-southeast-3c"
      region      = "ap-southeast-3"
      native_name = "apse3-az3 (ap-southeast-3c)"
      latitude    = -6.21
      longitude   = 106.85
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.melbourne.1" = {
      azid        = "apse4-az1"
      zone        = "ap-southeast-4a"
      region      = "ap-southeast-4"
      native_name = "apse4-az1 (ap-southeast-4a)"
      latitude    = -37.81
      longitude   = 144.96
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.melbourne.2" = {
      azid        = "apse4-az2"
      zone        = "ap-southeast-4b"
      region      = "ap-southeast-4"
      native_name = "apse4-az2 (ap-southeast-4b)"
      latitude    = -37.81
      longitude   = 144.96
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.melbourne.3" = {
      azid        = "apse4-az3"
      zone        = "ap-southeast-4c"
      region      = "ap-southeast-4"
      native_name = "apse4-az3 (ap-southeast-4c)"
      latitude    = -37.81
      longitude   = 144.96
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.malaysia.1" = {
      azid        = "apse5-az1"
      zone        = "ap-southeast-5a"
      region      = "ap-southeast-5"
      native_name = "apse5-az1 (ap-southeast-5a)"
      latitude    = 4.21
      longitude   = 101.98
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.malaysia.2" = {
      azid        = "apse5-az2"
      zone        = "ap-southeast-5b"
      region      = "ap-southeast-5"
      native_name = "apse5-az2 (ap-southeast-5b)"
      latitude    = 4.21
      longitude   = 101.98
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.malaysia.3" = {
      azid        = "apse5-az3"
      zone        = "ap-southeast-5c"
      region      = "ap-southeast-5"
      native_name = "apse5-az3 (ap-southeast-5c)"
      latitude    = 4.21
      longitude   = 101.98
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.newzealand.1" = {
      azid        = "apse6-az1"
      zone        = "ap-southeast-6a"
      region      = "ap-southeast-6"
      native_name = "apse6-az1 (ap-southeast-6a)"
      latitude    = -36.85
      longitude   = 174.76
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.newzealand.2" = {
      azid        = "apse6-az2"
      zone        = "ap-southeast-6b"
      region      = "ap-southeast-6"
      native_name = "apse6-az2 (ap-southeast-6b)"
      latitude    = -36.85
      longitude   = 174.76
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.newzealand.3" = {
      azid        = "apse6-az3"
      zone        = "ap-southeast-6c"
      region      = "ap-southeast-6"
      native_name = "apse6-az3 (ap-southeast-6c)"
      latitude    = -36.85
      longitude   = 174.76
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.thailand.1" = {
      azid        = "apse7-az1"
      zone        = "ap-southeast-7a"
      region      = "ap-southeast-7"
      native_name = "apse7-az1 (ap-southeast-7a)"
      latitude    = 15.87
      longitude   = 100.99
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.thailand.2" = {
      azid        = "apse7-az2"
      zone        = "ap-southeast-7b"
      region      = "ap-southeast-7"
      native_name = "apse7-az2 (ap-southeast-7b)"
      latitude    = 15.87
      longitude   = 100.99
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.thailand.3" = {
      azid        = "apse7-az3"
      zone        = "ap-southeast-7c"
      region      = "ap-southeast-7"
      native_name = "apse7-az3 (ap-southeast-7c)"
      latitude    = 15.87
      longitude   = 100.99
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.montreal.1" = {
      azid        = "cac1-az1"
      zone        = "ca-central-1a"
      region      = "ca-central-1"
      native_name = "cac1-az1 (ca-central-1a)"
      latitude    = 45.50
      longitude   = -73.57
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.montreal.2" = {
      azid        = "cac1-az2"
      zone        = "ca-central-1b"
      region      = "ca-central-1"
      native_name = "cac1-az2 (ca-central-1b)"
      latitude    = 45.50
      longitude   = -73.57
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.montreal.4" = {
      azid        = "cac1-az4"
      zone        = "ca-central-1d"
      region      = "ca-central-1"
      native_name = "cac1-az4 (ca-central-1d)"
      latitude    = 45.50
      longitude   = -73.57
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.calgary.1" = {
      azid        = "caw1-az1"
      zone        = "ca-west-1a"
      region      = "ca-west-1"
      native_name = "caw1-az1 (ca-west-1a)"
      latitude    = 51.04
      longitude   = -114.07
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.calgary.2" = {
      azid        = "caw1-az2"
      zone        = "ca-west-1b"
      region      = "ca-west-1"
      native_name = "caw1-az2 (ca-west-1b)"
      latitude    = 51.04
      longitude   = -114.07
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.calgary.3" = {
      azid        = "caw1-az3"
      zone        = "ca-west-1c"
      region      = "ca-west-1"
      native_name = "caw1-az3 (ca-west-1c)"
      latitude    = 51.04
      longitude   = -114.07
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.frankfurt.1" = {
      azid        = "euc1-az1"
      zone        = "eu-central-1c"
      region      = "eu-central-1"
      native_name = "euc1-az1 (eu-central-1c)"
      latitude    = 50.11
      longitude   = 8.68
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.frankfurt.2" = {
      azid        = "euc1-az2"
      zone        = "eu-central-1a"
      region      = "eu-central-1"
      native_name = "euc1-az2 (eu-central-1a)"
      latitude    = 50.11
      longitude   = 8.68
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.frankfurt.3" = {
      azid        = "euc1-az3"
      zone        = "eu-central-1b"
      region      = "eu-central-1"
      native_name = "euc1-az3 (eu-central-1b)"
      latitude    = 50.11
      longitude   = 8.68
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.hamburg.1" = {
      azid        = "euc1-ham1-az1"
      zone        = "eu-central-1-ham-1a"
      region      = "eu-central-1"
      native_name = "euc1-ham1-az1 (eu-central-1-ham-1a)"
      latitude    = 53.55
      longitude   = 9.99
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.warsaw.1" = {
      azid        = "euc1-waw1-az1"
      zone        = "eu-central-1-waw-1a"
      region      = "eu-central-1"
      native_name = "euc1-waw1-az1 (eu-central-1-waw-1a)"
      latitude    = 52.23
      longitude   = 21.01
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.zurich.1" = {
      azid        = "euc2-az1"
      zone        = "eu-central-2a"
      region      = "eu-central-2"
      native_name = "euc2-az1 (eu-central-2a)"
      latitude    = 47.38
      longitude   = 8.54
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.zurich.2" = {
      azid        = "euc2-az2"
      zone        = "eu-central-2b"
      region      = "eu-central-2"
      native_name = "euc2-az2 (eu-central-2b)"
      latitude    = 47.38
      longitude   = 8.54
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.zurich.3" = {
      azid        = "euc2-az3"
      zone        = "eu-central-2c"
      region      = "eu-central-2"
      native_name = "euc2-az3 (eu-central-2c)"
      latitude    = 47.38
      longitude   = 8.54
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.stockholm.1" = {
      azid        = "eun1-az1"
      zone        = "eu-north-1a"
      region      = "eu-north-1"
      native_name = "eun1-az1 (eu-north-1a)"
      latitude    = 59.33
      longitude   = 18.07
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.stockholm.2" = {
      azid        = "eun1-az2"
      zone        = "eu-north-1b"
      region      = "eu-north-1"
      native_name = "eun1-az2 (eu-north-1b)"
      latitude    = 59.33
      longitude   = 18.07
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.stockholm.3" = {
      azid        = "eun1-az3"
      zone        = "eu-north-1c"
      region      = "eu-north-1"
      native_name = "eun1-az3 (eu-north-1c)"
      latitude    = 59.33
      longitude   = 18.07
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.copenhagen.1" = {
      azid        = "eun1-cph1-az1"
      zone        = "eu-north-1-cph-1a"
      region      = "eu-north-1"
      native_name = "eun1-cph1-az1 (eu-north-1-cph-1a)"
      latitude    = 55.68
      longitude   = 12.57
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.finland.1" = {
      azid        = "eun1-hel1-az1"
      zone        = "eu-north-1-hel-1a"
      region      = "eu-north-1"
      native_name = "eun1-hel1-az1 (eu-north-1-hel-1a)"
      latitude    = 60.17
      longitude   = 24.94
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.milan.1" = {
      azid        = "eus1-az1"
      zone        = "eu-south-1a"
      region      = "eu-south-1"
      native_name = "eus1-az1 (eu-south-1a)"
      latitude    = 45.46
      longitude   = 9.19
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.milan.2" = {
      azid        = "eus1-az2"
      zone        = "eu-south-1b"
      region      = "eu-south-1"
      native_name = "eus1-az2 (eu-south-1b)"
      latitude    = 45.46
      longitude   = 9.19
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.milan.3" = {
      azid        = "eus1-az3"
      zone        = "eu-south-1c"
      region      = "eu-south-1"
      native_name = "eus1-az3 (eu-south-1c)"
      latitude    = 45.46
      longitude   = 9.19
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.spain.1" = {
      azid        = "eus2-az1"
      zone        = "eu-south-2a"
      region      = "eu-south-2"
      native_name = "eus2-az1 (eu-south-2a)"
      latitude    = 41.60
      longitude   = -0.91
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.spain.2" = {
      azid        = "eus2-az2"
      zone        = "eu-south-2b"
      region      = "eu-south-2"
      native_name = "eus2-az2 (eu-south-2b)"
      latitude    = 41.60
      longitude   = -0.91
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.spain.3" = {
      azid        = "eus2-az3"
      zone        = "eu-south-2c"
      region      = "eu-south-2"
      native_name = "eus2-az3 (eu-south-2c)"
      latitude    = 41.60
      longitude   = -0.91
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.ireland.1" = {
      azid        = "euw1-az1"
      zone        = "eu-west-1c"
      region      = "eu-west-1"
      native_name = "euw1-az1 (eu-west-1c)"
      latitude    = 53.78
      longitude   = -7.31
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.ireland.2" = {
      azid        = "euw1-az2"
      zone        = "eu-west-1a"
      region      = "eu-west-1"
      native_name = "euw1-az2 (eu-west-1a)"
      latitude    = 53.78
      longitude   = -7.31
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.ireland.3" = {
      azid        = "euw1-az3"
      zone        = "eu-west-1b"
      region      = "eu-west-1"
      native_name = "euw1-az3 (eu-west-1b)"
      latitude    = 53.78
      longitude   = -7.31
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.london.1" = {
      azid        = "euw2-az1"
      zone        = "eu-west-2c"
      region      = "eu-west-2"
      native_name = "euw2-az1 (eu-west-2c)"
      latitude    = 51.51
      longitude   = -0.13
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.london.2" = {
      azid        = "euw2-az2"
      zone        = "eu-west-2a"
      region      = "eu-west-2"
      native_name = "euw2-az2 (eu-west-2a)"
      latitude    = 51.51
      longitude   = -0.13
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.london.3" = {
      azid        = "euw2-az3"
      zone        = "eu-west-2b"
      region      = "eu-west-2"
      native_name = "euw2-az3 (eu-west-2b)"
      latitude    = 51.51
      longitude   = -0.13
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.paris.1" = {
      azid        = "euw3-az1"
      zone        = "eu-west-3a"
      region      = "eu-west-3"
      native_name = "euw3-az1 (eu-west-3a)"
      latitude    = 48.86
      longitude   = 2.35
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.paris.2" = {
      azid        = "euw3-az2"
      zone        = "eu-west-3b"
      region      = "eu-west-3"
      native_name = "euw3-az2 (eu-west-3b)"
      latitude    = 48.86
      longitude   = 2.35
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.paris.3" = {
      azid        = "euw3-az3"
      zone        = "eu-west-3c"
      region      = "eu-west-3"
      native_name = "euw3-az3 (eu-west-3c)"
      latitude    = 48.86
      longitude   = 2.35
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.telaviv.1" = {
      azid        = "ilc1-az1"
      zone        = "il-central-1a"
      region      = "il-central-1"
      native_name = "ilc1-az1 (il-central-1a)"
      latitude    = 32.09
      longitude   = 34.78
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.telaviv.2" = {
      azid        = "ilc1-az2"
      zone        = "il-central-1b"
      region      = "il-central-1"
      native_name = "ilc1-az2 (il-central-1b)"
      latitude    = 32.09
      longitude   = 34.78
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.telaviv.3" = {
      azid        = "ilc1-az3"
      zone        = "il-central-1c"
      region      = "il-central-1"
      native_name = "ilc1-az3 (il-central-1c)"
      latitude    = 32.09
      longitude   = 34.78
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.uae.1" = {
      azid        = "mec1-az1"
      zone        = "me-central-1a"
      region      = "me-central-1"
      native_name = "mec1-az1 (me-central-1a)"
      latitude    = 23.42
      longitude   = 53.85
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.uae.2" = {
      azid        = "mec1-az2"
      zone        = "me-central-1b"
      region      = "me-central-1"
      native_name = "mec1-az2 (me-central-1b)"
      latitude    = 23.42
      longitude   = 53.85
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.uae.3" = {
      azid        = "mec1-az3"
      zone        = "me-central-1c"
      region      = "me-central-1"
      native_name = "mec1-az3 (me-central-1c)"
      latitude    = 23.42
      longitude   = 53.85
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.bahrain.1" = {
      azid        = "mes1-az1"
      zone        = "me-south-1a"
      region      = "me-south-1"
      native_name = "mes1-az1 (me-south-1a)"
      latitude    = 26.07
      longitude   = 50.56
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.bahrain.2" = {
      azid        = "mes1-az2"
      zone        = "me-south-1b"
      region      = "me-south-1"
      native_name = "mes1-az2 (me-south-1b)"
      latitude    = 26.07
      longitude   = 50.56
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.bahrain.3" = {
      azid        = "mes1-az3"
      zone        = "me-south-1c"
      region      = "me-south-1"
      native_name = "mes1-az3 (me-south-1c)"
      latitude    = 26.07
      longitude   = 50.56
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.oman.1" = {
      azid        = "mes1-mct1-az1"
      zone        = "me-south-1-mct-1a"
      region      = "me-south-1"
      native_name = "mes1-mct1-az1 (me-south-1-mct-1a)"
      latitude    = 23.59
      longitude   = 58.38
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.mexico.1" = {
      azid        = "mxc1-az1"
      zone        = "mx-central-1a"
      region      = "mx-central-1"
      native_name = "mxc1-az1 (mx-central-1a)"
      latitude    = 23.63
      longitude   = -102.55
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.mexico.2" = {
      azid        = "mxc1-az2"
      zone        = "mx-central-1b"
      region      = "mx-central-1"
      native_name = "mxc1-az2 (mx-central-1b)"
      latitude    = 23.63
      longitude   = -102.55
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.mexico.3" = {
      azid        = "mxc1-az3"
      zone        = "mx-central-1c"
      region      = "mx-central-1"
      native_name = "mxc1-az3 (mx-central-1c)"
      latitude    = 23.63
      longitude   = -102.55
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.saopaulo.1" = {
      azid        = "sae1-az1"
      zone        = "sa-east-1a"
      region      = "sa-east-1"
      native_name = "sae1-az1 (sa-east-1a)"
      latitude    = -23.56
      longitude   = -46.64
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.saopaulo.2" = {
      azid        = "sae1-az2"
      zone        = "sa-east-1b"
      region      = "sa-east-1"
      native_name = "sae1-az2 (sa-east-1b)"
      latitude    = -23.56
      longitude   = -46.64
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.saopaulo.3" = {
      azid        = "sae1-az3"
      zone        = "sa-east-1c"
      region      = "sa-east-1"
      native_name = "sae1-az3 (sa-east-1c)"
      latitude    = -23.56
      longitude   = -46.64
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.virginia.1" = {
      azid        = "use1-az1"
      zone        = "us-east-1c"
      region      = "us-east-1"
      native_name = "use1-az1 (us-east-1c)"
      latitude    = 39.04
      longitude   = -77.49
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.virginia.2" = {
      azid        = "use1-az2"
      zone        = "us-east-1d"
      region      = "us-east-1"
      native_name = "use1-az2 (us-east-1d)"
      latitude    = 39.04
      longitude   = -77.49
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.virginia.3" = {
      azid        = "use1-az3"
      zone        = "us-east-1e"
      region      = "us-east-1"
      native_name = "use1-az3 (us-east-1e)"
      latitude    = 39.04
      longitude   = -77.49
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.virginia.4" = {
      azid        = "use1-az4"
      zone        = "us-east-1a"
      region      = "us-east-1"
      native_name = "use1-az4 (us-east-1a)"
      latitude    = 39.04
      longitude   = -77.49
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.virginia.5" = {
      azid        = "use1-az5"
      zone        = "us-east-1f"
      region      = "us-east-1"
      native_name = "use1-az5 (us-east-1f)"
      latitude    = 39.04
      longitude   = -77.49
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.virginia.6" = {
      azid        = "use1-az6"
      zone        = "us-east-1b"
      region      = "us-east-1"
      native_name = "use1-az6 (us-east-1b)"
      latitude    = 39.04
      longitude   = -77.49
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.boston.1" = {
      azid        = "use1-bos1-az1"
      zone        = "us-east-1-bos-1a"
      region      = "us-east-1"
      native_name = "use1-bos1-az1 (us-east-1-bos-1a)"
      latitude    = 42.36
      longitude   = -71.06
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.buenosaires.1" = {
      azid        = "use1-bue1-az1"
      zone        = "us-east-1-bue-1a"
      region      = "us-east-1"
      native_name = "use1-bue1-az1 (us-east-1-bue-1a)"
      latitude    = -34.60
      longitude   = -58.38
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.chicago.1" = {
      azid        = "use1-chi2-az1"
      zone        = "us-east-1-chi-2a"
      region      = "us-east-1"
      native_name = "use1-chi2-az1 (us-east-1-chi-2a)"
      latitude    = 41.88
      longitude   = -87.63
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.dallas.1" = {
      azid        = "use1-dfw2-az1"
      zone        = "us-east-1-dfw-2a"
      region      = "us-east-1"
      native_name = "use1-dfw2-az1 (us-east-1-dfw-2a)"
      latitude    = 32.78
      longitude   = -96.80
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.lima.1" = {
      azid        = "use1-lim1-az1"
      zone        = "us-east-1-lim-1a"
      region      = "us-east-1"
      native_name = "use1-lim1-az1 (us-east-1-lim-1a)"
      latitude    = -12.05
      longitude   = -77.04
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.kansas.1" = {
      azid        = "use1-mci1-az1"
      zone        = "us-east-1-mci-1a"
      region      = "us-east-1"
      native_name = "use1-mci1-az1 (us-east-1-mci-1a)"
      latitude    = 39.10
      longitude   = -94.58
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.miami.1" = {
      azid        = "use1-mia2-az1"
      zone        = "us-east-1-mia-2a"
      region      = "us-east-1"
      native_name = "use1-mia2-az1 (us-east-1-mia-2a)"
      latitude    = 25.76
      longitude   = -80.19
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.minneapolis.1" = {
      azid        = "use1-msp1-az1"
      zone        = "us-east-1-msp-1a"
      region      = "us-east-1"
      native_name = "use1-msp1-az1 (us-east-1-msp-1a)"
      latitude    = 44.98
      longitude   = -93.26
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.newyork.1" = {
      azid        = "use1-nyc2-az1"
      zone        = "us-east-1-nyc-2a"
      region      = "us-east-1"
      native_name = "use1-nyc2-az1 (us-east-1-nyc-2a)"
      latitude    = 40.71
      longitude   = -74.01
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.philadelphia.1" = {
      azid        = "use1-phl1-az1"
      zone        = "us-east-1-phl-1a"
      region      = "us-east-1"
      native_name = "use1-phl1-az1 (us-east-1-phl-1a)"
      latitude    = 39.95
      longitude   = -75.17
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.queretaro.1" = {
      azid        = "use1-qro1-az1"
      zone        = "us-east-1-qro-1a"
      region      = "us-east-1"
      native_name = "use1-qro1-az1 (us-east-1-qro-1a)"
      latitude    = 23.63
      longitude   = -102.55
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.santiago.1" = {
      azid        = "use1-scl1-az1"
      zone        = "us-east-1-scl-1a"
      region      = "us-east-1"
      native_name = "use1-scl1-az1 (us-east-1-scl-1a)"
      latitude    = -33.45
      longitude   = -70.67
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.ohio.1" = {
      azid        = "use2-az1"
      zone        = "us-east-2a"
      region      = "us-east-2"
      native_name = "use2-az1 (us-east-2a)"
      latitude    = 40.42
      longitude   = -82.91
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.ohio.2" = {
      azid        = "use2-az2"
      zone        = "us-east-2b"
      region      = "us-east-2"
      native_name = "use2-az2 (us-east-2b)"
      latitude    = 40.42
      longitude   = -82.91
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.ohio.3" = {
      azid        = "use2-az3"
      zone        = "us-east-2c"
      region      = "us-east-2"
      native_name = "use2-az3 (us-east-2c)"
      latitude    = 40.42
      longitude   = -82.91
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.sanjose.1" = {
      azid        = "usw1-az1"
      zone        = "us-west-1a"
      region      = "us-west-1"
      native_name = "usw1-az1 (us-west-1a)"
      latitude    = 37.34
      longitude   = -121.89
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.sanjose.3" = {
      azid        = "usw1-az3"
      zone        = "us-west-1c"
      region      = "us-west-1"
      native_name = "usw1-az3 (us-west-1c)"
      latitude    = 37.34
      longitude   = -121.89
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.oregon.1" = {
      azid        = "usw2-az1"
      zone        = "us-west-2b"
      region      = "us-west-2"
      native_name = "usw2-az1 (us-west-2b)"
      latitude    = 45.84
      longitude   = -119.70
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.oregon.2" = {
      azid        = "usw2-az2"
      zone        = "us-west-2a"
      region      = "us-west-2"
      native_name = "usw2-az2 (us-west-2a)"
      latitude    = 45.84
      longitude   = -119.70
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.oregon.3" = {
      azid        = "usw2-az3"
      zone        = "us-west-2c"
      region      = "us-west-2"
      native_name = "usw2-az3 (us-west-2c)"
      latitude    = 45.84
      longitude   = -119.70
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.oregon.4" = {
      azid        = "usw2-az4"
      zone        = "us-west-2d"
      region      = "us-west-2"
      native_name = "usw2-az4 (us-west-2d)"
      latitude    = 45.84
      longitude   = -119.70
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.denver.1" = {
      azid        = "usw2-den1-az1"
      zone        = "us-west-2-den-1a"
      region      = "us-west-2"
      native_name = "usw2-den1-az1 (us-west-2-den-1a)"
      latitude    = 39.74
      longitude   = -104.99
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.lasvegas.1" = {
      azid        = "usw2-las1-az1"
      zone        = "us-west-2-las-1a"
      region      = "us-west-2"
      native_name = "usw2-las1-az1 (us-west-2-las-1a)"
      latitude    = 36.17
      longitude   = -115.14
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.losangeles.1" = {
      azid        = "usw2-lax1-az1"
      zone        = "us-west-2-lax-1a"
      region      = "us-west-2"
      native_name = "usw2-lax1-az1 (us-west-2-lax-1a)"
      latitude    = 34.05
      longitude   = -118.24
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.losangeles.2" = {
      azid        = "usw2-lax1-az2"
      zone        = "us-west-2-lax-1b"
      region      = "us-west-2"
      native_name = "usw2-lax1-az2 (us-west-2-lax-1b)"
      latitude    = 34.05
      longitude   = -118.24
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.portland.1" = {
      azid        = "usw2-pdx1-az1"
      zone        = "us-west-2-pdx-1a"
      region      = "us-west-2"
      native_name = "usw2-pdx1-az1 (us-west-2-pdx-1a)"
      latitude    = 45.52
      longitude   = -122.68
      seller_name = "Amazon"
      seller_code = "amazon"
    }

    "amazon.seattle.1" = {
      azid        = "usw2-sea1-az1"
      zone        = "us-west-2-sea-1a"
      region      = "us-west-2"
      native_name = "usw2-sea1-az1 (us-west-2-sea-1a)"
      latitude    = 47.61
      longitude   = -122.33
      seller_name = "Amazon"
      seller_code = "amazon"
    }

  }

  regions = [
    "ap-south-2",
    "ap-south-1",
    "eu-south-1",
    "eu-south-2",
    "me-central-1",
    "il-central-1",
    "ca-central-1",
    "ap-east-2",
    "mx-central-1",
    "eu-central-1",
    "eu-central-2",
    "us-west-1",
    "us-west-2",
    "af-south-1",
    "eu-north-1",
    "eu-west-3",
    "eu-west-2",
    "eu-west-1",
    "ap-northeast-3",
    "ap-northeast-2",
    "me-south-1",
    "ap-northeast-1",
    "sa-east-1",
    "ap-east-1",
    "ca-west-1",
    "ap-southeast-1",
    "ap-southeast-2",
    "ap-southeast-3",
    "ap-southeast-4",
    "us-east-1",
    "ap-southeast-5",
    "us-east-2",
    "ap-southeast-7",
  ]
}

locals {

  relays = {

    "amazon.ohio.1" = { datacenter_name = "amazon.ohio.1" },
    "amazon.ohio.2" = { datacenter_name = "amazon.ohio.2" },
  }

}

module "relay_amazon_ohio_1" {
	  source            = "./relay"
	  name              = "amazon.ohio.1"
	  zone              = local.datacenter_map["amazon.ohio.1"].zone
	  region            = local.datacenter_map["amazon.ohio.1"].region
	  type              = "m5a.xlarge"
	  ami               = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
	  security_group_id = module.region_us_east_2.security_group_id
	  vpn_address       = var.vpn_address
	  providers = {
	    aws = aws.us-east-2
	  }
	}
	module "relay_amazon_ohio_2" {
	  source            = "./relay"
	  name              = "amazon.ohio.2"
	  zone              = local.datacenter_map["amazon.ohio.2"].zone
	  region            = local.datacenter_map["amazon.ohio.2"].region
	  type              = "m5a.xlarge"
	  ami               = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
	  security_group_id = module.region_us_east_2.security_group_id
	  vpn_address       = var.vpn_address
	  providers = {
	    aws = aws.us-east-2
	  }
	}
	output "relays" {

	  description = "Data for each amazon relay setup by Terraform"

	  value = {

	    "amazon.ohio.1" = {
	      "relay_name"       = "amazon.ohio.1"
	      "datacenter_name"  = "amazon.ohio.1"
	      "seller_name"      = "Amazon"
	      "seller_code"      = "amazon"
	      "public_ip"        = module.relay_amazon_ohio_1.public_address
	      "public_port"      = 40000
	      "internal_ip"      = module.relay_amazon_ohio_1.internal_address
	      "internal_port"    = 40000
	      "internal_group"   = "us-east-2"
	      "ssh_ip"           = module.relay_amazon_ohio_1.public_address
	      "ssh_port"         = 22
	      "ssh_user"         = "ubuntu"
	      "bandwidth_price"  = 2
	    }

	    "amazon.ohio.2" = {
	      "relay_name"       = "amazon.ohio.2"
	      "datacenter_name"  = "amazon.ohio.2"
	      "seller_name"      = "Amazon"
	      "seller_code"      = "amazon"
	      "public_ip"        = module.relay_amazon_ohio_2.public_address
	      "public_port"      = 40000
	      "internal_ip"      = module.relay_amazon_ohio_2.internal_address
	      "internal_port"    = 40000
	      "internal_group"   = "us-east-2"
	      "ssh_ip"           = module.relay_amazon_ohio_2.public_address
	      "ssh_port"         = 22
	      "ssh_user"         = "ubuntu"
	      "bandwidth_price"  = 2
	    }

	
  }

}
