
locals {

  datacenter_map = {

    "google.taiwan.1" = {
      zone   = "asia-east1-a"
      region = "asia-east"
    }

    "google.taiwan.2" = {
      zone   = "asia-east1-b"
      region = "asia-east"
    }

    "google.taiwan.3" = {
      zone   = "asia-east1-c"
      region = "asia-east"
    }

    "google.hongkong.1" = {
      zone   = "asia-east2-a"
      region = "asia-east"
    }

    "google.hongkong.2" = {
      zone   = "asia-east2-b"
      region = "asia-east"
    }

    "google.hongkong.3" = {
      zone   = "asia-east2-c"
      region = "asia-east"
    }

    "google.tokyo.1" = {
      zone   = "asia-northeast1-a"
      region = "asia-northeast"
    }

    "google.tokyo.2" = {
      zone   = "asia-northeast1-b"
      region = "asia-northeast"
    }

    "google.tokyo.3" = {
      zone   = "asia-northeast1-c"
      region = "asia-northeast"
    }

    "google.osaka.1" = {
      zone   = "asia-northeast2-a"
      region = "asia-northeast"
    }

    "google.osaka.2" = {
      zone   = "asia-northeast2-b"
      region = "asia-northeast"
    }

    "google.osaka.3" = {
      zone   = "asia-northeast2-c"
      region = "asia-northeast"
    }

    "google.seoul.1" = {
      zone   = "asia-northeast3-a"
      region = "asia-northeast"
    }

    "google.seoul.2" = {
      zone   = "asia-northeast3-b"
      region = "asia-northeast"
    }

    "google.seoul.3" = {
      zone   = "asia-northeast3-c"
      region = "asia-northeast"
    }

    "google.mumbai.1" = {
      zone   = "asia-south1-a"
      region = "asia-south"
    }

    "google.mumbai.2" = {
      zone   = "asia-south1-b"
      region = "asia-south"
    }

    "google.mumbai.3" = {
      zone   = "asia-south1-c"
      region = "asia-south"
    }

    "google.delhi.1" = {
      zone   = "asia-south2-a"
      region = "asia-south"
    }

    "google.delhi.2" = {
      zone   = "asia-south2-b"
      region = "asia-south"
    }

    "google.delhi.3" = {
      zone   = "asia-south2-c"
      region = "asia-south"
    }

    "google.singapore.1" = {
      zone   = "asia-southeast1-a"
      region = "asia-southeast"
    }

    "google.singapore.2" = {
      zone   = "asia-southeast1-b"
      region = "asia-southeast"
    }

    "google.singapore.3" = {
      zone   = "asia-southeast1-c"
      region = "asia-southeast"
    }

    "google.jakarta.1" = {
      zone   = "asia-southeast2-a"
      region = "asia-southeast"
    }

    "google.jakarta.2" = {
      zone   = "asia-southeast2-b"
      region = "asia-southeast"
    }

    "google.jakarta.3" = {
      zone   = "asia-southeast2-c"
      region = "asia-southeast"
    }

    "google.sydney.1" = {
      zone   = "australia-southeast1-a"
      region = "australia-southeast"
    }

    "google.sydney.2" = {
      zone   = "australia-southeast1-b"
      region = "australia-southeast"
    }

    "google.sydney.3" = {
      zone   = "australia-southeast1-c"
      region = "australia-southeast"
    }

    "google.melbourne.1" = {
      zone   = "australia-southeast2-a"
      region = "australia-southeast"
    }

    "google.melbourne.2" = {
      zone   = "australia-southeast2-b"
      region = "australia-southeast"
    }

    "google.melbourne.3" = {
      zone   = "australia-southeast2-c"
      region = "australia-southeast"
    }

    "google.warsaw.1" = {
      zone   = "europe-central2-a"
      region = "europe-central"
    }

    "google.warsaw.2" = {
      zone   = "europe-central2-b"
      region = "europe-central"
    }

    "google.warsaw.3" = {
      zone   = "europe-central2-c"
      region = "europe-central"
    }

    "google.finland.1" = {
      zone   = "europe-north1-a"
      region = "europe-north"
    }

    "google.finland.2" = {
      zone   = "europe-north1-b"
      region = "europe-north"
    }

    "google.finland.3" = {
      zone   = "europe-north1-c"
      region = "europe-north"
    }

    "google.madrid.1" = {
      zone   = "europe-southwest1-a"
      region = "europe-southwest"
    }

    "google.madrid.2" = {
      zone   = "europe-southwest1-b"
      region = "europe-southwest"
    }

    "google.madrid.3" = {
      zone   = "europe-southwest1-c"
      region = "europe-southwest"
    }

    "google.belgium.2" = {
      zone   = "europe-west1-b"
      region = "europe-west"
    }

    "google.belgium.3" = {
      zone   = "europe-west1-c"
      region = "europe-west"
    }

    "google.belgium.4" = {
      zone   = "europe-west1-d"
      region = "europe-west"
    }

    "google.london.1" = {
      zone   = "europe-west2-a"
      region = "europe-west"
    }

    "google.london.2" = {
      zone   = "europe-west2-b"
      region = "europe-west"
    }

    "google.london.3" = {
      zone   = "europe-west2-c"
      region = "europe-west"
    }

    "google.frankfurt.1" = {
      zone   = "europe-west3-a"
      region = "europe-west"
    }

    "google.frankfurt.2" = {
      zone   = "europe-west3-b"
      region = "europe-west"
    }

    "google.frankfurt.3" = {
      zone   = "europe-west3-c"
      region = "europe-west"
    }

    "google.netherlands.1" = {
      zone   = "europe-west4-a"
      region = "europe-west"
    }

    "google.netherlands.2" = {
      zone   = "europe-west4-b"
      region = "europe-west"
    }

    "google.netherlands.3" = {
      zone   = "europe-west4-c"
      region = "europe-west"
    }

    "google.zurich.1" = {
      zone   = "europe-west6-a"
      region = "europe-west"
    }

    "google.zurich.2" = {
      zone   = "europe-west6-b"
      region = "europe-west"
    }

    "google.zurich.3" = {
      zone   = "europe-west6-c"
      region = "europe-west"
    }

    "google.milan.1" = {
      zone   = "europe-west8-a"
      region = "europe-west"
    }

    "google.milan.2" = {
      zone   = "europe-west8-b"
      region = "europe-west"
    }

    "google.milan.3" = {
      zone   = "europe-west8-c"
      region = "europe-west"
    }

    "google.paris.1" = {
      zone   = "europe-west9-a"
      region = "europe-west"
    }

    "google.paris.2" = {
      zone   = "europe-west9-b"
      region = "europe-west"
    }

    "google.paris.3" = {
      zone   = "europe-west9-c"
      region = "europe-west"
    }

    "google.telaviv.1" = {
      zone   = "me-west1-a"
      region = "me-west"
    }

    "google.telaviv.2" = {
      zone   = "me-west1-b"
      region = "me-west"
    }

    "google.telaviv.3" = {
      zone   = "me-west1-c"
      region = "me-west"
    }

    "google.montreal.1" = {
      zone   = "northamerica-northeast1-a"
      region = "northamerica-northeast"
    }

    "google.montreal.2" = {
      zone   = "northamerica-northeast1-b"
      region = "northamerica-northeast"
    }

    "google.montreal.3" = {
      zone   = "northamerica-northeast1-c"
      region = "northamerica-northeast"
    }

    "google.toronto.1" = {
      zone   = "northamerica-northeast2-a"
      region = "northamerica-northeast"
    }

    "google.toronto.2" = {
      zone   = "northamerica-northeast2-b"
      region = "northamerica-northeast"
    }

    "google.toronto.3" = {
      zone   = "northamerica-northeast2-c"
      region = "northamerica-northeast"
    }

    "google.saopaulo.1" = {
      zone   = "southamerica-east1-a"
      region = "southamerica-east"
    }

    "google.saopaulo.2" = {
      zone   = "southamerica-east1-b"
      region = "southamerica-east"
    }

    "google.saopaulo.3" = {
      zone   = "southamerica-east1-c"
      region = "southamerica-east"
    }

    "google.santiago.1" = {
      zone   = "southamerica-west1-a"
      region = "southamerica-west"
    }

    "google.santiago.2" = {
      zone   = "southamerica-west1-b"
      region = "southamerica-west"
    }

    "google.santiago.3" = {
      zone   = "southamerica-west1-c"
      region = "southamerica-west"
    }

    "google.iowa.1" = {
      zone   = "us-central1-a"
      region = "us-central"
    }

    "google.iowa.2" = {
      zone   = "us-central1-b"
      region = "us-central"
    }

    "google.iowa.3" = {
      zone   = "us-central1-c"
      region = "us-central"
    }

    "google.iowa.6" = {
      zone   = "us-central1-f"
      region = "us-central"
    }

    "google.southcarolina.2" = {
      zone   = "us-east1-b"
      region = "us-east"
    }

    "google.southcarolina.3" = {
      zone   = "us-east1-c"
      region = "us-east"
    }

    "google.southcarolina.4" = {
      zone   = "us-east1-d"
      region = "us-east"
    }

    "google.virginia.1" = {
      zone   = "us-east4-a"
      region = "us-east"
    }

    "google.virginia.2" = {
      zone   = "us-east4-b"
      region = "us-east"
    }

    "google.virginia.3" = {
      zone   = "us-east4-c"
      region = "us-east"
    }

    "google.ohio.1" = {
      zone   = "us-east5-a"
      region = "us-east"
    }

    "google.ohio.2" = {
      zone   = "us-east5-b"
      region = "us-east"
    }

    "google.ohio.3" = {
      zone   = "us-east5-c"
      region = "us-east"
    }

    "google.dallas.1" = {
      zone   = "us-south1-a"
      region = "us-south"
    }

    "google.dallas.2" = {
      zone   = "us-south1-b"
      region = "us-south"
    }

    "google.dallas.3" = {
      zone   = "us-south1-c"
      region = "us-south"
    }

    "google.oregon.1" = {
      zone   = "us-west1-a"
      region = "us-west"
    }

    "google.oregon.2" = {
      zone   = "us-west1-b"
      region = "us-west"
    }

    "google.oregon.3" = {
      zone   = "us-west1-c"
      region = "us-west"
    }

    "google.losangeles.1" = {
      zone   = "us-west2-a"
      region = "us-west"
    }

    "google.losangeles.2" = {
      zone   = "us-west2-b"
      region = "us-west"
    }

    "google.losangeles.3" = {
      zone   = "us-west2-c"
      region = "us-west"
    }

    "google.saltlakecity.1" = {
      zone   = "us-west3-a"
      region = "us-west"
    }

    "google.saltlakecity.2" = {
      zone   = "us-west3-b"
      region = "us-west"
    }

    "google.saltlakecity.3" = {
      zone   = "us-west3-c"
      region = "us-west"
    }

    "google.lasvegas.1" = {
      zone   = "us-west4-a"
      region = "us-west"
    }

    "google.lasvegas.2" = {
      zone   = "us-west4-b"
      region = "us-west"
    }

    "google.lasvegas.3" = {
      zone   = "us-west4-c"
      region = "us-west"
    }

  }

}
