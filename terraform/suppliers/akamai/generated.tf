
locals {

  datacenter_map = {

    "akamai.southcarolina.2" = {
      zone   = "us-east1-b"
    }

    "akamai.southcarolina.3" = {
      zone   = "us-east1-c"
    }

    "akamai.southcarolina.4" = {
      zone   = "us-east1-d"
    }

    "akamai.virginia.3" = {
      zone   = "us-east4-c"
    }

    "akamai.virginia.2" = {
      zone   = "us-east4-b"
    }

    "akamai.virginia.1" = {
      zone   = "us-east4-a"
    }

    "akamai.iowa.3" = {
      zone   = "us-central1-c"
    }

    "akamai.iowa.1" = {
      zone   = "us-central1-a"
    }

    "akamai.iowa.6" = {
      zone   = "us-central1-f"
    }

    "akamai.iowa.2" = {
      zone   = "us-central1-b"
    }

    "akamai.oregon.2" = {
      zone   = "us-west1-b"
    }

    "akamai.oregon.3" = {
      zone   = "us-west1-c"
    }

    "akamai.oregon.1" = {
      zone   = "us-west1-a"
    }

    "akamai.netherlands.1" = {
      zone   = "europe-west4-a"
    }

    "akamai.netherlands.2" = {
      zone   = "europe-west4-b"
    }

    "akamai.netherlands.3" = {
      zone   = "europe-west4-c"
    }

    "akamai.belgium.2" = {
      zone   = "europe-west1-b"
    }

    "akamai.belgium.4" = {
      zone   = "europe-west1-d"
    }

    "akamai.belgium.3" = {
      zone   = "europe-west1-c"
    }

    "akamai.frankfurt.3" = {
      zone   = "europe-west3-c"
    }

    "akamai.frankfurt.1" = {
      zone   = "europe-west3-a"
    }

    "akamai.frankfurt.2" = {
      zone   = "europe-west3-b"
    }

    "akamai.london.3" = {
      zone   = "europe-west2-c"
    }

    "akamai.london.2" = {
      zone   = "europe-west2-b"
    }

    "akamai.london.1" = {
      zone   = "europe-west2-a"
    }

    "akamai.taiwan.2" = {
      zone   = "asia-east1-b"
    }

    "akamai.taiwan.1" = {
      zone   = "asia-east1-a"
    }

    "akamai.taiwan.3" = {
      zone   = "asia-east1-c"
    }

    "akamai.singapore.2" = {
      zone   = "asia-southeast1-b"
    }

    "akamai.singapore.1" = {
      zone   = "asia-southeast1-a"
    }

    "akamai.singapore.3" = {
      zone   = "asia-southeast1-c"
    }

    "akamai.tokyo.2" = {
      zone   = "asia-northeast1-b"
    }

    "akamai.tokyo.3" = {
      zone   = "asia-northeast1-c"
    }

    "akamai.tokyo.1" = {
      zone   = "asia-northeast1-a"
    }

    "akamai.mumbai.3" = {
      zone   = "asia-south1-c"
    }

    "akamai.mumbai.2" = {
      zone   = "asia-south1-b"
    }

    "akamai.mumbai.1" = {
      zone   = "asia-south1-a"
    }

    "akamai.sydney.2" = {
      zone   = "australia-southeast1-b"
    }

    "akamai.sydney.3" = {
      zone   = "australia-southeast1-c"
    }

    "akamai.sydney.1" = {
      zone   = "australia-southeast1-a"
    }

    "akamai.saopaulo.2" = {
      zone   = "southamerica-east1-b"
    }

    "akamai.saopaulo.3" = {
      zone   = "southamerica-east1-c"
    }

    "akamai.saopaulo.1" = {
      zone   = "southamerica-east1-a"
    }

    "akamai.hongkong.1" = {
      zone   = "asia-east2-a"
    }

    "akamai.hongkong.2" = {
      zone   = "asia-east2-b"
    }

    "akamai.hongkong.3" = {
      zone   = "asia-east2-c"
    }

    "akamai.osaka.1" = {
      zone   = "asia-northeast2-a"
    }

    "akamai.osaka.2" = {
      zone   = "asia-northeast2-b"
    }

    "akamai.osaka.3" = {
      zone   = "asia-northeast2-c"
    }

    "akamai.seoul.1" = {
      zone   = "asia-northeast3-a"
    }

    "akamai.seoul.2" = {
      zone   = "asia-northeast3-b"
    }

    "akamai.seoul.3" = {
      zone   = "asia-northeast3-c"
    }

    "akamai.delhi.1" = {
      zone   = "asia-south2-a"
    }

    "akamai.delhi.2" = {
      zone   = "asia-south2-b"
    }

    "akamai.delhi.3" = {
      zone   = "asia-south2-c"
    }

    "akamai.jakarta.1" = {
      zone   = "asia-southeast2-a"
    }

    "akamai.jakarta.2" = {
      zone   = "asia-southeast2-b"
    }

    "akamai.jakarta.3" = {
      zone   = "asia-southeast2-c"
    }

    "akamai.melbourne.1" = {
      zone   = "australia-southeast2-a"
    }

    "akamai.melbourne.2" = {
      zone   = "australia-southeast2-b"
    }

    "akamai.melbourne.3" = {
      zone   = "australia-southeast2-c"
    }

    "akamai.warsaw.1" = {
      zone   = "europe-central2-a"
    }

    "akamai.warsaw.2" = {
      zone   = "europe-central2-b"
    }

    "akamai.warsaw.3" = {
      zone   = "europe-central2-c"
    }

    "akamai.finland.1" = {
      zone   = "europe-north1-a"
    }

    "akamai.finland.2" = {
      zone   = "europe-north1-b"
    }

    "akamai.finland.3" = {
      zone   = "europe-north1-c"
    }

    "akamai.madrid.1" = {
      zone   = "europe-southwest1-a"
    }

    "akamai.madrid.2" = {
      zone   = "europe-southwest1-b"
    }

    "akamai.madrid.3" = {
      zone   = "europe-southwest1-c"
    }

    "akamai.zurich.1" = {
      zone   = "europe-west6-a"
    }

    "akamai.zurich.2" = {
      zone   = "europe-west6-b"
    }

    "akamai.zurich.3" = {
      zone   = "europe-west6-c"
    }

    "akamai.milan.1" = {
      zone   = "europe-west8-a"
    }

    "akamai.milan.2" = {
      zone   = "europe-west8-b"
    }

    "akamai.milan.3" = {
      zone   = "europe-west8-c"
    }

    "akamai.paris.1" = {
      zone   = "europe-west9-a"
    }

    "akamai.paris.2" = {
      zone   = "europe-west9-b"
    }

    "akamai.paris.3" = {
      zone   = "europe-west9-c"
    }

    "akamai.telaviv.1" = {
      zone   = "me-west1-a"
    }

    "akamai.telaviv.2" = {
      zone   = "me-west1-b"
    }

    "akamai.telaviv.3" = {
      zone   = "me-west1-c"
    }

    "akamai.montreal.1" = {
      zone   = "northamerica-northeast1-a"
    }

    "akamai.montreal.2" = {
      zone   = "northamerica-northeast1-b"
    }

    "akamai.montreal.3" = {
      zone   = "northamerica-northeast1-c"
    }

    "akamai.toronto.1" = {
      zone   = "northamerica-northeast2-a"
    }

    "akamai.toronto.2" = {
      zone   = "northamerica-northeast2-b"
    }

    "akamai.toronto.3" = {
      zone   = "northamerica-northeast2-c"
    }

    "akamai.santiago.1" = {
      zone   = "southamerica-west1-a"
    }

    "akamai.santiago.2" = {
      zone   = "southamerica-west1-b"
    }

    "akamai.santiago.3" = {
      zone   = "southamerica-west1-c"
    }

    "akamai.ohio.1" = {
      zone   = "us-east5-a"
    }

    "akamai.ohio.2" = {
      zone   = "us-east5-b"
    }

    "akamai.ohio.3" = {
      zone   = "us-east5-c"
    }

    "akamai.dallas.1" = {
      zone   = "us-south1-a"
    }

    "akamai.dallas.2" = {
      zone   = "us-south1-b"
    }

    "akamai.dallas.3" = {
      zone   = "us-south1-c"
    }

    "akamai.losangeles.1" = {
      zone   = "us-west2-a"
    }

    "akamai.losangeles.2" = {
      zone   = "us-west2-b"
    }

    "akamai.losangeles.3" = {
      zone   = "us-west2-c"
    }

    "akamai.saltlakecity.1" = {
      zone   = "us-west3-a"
    }

    "akamai.saltlakecity.2" = {
      zone   = "us-west3-b"
    }

    "akamai.saltlakecity.3" = {
      zone   = "us-west3-c"
    }

    "akamai.lasvegas.1" = {
      zone   = "us-west4-a"
    }

    "akamai.lasvegas.2" = {
      zone   = "us-west4-b"
    }

    "akamai.lasvegas.3" = {
      zone   = "us-west4-c"
    }

  }

}
