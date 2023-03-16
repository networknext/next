
locals {

  datacenter_map = {

    "google.taiwan.1" = {
      zone   = "asia-east1-a"
      region = "asia-east1"
    }

    "google.taiwan.2" = {
      zone   = "asia-east1-b"
      region = "asia-east1"
    }

    "google.taiwan.3" = {
      zone   = "asia-east1-c"
      region = "asia-east1"
    }

    "google.hongkong.1" = {
      zone   = "asia-east2-a"
      region = "asia-east2"
    }

    "google.hongkong.2" = {
      zone   = "asia-east2-b"
      region = "asia-east2"
    }

    "google.hongkong.3" = {
      zone   = "asia-east2-c"
      region = "asia-east2"
    }

    "google.tokyo.1" = {
      zone   = "asia-northeast1-a"
      region = "asia-northeast1"
    }

    "google.tokyo.2" = {
      zone   = "asia-northeast1-b"
      region = "asia-northeast1"
    }

    "google.tokyo.3" = {
      zone   = "asia-northeast1-c"
      region = "asia-northeast1"
    }

    "google.osaka.1" = {
      zone   = "asia-northeast2-a"
      region = "asia-northeast2"
    }

    "google.osaka.2" = {
      zone   = "asia-northeast2-b"
      region = "asia-northeast2"
    }

    "google.osaka.3" = {
      zone   = "asia-northeast2-c"
      region = "asia-northeast2"
    }

    "google.seoul.1" = {
      zone   = "asia-northeast3-a"
      region = "asia-northeast3"
    }

    "google.seoul.2" = {
      zone   = "asia-northeast3-b"
      region = "asia-northeast3"
    }

    "google.seoul.3" = {
      zone   = "asia-northeast3-c"
      region = "asia-northeast3"
    }

    "google.mumbai.1" = {
      zone   = "asia-south1-a"
      region = "asia-south1"
    }

    "google.mumbai.2" = {
      zone   = "asia-south1-b"
      region = "asia-south1"
    }

    "google.mumbai.3" = {
      zone   = "asia-south1-c"
      region = "asia-south1"
    }

    "google.delhi.1" = {
      zone   = "asia-south2-a"
      region = "asia-south2"
    }

    "google.delhi.2" = {
      zone   = "asia-south2-b"
      region = "asia-south2"
    }

    "google.delhi.3" = {
      zone   = "asia-south2-c"
      region = "asia-south2"
    }

    "google.singapore.1" = {
      zone   = "asia-southeast1-a"
      region = "asia-southeast1"
    }

    "google.singapore.2" = {
      zone   = "asia-southeast1-b"
      region = "asia-southeast1"
    }

    "google.singapore.3" = {
      zone   = "asia-southeast1-c"
      region = "asia-southeast1"
    }

    "google.jakarta.1" = {
      zone   = "asia-southeast2-a"
      region = "asia-southeast2"
    }

    "google.jakarta.2" = {
      zone   = "asia-southeast2-b"
      region = "asia-southeast2"
    }

    "google.jakarta.3" = {
      zone   = "asia-southeast2-c"
      region = "asia-southeast2"
    }

    "google.sydney.1" = {
      zone   = "australia-southeast1-a"
      region = "australia-southeast1"
    }

    "google.sydney.2" = {
      zone   = "australia-southeast1-b"
      region = "australia-southeast1"
    }

    "google.sydney.3" = {
      zone   = "australia-southeast1-c"
      region = "australia-southeast1"
    }

    "google.melbourne.1" = {
      zone   = "australia-southeast2-a"
      region = "australia-southeast2"
    }

    "google.melbourne.2" = {
      zone   = "australia-southeast2-b"
      region = "australia-southeast2"
    }

    "google.melbourne.3" = {
      zone   = "australia-southeast2-c"
      region = "australia-southeast2"
    }

    "google.warsaw.1" = {
      zone   = "europe-central2-a"
      region = "europe-central2"
    }

    "google.warsaw.2" = {
      zone   = "europe-central2-b"
      region = "europe-central2"
    }

    "google.warsaw.3" = {
      zone   = "europe-central2-c"
      region = "europe-central2"
    }

    "google.finland.1" = {
      zone   = "europe-north1-a"
      region = "europe-north1"
    }

    "google.finland.2" = {
      zone   = "europe-north1-b"
      region = "europe-north1"
    }

    "google.finland.3" = {
      zone   = "europe-north1-c"
      region = "europe-north1"
    }

    "google.madrid.1" = {
      zone   = "europe-southwest1-a"
      region = "europe-southwest1"
    }

    "google.madrid.2" = {
      zone   = "europe-southwest1-b"
      region = "europe-southwest1"
    }

    "google.madrid.3" = {
      zone   = "europe-southwest1-c"
      region = "europe-southwest1"
    }

    "google.belgium.2" = {
      zone   = "europe-west1-b"
      region = "europe-west1"
    }

    "google.belgium.3" = {
      zone   = "europe-west1-c"
      region = "europe-west1"
    }

    "google.belgium.4" = {
      zone   = "europe-west1-d"
      region = "europe-west1"
    }

    "google.london.1" = {
      zone   = "europe-west2-a"
      region = "europe-west2"
    }

    "google.london.2" = {
      zone   = "europe-west2-b"
      region = "europe-west2"
    }

    "google.london.3" = {
      zone   = "europe-west2-c"
      region = "europe-west2"
    }

    "google.frankfurt.1" = {
      zone   = "europe-west3-a"
      region = "europe-west3"
    }

    "google.frankfurt.2" = {
      zone   = "europe-west3-b"
      region = "europe-west3"
    }

    "google.frankfurt.3" = {
      zone   = "europe-west3-c"
      region = "europe-west3"
    }

    "google.netherlands.1" = {
      zone   = "europe-west4-a"
      region = "europe-west4"
    }

    "google.netherlands.2" = {
      zone   = "europe-west4-b"
      region = "europe-west4"
    }

    "google.netherlands.3" = {
      zone   = "europe-west4-c"
      region = "europe-west4"
    }

    "google.zurich.1" = {
      zone   = "europe-west6-a"
      region = "europe-west6"
    }

    "google.zurich.2" = {
      zone   = "europe-west6-b"
      region = "europe-west6"
    }

    "google.zurich.3" = {
      zone   = "europe-west6-c"
      region = "europe-west6"
    }

    "google.milan.1" = {
      zone   = "europe-west8-a"
      region = "europe-west8"
    }

    "google.milan.2" = {
      zone   = "europe-west8-b"
      region = "europe-west8"
    }

    "google.milan.3" = {
      zone   = "europe-west8-c"
      region = "europe-west8"
    }

    "google.paris.1" = {
      zone   = "europe-west9-a"
      region = "europe-west9"
    }

    "google.paris.2" = {
      zone   = "europe-west9-b"
      region = "europe-west9"
    }

    "google.paris.3" = {
      zone   = "europe-west9-c"
      region = "europe-west9"
    }

    "google.telaviv.1" = {
      zone   = "me-west1-a"
      region = "me-west1"
    }

    "google.telaviv.2" = {
      zone   = "me-west1-b"
      region = "me-west1"
    }

    "google.telaviv.3" = {
      zone   = "me-west1-c"
      region = "me-west1"
    }

    "google.montreal.1" = {
      zone   = "northamerica-northeast1-a"
      region = "northamerica-northeast1"
    }

    "google.montreal.2" = {
      zone   = "northamerica-northeast1-b"
      region = "northamerica-northeast1"
    }

    "google.montreal.3" = {
      zone   = "northamerica-northeast1-c"
      region = "northamerica-northeast1"
    }

    "google.toronto.1" = {
      zone   = "northamerica-northeast2-a"
      region = "northamerica-northeast2"
    }

    "google.toronto.2" = {
      zone   = "northamerica-northeast2-b"
      region = "northamerica-northeast2"
    }

    "google.toronto.3" = {
      zone   = "northamerica-northeast2-c"
      region = "northamerica-northeast2"
    }

    "google.saopaulo.1" = {
      zone   = "southamerica-east1-a"
      region = "southamerica-east1"
    }

    "google.saopaulo.2" = {
      zone   = "southamerica-east1-b"
      region = "southamerica-east1"
    }

    "google.saopaulo.3" = {
      zone   = "southamerica-east1-c"
      region = "southamerica-east1"
    }

    "google.santiago.1" = {
      zone   = "southamerica-west1-a"
      region = "southamerica-west1"
    }

    "google.santiago.2" = {
      zone   = "southamerica-west1-b"
      region = "southamerica-west1"
    }

    "google.santiago.3" = {
      zone   = "southamerica-west1-c"
      region = "southamerica-west1"
    }

    "google.iowa.1" = {
      zone   = "us-central1-a"
      region = "us-central1"
    }

    "google.iowa.2" = {
      zone   = "us-central1-b"
      region = "us-central1"
    }

    "google.iowa.3" = {
      zone   = "us-central1-c"
      region = "us-central1"
    }

    "google.iowa.6" = {
      zone   = "us-central1-f"
      region = "us-central1"
    }

    "google.southcarolina.2" = {
      zone   = "us-east1-b"
      region = "us-east1"
    }

    "google.southcarolina.3" = {
      zone   = "us-east1-c"
      region = "us-east1"
    }

    "google.southcarolina.4" = {
      zone   = "us-east1-d"
      region = "us-east1"
    }

    "google.virginia.1" = {
      zone   = "us-east4-a"
      region = "us-east4"
    }

    "google.virginia.2" = {
      zone   = "us-east4-b"
      region = "us-east4"
    }

    "google.virginia.3" = {
      zone   = "us-east4-c"
      region = "us-east4"
    }

    "google.ohio.1" = {
      zone   = "us-east5-a"
      region = "us-east5"
    }

    "google.ohio.2" = {
      zone   = "us-east5-b"
      region = "us-east5"
    }

    "google.ohio.3" = {
      zone   = "us-east5-c"
      region = "us-east5"
    }

    "google.dallas.1" = {
      zone   = "us-south1-a"
      region = "us-south1"
    }

    "google.dallas.2" = {
      zone   = "us-south1-b"
      region = "us-south1"
    }

    "google.dallas.3" = {
      zone   = "us-south1-c"
      region = "us-south1"
    }

    "google.oregon.1" = {
      zone   = "us-west1-a"
      region = "us-west1"
    }

    "google.oregon.2" = {
      zone   = "us-west1-b"
      region = "us-west1"
    }

    "google.oregon.3" = {
      zone   = "us-west1-c"
      region = "us-west1"
    }

    "google.losangeles.1" = {
      zone   = "us-west2-a"
      region = "us-west2"
    }

    "google.losangeles.2" = {
      zone   = "us-west2-b"
      region = "us-west2"
    }

    "google.losangeles.3" = {
      zone   = "us-west2-c"
      region = "us-west2"
    }

    "google.saltlakecity.1" = {
      zone   = "us-west3-a"
      region = "us-west3"
    }

    "google.saltlakecity.2" = {
      zone   = "us-west3-b"
      region = "us-west3"
    }

    "google.saltlakecity.3" = {
      zone   = "us-west3-c"
      region = "us-west3"
    }

    "google.lasvegas.1" = {
      zone   = "us-west4-a"
      region = "us-west4"
    }

    "google.lasvegas.2" = {
      zone   = "us-west4-b"
      region = "us-west4"
    }

    "google.lasvegas.3" = {
      zone   = "us-west4-c"
      region = "us-west4"
    }

  }

}
