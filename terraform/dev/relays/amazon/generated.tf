
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
  alias                    = "ca-central-1"
  region                   = "ca-central-1"
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
  alias                    = "us-east-2"
  region                   = "us-east-2"
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

module "region_ca_central_1" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.ca-central-1
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

module "region_us_east_2" { 
  source              = "./region"
  vpn_address         = var.vpn_address
  ssh_public_key_file = var.ssh_public_key_file
  providers = {
    aws = aws.us-east-2
  }
}

locals {

  datacenter_map = {

    "amazon.johannesburg.1" = {
      azid   = "afs1-az1"
      zone   = "af-south-1a"
      region = "af-south-1"
    }

    "amazon.johannesburg.2" = {
      azid   = "afs1-az2"
      zone   = "af-south-1b"
      region = "af-south-1"
    }

    "amazon.johannesburg.3" = {
      azid   = "afs1-az3"
      zone   = "af-south-1c"
      region = "af-south-1"
    }

    "amazon.hongkong.1" = {
      azid   = "ape1-az1"
      zone   = "ap-east-1a"
      region = "ap-east-1"
    }

    "amazon.hongkong.2" = {
      azid   = "ape1-az2"
      zone   = "ap-east-1b"
      region = "ap-east-1"
    }

    "amazon.hongkong.3" = {
      azid   = "ape1-az3"
      zone   = "ap-east-1c"
      region = "ap-east-1"
    }

    "amazon.tokyo.1" = {
      azid   = "apne1-az1"
      zone   = "ap-northeast-1c"
      region = "ap-northeast-1"
    }

    "amazon.tokyo.2" = {
      azid   = "apne1-az2"
      zone   = "ap-northeast-1d"
      region = "ap-northeast-1"
    }

    "amazon.tokyo.4" = {
      azid   = "apne1-az4"
      zone   = "ap-northeast-1a"
      region = "ap-northeast-1"
    }

    "amazon.seoul.1" = {
      azid   = "apne2-az1"
      zone   = "ap-northeast-2a"
      region = "ap-northeast-2"
    }

    "amazon.seoul.2" = {
      azid   = "apne2-az2"
      zone   = "ap-northeast-2b"
      region = "ap-northeast-2"
    }

    "amazon.seoul.3" = {
      azid   = "apne2-az3"
      zone   = "ap-northeast-2c"
      region = "ap-northeast-2"
    }

    "amazon.seoul.4" = {
      azid   = "apne2-az4"
      zone   = "ap-northeast-2d"
      region = "ap-northeast-2"
    }

    "amazon.osaka.1" = {
      azid   = "apne3-az1"
      zone   = "ap-northeast-3b"
      region = "ap-northeast-3"
    }

    "amazon.osaka.2" = {
      azid   = "apne3-az2"
      zone   = "ap-northeast-3c"
      region = "ap-northeast-3"
    }

    "amazon.osaka.3" = {
      azid   = "apne3-az3"
      zone   = "ap-northeast-3a"
      region = "ap-northeast-3"
    }

    "amazon.mumbai.1" = {
      azid   = "aps1-az1"
      zone   = "ap-south-1a"
      region = "ap-south-1"
    }

    "amazon.mumbai.2" = {
      azid   = "aps1-az2"
      zone   = "ap-south-1c"
      region = "ap-south-1"
    }

    "amazon.mumbai.3" = {
      azid   = "aps1-az3"
      zone   = "ap-south-1b"
      region = "ap-south-1"
    }

    "amazon.hyderabad.1" = {
      azid   = "aps2-az1"
      zone   = "ap-south-2a"
      region = "ap-south-2"
    }

    "amazon.hyderabad.2" = {
      azid   = "aps2-az2"
      zone   = "ap-south-2b"
      region = "ap-south-2"
    }

    "amazon.hyderabad.3" = {
      azid   = "aps2-az3"
      zone   = "ap-south-2c"
      region = "ap-south-2"
    }

    "amazon.singapore.1" = {
      azid   = "apse1-az1"
      zone   = "ap-southeast-1a"
      region = "ap-southeast-1"
    }

    "amazon.singapore.2" = {
      azid   = "apse1-az2"
      zone   = "ap-southeast-1b"
      region = "ap-southeast-1"
    }

    "amazon.singapore.3" = {
      azid   = "apse1-az3"
      zone   = "ap-southeast-1c"
      region = "ap-southeast-1"
    }

    "amazon.sydney.1" = {
      azid   = "apse2-az1"
      zone   = "ap-southeast-2b"
      region = "ap-southeast-2"
    }

    "amazon.sydney.2" = {
      azid   = "apse2-az2"
      zone   = "ap-southeast-2c"
      region = "ap-southeast-2"
    }

    "amazon.sydney.3" = {
      azid   = "apse2-az3"
      zone   = "ap-southeast-2a"
      region = "ap-southeast-2"
    }

    "amazon.jakarta.1" = {
      azid   = "apse3-az1"
      zone   = "ap-southeast-3a"
      region = "ap-southeast-3"
    }

    "amazon.jakarta.2" = {
      azid   = "apse3-az2"
      zone   = "ap-southeast-3b"
      region = "ap-southeast-3"
    }

    "amazon.jakarta.3" = {
      azid   = "apse3-az3"
      zone   = "ap-southeast-3c"
      region = "ap-southeast-3"
    }

    "amazon.melbourne.1" = {
      azid   = "apse4-az1"
      zone   = "ap-southeast-4a"
      region = "ap-southeast-4"
    }

    "amazon.melbourne.2" = {
      azid   = "apse4-az2"
      zone   = "ap-southeast-4b"
      region = "ap-southeast-4"
    }

    "amazon.melbourne.3" = {
      azid   = "apse4-az3"
      zone   = "ap-southeast-4c"
      region = "ap-southeast-4"
    }

    "amazon.montreal.1" = {
      azid   = "cac1-az1"
      zone   = "ca-central-1a"
      region = "ca-central-1"
    }

    "amazon.montreal.2" = {
      azid   = "cac1-az2"
      zone   = "ca-central-1b"
      region = "ca-central-1"
    }

    "amazon.montreal.4" = {
      azid   = "cac1-az4"
      zone   = "ca-central-1d"
      region = "ca-central-1"
    }

    "amazon.frankfurt.1" = {
      azid   = "euc1-az1"
      zone   = "eu-central-1c"
      region = "eu-central-1"
    }

    "amazon.frankfurt.2" = {
      azid   = "euc1-az2"
      zone   = "eu-central-1a"
      region = "eu-central-1"
    }

    "amazon.frankfurt.3" = {
      azid   = "euc1-az3"
      zone   = "eu-central-1b"
      region = "eu-central-1"
    }

    "amazon.zurich.1" = {
      azid   = "euc2-az1"
      zone   = "eu-central-2a"
      region = "eu-central-2"
    }

    "amazon.zurich.2" = {
      azid   = "euc2-az2"
      zone   = "eu-central-2b"
      region = "eu-central-2"
    }

    "amazon.zurich.3" = {
      azid   = "euc2-az3"
      zone   = "eu-central-2c"
      region = "eu-central-2"
    }

    "amazon.stockholm.1" = {
      azid   = "eun1-az1"
      zone   = "eu-north-1a"
      region = "eu-north-1"
    }

    "amazon.stockholm.2" = {
      azid   = "eun1-az2"
      zone   = "eu-north-1b"
      region = "eu-north-1"
    }

    "amazon.stockholm.3" = {
      azid   = "eun1-az3"
      zone   = "eu-north-1c"
      region = "eu-north-1"
    }

    "amazon.milan.1" = {
      azid   = "eus1-az1"
      zone   = "eu-south-1a"
      region = "eu-south-1"
    }

    "amazon.milan.2" = {
      azid   = "eus1-az2"
      zone   = "eu-south-1b"
      region = "eu-south-1"
    }

    "amazon.milan.3" = {
      azid   = "eus1-az3"
      zone   = "eu-south-1c"
      region = "eu-south-1"
    }

    "amazon.spain.1" = {
      azid   = "eus2-az1"
      zone   = "eu-south-2a"
      region = "eu-south-2"
    }

    "amazon.spain.2" = {
      azid   = "eus2-az2"
      zone   = "eu-south-2b"
      region = "eu-south-2"
    }

    "amazon.spain.3" = {
      azid   = "eus2-az3"
      zone   = "eu-south-2c"
      region = "eu-south-2"
    }

    "amazon.ireland.1" = {
      azid   = "euw1-az1"
      zone   = "eu-west-1a"
      region = "eu-west-1"
    }

    "amazon.ireland.2" = {
      azid   = "euw1-az2"
      zone   = "eu-west-1b"
      region = "eu-west-1"
    }

    "amazon.ireland.3" = {
      azid   = "euw1-az3"
      zone   = "eu-west-1c"
      region = "eu-west-1"
    }

    "amazon.london.1" = {
      azid   = "euw2-az1"
      zone   = "eu-west-2c"
      region = "eu-west-2"
    }

    "amazon.london.2" = {
      azid   = "euw2-az2"
      zone   = "eu-west-2a"
      region = "eu-west-2"
    }

    "amazon.london.3" = {
      azid   = "euw2-az3"
      zone   = "eu-west-2b"
      region = "eu-west-2"
    }

    "amazon.paris.1" = {
      azid   = "euw3-az1"
      zone   = "eu-west-3a"
      region = "eu-west-3"
    }

    "amazon.paris.2" = {
      azid   = "euw3-az2"
      zone   = "eu-west-3b"
      region = "eu-west-3"
    }

    "amazon.paris.3" = {
      azid   = "euw3-az3"
      zone   = "eu-west-3c"
      region = "eu-west-3"
    }

    "amazon.uae.1" = {
      azid   = "mec1-az1"
      zone   = "me-central-1a"
      region = "me-central-1"
    }

    "amazon.uae.2" = {
      azid   = "mec1-az2"
      zone   = "me-central-1b"
      region = "me-central-1"
    }

    "amazon.uae.3" = {
      azid   = "mec1-az3"
      zone   = "me-central-1c"
      region = "me-central-1"
    }

    "amazon.bahrain.1" = {
      azid   = "mes1-az1"
      zone   = "me-south-1a"
      region = "me-south-1"
    }

    "amazon.bahrain.2" = {
      azid   = "mes1-az2"
      zone   = "me-south-1b"
      region = "me-south-1"
    }

    "amazon.bahrain.3" = {
      azid   = "mes1-az3"
      zone   = "me-south-1c"
      region = "me-south-1"
    }

    "amazon.saopaulo.1" = {
      azid   = "sae1-az1"
      zone   = "sa-east-1a"
      region = "sa-east-1"
    }

    "amazon.saopaulo.2" = {
      azid   = "sae1-az2"
      zone   = "sa-east-1b"
      region = "sa-east-1"
    }

    "amazon.saopaulo.3" = {
      azid   = "sae1-az3"
      zone   = "sa-east-1c"
      region = "sa-east-1"
    }

    "amazon.virginia.1" = {
      azid   = "use1-az1"
      zone   = "us-east-1c"
      region = "us-east-1"
    }

    "amazon.virginia.2" = {
      azid   = "use1-az2"
      zone   = "us-east-1d"
      region = "us-east-1"
    }

    "amazon.virginia.3" = {
      azid   = "use1-az3"
      zone   = "us-east-1e"
      region = "us-east-1"
    }

    "amazon.virginia.4" = {
      azid   = "use1-az4"
      zone   = "us-east-1a"
      region = "us-east-1"
    }

    "amazon.virginia.5" = {
      azid   = "use1-az5"
      zone   = "us-east-1f"
      region = "us-east-1"
    }

    "amazon.virginia.6" = {
      azid   = "use1-az6"
      zone   = "us-east-1b"
      region = "us-east-1"
    }

    "amazon.ohio.1" = {
      azid   = "use2-az1"
      zone   = "us-east-2a"
      region = "us-east-2"
    }

    "amazon.ohio.2" = {
      azid   = "use2-az2"
      zone   = "us-east-2b"
      region = "us-east-2"
    }

    "amazon.ohio.3" = {
      azid   = "use2-az3"
      zone   = "us-east-2c"
      region = "us-east-2"
    }

    "amazon.sanjose.1" = {
      azid   = "usw1-az1"
      zone   = "us-west-1b"
      region = "us-west-1"
    }

    "amazon.sanjose.3" = {
      azid   = "usw1-az3"
      zone   = "us-west-1a"
      region = "us-west-1"
    }

    "amazon.oregon.1" = {
      azid   = "usw2-az1"
      zone   = "us-west-2b"
      region = "us-west-2"
    }

    "amazon.oregon.2" = {
      azid   = "usw2-az2"
      zone   = "us-west-2a"
      region = "us-west-2"
    }

    "amazon.oregon.3" = {
      azid   = "usw2-az3"
      zone   = "us-west-2c"
      region = "us-west-2"
    }

    "amazon.oregon.4" = {
      azid   = "usw2-az4"
      zone   = "us-west-2d"
      region = "us-west-2"
    }

  }

  regions = [
    "ap-south-2",
    "ap-south-1",
    "eu-south-1",
    "eu-south-2",
    "me-central-1",
    "ca-central-1",
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
    "ap-southeast-1",
    "ap-southeast-2",
    "ap-southeast-3",
    "ap-southeast-4",
    "us-east-1",
    "us-east-2",
  ]
}

locals {
  
  relays = {

    "amazon.virginia.1" = {
      datacenter_name = "amazon.virginia.1"
    }    

    "amazon.tokyo.1" = {
      datacenter_name = "amazon.tokyo.1"
    }    

  }

}

module "relay_amazon_virginia_1" {
  source            = "./relay"
  name              = "amazon.virginia.1"
  zone              = local.datacenter_map["amazon.virginia.1"].zone
  region            = local.datacenter_map["amazon.virginia.1"].region
  type              = "m5a.large"
  ami               = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
  security_group_id = module.region_us_east_1.security_group_id
  providers = {
    aws = aws.us-east-1
  }
}

module "relay_amazon_tokyo_1" {
  source            = "./relay"
  name              = "amazon.tokyo.1"
  zone              = local.datacenter_map["amazon.tokyo.1"].zone
  region            = local.datacenter_map["amazon.tokyo.1"].region
  type              = "m5a.large"
  ami               = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
  security_group_id = module.region_ap_northeast_1.security_group_id
  providers = {
    aws = aws.ap-northeast-1
  }
}

output "relays" {

  description = "Data for each amazon relay setup by Terraform"

  value = {

    "amazon.virginia.1" = {
      "relay_name"       = "amazon.virginia.1"
      "datacenter_name"  = "amazon.virginia.1"
      "supplier_name"    = "amazon"
      "public_address"   = "${module.relay_amazon_virginia_1.public_address}:40000"
      "internal_address" = "${module.relay_amazon_virginia_1.internal_address}:40000"
      "internal_group"   = "amazon.virginia.1"
      "ssh_address"      = "${module.relay_amazon_virginia_1.public_address}:22"
      "ssh_user"         = "ubuntu"
    }    

    "amazon.tokyo.1" = {
      "relay_name"       = "amazon.tokyo.1"
      "datacenter_name"  = "amazon.tokyo.1"
      "supplier_name"    = "amazon"
      "public_address"   = "${module.relay_amazon_tokyo_1.public_address}:40000"
      "internal_address" = "${module.relay_amazon_tokyo_1.internal_address}:40000"
      "internal_group"   = "amazon.tokyo.1"
      "ssh_address"      = "${module.relay_amazon_tokyo_1.public_address}:22"
      "ssh_user"         = "ubuntu"
    }    

  }
}
