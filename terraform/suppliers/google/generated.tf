
locals {

  datacenter_map = {

    "google.taiwan.1" = {
      zone        = "asia-east1-a"
      region      = "asia-east1"
      native_name = "asia-east1-a"
      longitude   = 25.11
      latitude    = 121.60
      seller_name = "google"
    }

    "google.taiwan.2" = {
      zone        = "asia-east1-b"
      region      = "asia-east1"
      native_name = "asia-east1-b"
      longitude   = 25.11
      latitude    = 121.60
      seller_name = "google"
    }

    "google.taiwan.3" = {
      zone        = "asia-east1-c"
      region      = "asia-east1"
      native_name = "asia-east1-c"
      longitude   = 25.11
      latitude    = 121.60
      seller_name = "google"
    }

    "google.hongkong.1" = {
      zone        = "asia-east2-a"
      region      = "asia-east2"
      native_name = "asia-east2-a"
      longitude   = 22.32
      latitude    = 114.17
      seller_name = "google"
    }

    "google.hongkong.2" = {
      zone        = "asia-east2-b"
      region      = "asia-east2"
      native_name = "asia-east2-b"
      longitude   = 22.32
      latitude    = 114.17
      seller_name = "google"
    }

    "google.hongkong.3" = {
      zone        = "asia-east2-c"
      region      = "asia-east2"
      native_name = "asia-east2-c"
      longitude   = 22.32
      latitude    = 114.17
      seller_name = "google"
    }

    "google.tokyo.1" = {
      zone        = "asia-northeast1-a"
      region      = "asia-northeast1"
      native_name = "asia-northeast1-a"
      longitude   = 35.68
      latitude    = 139.65
      seller_name = "google"
    }

    "google.tokyo.2" = {
      zone        = "asia-northeast1-b"
      region      = "asia-northeast1"
      native_name = "asia-northeast1-b"
      longitude   = 35.68
      latitude    = 139.65
      seller_name = "google"
    }

    "google.tokyo.3" = {
      zone        = "asia-northeast1-c"
      region      = "asia-northeast1"
      native_name = "asia-northeast1-c"
      longitude   = 35.68
      latitude    = 139.65
      seller_name = "google"
    }

    "google.osaka.1" = {
      zone        = "asia-northeast2-a"
      region      = "asia-northeast2"
      native_name = "asia-northeast2-a"
      longitude   = 34.69
      latitude    = 135.50
      seller_name = "google"
    }

    "google.osaka.2" = {
      zone        = "asia-northeast2-b"
      region      = "asia-northeast2"
      native_name = "asia-northeast2-b"
      longitude   = 34.69
      latitude    = 135.50
      seller_name = "google"
    }

    "google.osaka.3" = {
      zone        = "asia-northeast2-c"
      region      = "asia-northeast2"
      native_name = "asia-northeast2-c"
      longitude   = 34.69
      latitude    = 135.50
      seller_name = "google"
    }

    "google.seoul.1" = {
      zone        = "asia-northeast3-a"
      region      = "asia-northeast3"
      native_name = "asia-northeast3-a"
      longitude   = 37.57
      latitude    = 126.98
      seller_name = "google"
    }

    "google.seoul.2" = {
      zone        = "asia-northeast3-b"
      region      = "asia-northeast3"
      native_name = "asia-northeast3-b"
      longitude   = 37.57
      latitude    = 126.98
      seller_name = "google"
    }

    "google.seoul.3" = {
      zone        = "asia-northeast3-c"
      region      = "asia-northeast3"
      native_name = "asia-northeast3-c"
      longitude   = 37.57
      latitude    = 126.98
      seller_name = "google"
    }

    "google.mumbai.1" = {
      zone        = "asia-south1-a"
      region      = "asia-south1"
      native_name = "asia-south1-a"
      longitude   = 19.08
      latitude    = 72.88
      seller_name = "google"
    }

    "google.mumbai.2" = {
      zone        = "asia-south1-b"
      region      = "asia-south1"
      native_name = "asia-south1-b"
      longitude   = 19.08
      latitude    = 72.88
      seller_name = "google"
    }

    "google.mumbai.3" = {
      zone        = "asia-south1-c"
      region      = "asia-south1"
      native_name = "asia-south1-c"
      longitude   = 19.08
      latitude    = 72.88
      seller_name = "google"
    }

    "google.delhi.1" = {
      zone        = "asia-south2-a"
      region      = "asia-south2"
      native_name = "asia-south2-a"
      longitude   = 28.70
      latitude    = 77.10
      seller_name = "google"
    }

    "google.delhi.2" = {
      zone        = "asia-south2-b"
      region      = "asia-south2"
      native_name = "asia-south2-b"
      longitude   = 28.70
      latitude    = 77.10
      seller_name = "google"
    }

    "google.delhi.3" = {
      zone        = "asia-south2-c"
      region      = "asia-south2"
      native_name = "asia-south2-c"
      longitude   = 28.70
      latitude    = 77.10
      seller_name = "google"
    }

    "google.singapore.1" = {
      zone        = "asia-southeast1-a"
      region      = "asia-southeast1"
      native_name = "asia-southeast1-a"
      longitude   = 1.35
      latitude    = 103.82
      seller_name = "google"
    }

    "google.singapore.2" = {
      zone        = "asia-southeast1-b"
      region      = "asia-southeast1"
      native_name = "asia-southeast1-b"
      longitude   = 1.35
      latitude    = 103.82
      seller_name = "google"
    }

    "google.singapore.3" = {
      zone        = "asia-southeast1-c"
      region      = "asia-southeast1"
      native_name = "asia-southeast1-c"
      longitude   = 1.35
      latitude    = 103.82
      seller_name = "google"
    }

    "google.jakarta.1" = {
      zone        = "asia-southeast2-a"
      region      = "asia-southeast2"
      native_name = "asia-southeast2-a"
      longitude   = 6.21
      latitude    = 106.85
      seller_name = "google"
    }

    "google.jakarta.2" = {
      zone        = "asia-southeast2-b"
      region      = "asia-southeast2"
      native_name = "asia-southeast2-b"
      longitude   = 6.21
      latitude    = 106.85
      seller_name = "google"
    }

    "google.jakarta.3" = {
      zone        = "asia-southeast2-c"
      region      = "asia-southeast2"
      native_name = "asia-southeast2-c"
      longitude   = 6.21
      latitude    = 106.85
      seller_name = "google"
    }

    "google.sydney.1" = {
      zone        = "australia-southeast1-a"
      region      = "australia-southeast1"
      native_name = "australia-southeast1-a"
      longitude   = -33.87
      latitude    = 151.21
      seller_name = "google"
    }

    "google.sydney.2" = {
      zone        = "australia-southeast1-b"
      region      = "australia-southeast1"
      native_name = "australia-southeast1-b"
      longitude   = -33.87
      latitude    = 151.21
      seller_name = "google"
    }

    "google.sydney.3" = {
      zone        = "australia-southeast1-c"
      region      = "australia-southeast1"
      native_name = "australia-southeast1-c"
      longitude   = -33.87
      latitude    = 151.21
      seller_name = "google"
    }

    "google.melbourne.1" = {
      zone        = "australia-southeast2-a"
      region      = "australia-southeast2"
      native_name = "australia-southeast2-a"
      longitude   = -37.81
      latitude    = 144.96
      seller_name = "google"
    }

    "google.melbourne.2" = {
      zone        = "australia-southeast2-b"
      region      = "australia-southeast2"
      native_name = "australia-southeast2-b"
      longitude   = -37.81
      latitude    = 144.96
      seller_name = "google"
    }

    "google.melbourne.3" = {
      zone        = "australia-southeast2-c"
      region      = "australia-southeast2"
      native_name = "australia-southeast2-c"
      longitude   = -37.81
      latitude    = 144.96
      seller_name = "google"
    }

    "google.warsaw.1" = {
      zone        = "europe-central2-a"
      region      = "europe-central2"
      native_name = "europe-central2-a"
      longitude   = 52.23
      latitude    = 21.01
      seller_name = "google"
    }

    "google.warsaw.2" = {
      zone        = "europe-central2-b"
      region      = "europe-central2"
      native_name = "europe-central2-b"
      longitude   = 52.23
      latitude    = 21.01
      seller_name = "google"
    }

    "google.warsaw.3" = {
      zone        = "europe-central2-c"
      region      = "europe-central2"
      native_name = "europe-central2-c"
      longitude   = 52.23
      latitude    = 21.01
      seller_name = "google"
    }

    "google.finland.1" = {
      zone        = "europe-north1-a"
      region      = "europe-north1"
      native_name = "europe-north1-a"
      longitude   = 60.57
      latitude    = 27.19
      seller_name = "google"
    }

    "google.finland.2" = {
      zone        = "europe-north1-b"
      region      = "europe-north1"
      native_name = "europe-north1-b"
      longitude   = 60.57
      latitude    = 27.19
      seller_name = "google"
    }

    "google.finland.3" = {
      zone        = "europe-north1-c"
      region      = "europe-north1"
      native_name = "europe-north1-c"
      longitude   = 60.57
      latitude    = 27.19
      seller_name = "google"
    }

    "google.madrid.1" = {
      zone        = "europe-southwest1-a"
      region      = "europe-southwest1"
      native_name = "europe-southwest1-a"
      longitude   = 40.42
      latitude    = 3.70
      seller_name = "google"
    }

    "google.madrid.2" = {
      zone        = "europe-southwest1-b"
      region      = "europe-southwest1"
      native_name = "europe-southwest1-b"
      longitude   = 40.42
      latitude    = 3.70
      seller_name = "google"
    }

    "google.madrid.3" = {
      zone        = "europe-southwest1-c"
      region      = "europe-southwest1"
      native_name = "europe-southwest1-c"
      longitude   = 40.42
      latitude    = 3.70
      seller_name = "google"
    }

    "google.belgium.2" = {
      zone        = "europe-west1-b"
      region      = "europe-west1"
      native_name = "europe-west1-b"
      longitude   = 50.47
      latitude    = 3.82
      seller_name = "google"
    }

    "google.belgium.3" = {
      zone        = "europe-west1-c"
      region      = "europe-west1"
      native_name = "europe-west1-c"
      longitude   = 50.47
      latitude    = 3.82
      seller_name = "google"
    }

    "google.belgium.4" = {
      zone        = "europe-west1-d"
      region      = "europe-west1"
      native_name = "europe-west1-d"
      longitude   = 50.47
      latitude    = 3.82
      seller_name = "google"
    }

    "google.london.1" = {
      zone        = "europe-west2-a"
      region      = "europe-west2"
      native_name = "europe-west2-a"
      longitude   = 51.51
      latitude    = -0.13
      seller_name = "google"
    }

    "google.london.2" = {
      zone        = "europe-west2-b"
      region      = "europe-west2"
      native_name = "europe-west2-b"
      longitude   = 51.51
      latitude    = -0.13
      seller_name = "google"
    }

    "google.london.3" = {
      zone        = "europe-west2-c"
      region      = "europe-west2"
      native_name = "europe-west2-c"
      longitude   = 51.51
      latitude    = -0.13
      seller_name = "google"
    }

    "google.frankfurt.1" = {
      zone        = "europe-west3-a"
      region      = "europe-west3"
      native_name = "europe-west3-a"
      longitude   = 50.11
      latitude    = 8.68
      seller_name = "google"
    }

    "google.frankfurt.2" = {
      zone        = "europe-west3-b"
      region      = "europe-west3"
      native_name = "europe-west3-b"
      longitude   = 50.11
      latitude    = 8.68
      seller_name = "google"
    }

    "google.frankfurt.3" = {
      zone        = "europe-west3-c"
      region      = "europe-west3"
      native_name = "europe-west3-c"
      longitude   = 50.11
      latitude    = 8.68
      seller_name = "google"
    }

    "google.netherlands.1" = {
      zone        = "europe-west4-a"
      region      = "europe-west4"
      native_name = "europe-west4-a"
      longitude   = 53.44
      latitude    = 6.84
      seller_name = "google"
    }

    "google.netherlands.2" = {
      zone        = "europe-west4-b"
      region      = "europe-west4"
      native_name = "europe-west4-b"
      longitude   = 53.44
      latitude    = 6.84
      seller_name = "google"
    }

    "google.netherlands.3" = {
      zone        = "europe-west4-c"
      region      = "europe-west4"
      native_name = "europe-west4-c"
      longitude   = 53.44
      latitude    = 6.84
      seller_name = "google"
    }

    "google.zurich.1" = {
      zone        = "europe-west6-a"
      region      = "europe-west6"
      native_name = "europe-west6-a"
      longitude   = 47.38
      latitude    = 8.54
      seller_name = "google"
    }

    "google.zurich.2" = {
      zone        = "europe-west6-b"
      region      = "europe-west6"
      native_name = "europe-west6-b"
      longitude   = 47.38
      latitude    = 8.54
      seller_name = "google"
    }

    "google.zurich.3" = {
      zone        = "europe-west6-c"
      region      = "europe-west6"
      native_name = "europe-west6-c"
      longitude   = 47.38
      latitude    = 8.54
      seller_name = "google"
    }

    "google.milan.1" = {
      zone        = "europe-west8-a"
      region      = "europe-west8"
      native_name = "europe-west8-a"
      longitude   = 45.46
      latitude    = 9.19
      seller_name = "google"
    }

    "google.milan.2" = {
      zone        = "europe-west8-b"
      region      = "europe-west8"
      native_name = "europe-west8-b"
      longitude   = 45.46
      latitude    = 9.19
      seller_name = "google"
    }

    "google.milan.3" = {
      zone        = "europe-west8-c"
      region      = "europe-west8"
      native_name = "europe-west8-c"
      longitude   = 45.46
      latitude    = 9.19
      seller_name = "google"
    }

    "google.paris.1" = {
      zone        = "europe-west9-a"
      region      = "europe-west9"
      native_name = "europe-west9-a"
      longitude   = 48.86
      latitude    = 2.35
      seller_name = "google"
    }

    "google.paris.2" = {
      zone        = "europe-west9-b"
      region      = "europe-west9"
      native_name = "europe-west9-b"
      longitude   = 48.86
      latitude    = 2.35
      seller_name = "google"
    }

    "google.paris.3" = {
      zone        = "europe-west9-c"
      region      = "europe-west9"
      native_name = "europe-west9-c"
      longitude   = 48.86
      latitude    = 2.35
      seller_name = "google"
    }

    "google.telaviv.1" = {
      zone        = "me-west1-a"
      region      = "me-west1"
      native_name = "me-west1-a"
      longitude   = 32.09
      latitude    = 34.78
      seller_name = "google"
    }

    "google.telaviv.2" = {
      zone        = "me-west1-b"
      region      = "me-west1"
      native_name = "me-west1-b"
      longitude   = 32.09
      latitude    = 34.78
      seller_name = "google"
    }

    "google.telaviv.3" = {
      zone        = "me-west1-c"
      region      = "me-west1"
      native_name = "me-west1-c"
      longitude   = 32.09
      latitude    = 34.78
      seller_name = "google"
    }

    "google.montreal.1" = {
      zone        = "northamerica-northeast1-a"
      region      = "northamerica-northeast1"
      native_name = "northamerica-northeast1-a"
      longitude   = 45.50
      latitude    = -73.57
      seller_name = "google"
    }

    "google.montreal.2" = {
      zone        = "northamerica-northeast1-b"
      region      = "northamerica-northeast1"
      native_name = "northamerica-northeast1-b"
      longitude   = 45.50
      latitude    = -73.57
      seller_name = "google"
    }

    "google.montreal.3" = {
      zone        = "northamerica-northeast1-c"
      region      = "northamerica-northeast1"
      native_name = "northamerica-northeast1-c"
      longitude   = 45.50
      latitude    = -73.57
      seller_name = "google"
    }

    "google.toronto.1" = {
      zone        = "northamerica-northeast2-a"
      region      = "northamerica-northeast2"
      native_name = "northamerica-northeast2-a"
      longitude   = 43.65
      latitude    = -79.38
      seller_name = "google"
    }

    "google.toronto.2" = {
      zone        = "northamerica-northeast2-b"
      region      = "northamerica-northeast2"
      native_name = "northamerica-northeast2-b"
      longitude   = 43.65
      latitude    = -79.38
      seller_name = "google"
    }

    "google.toronto.3" = {
      zone        = "northamerica-northeast2-c"
      region      = "northamerica-northeast2"
      native_name = "northamerica-northeast2-c"
      longitude   = 43.65
      latitude    = -79.38
      seller_name = "google"
    }

    "google.saopaulo.1" = {
      zone        = "southamerica-east1-a"
      region      = "southamerica-east1"
      native_name = "southamerica-east1-a"
      longitude   = -23.56
      latitude    = -46.64
      seller_name = "google"
    }

    "google.saopaulo.2" = {
      zone        = "southamerica-east1-b"
      region      = "southamerica-east1"
      native_name = "southamerica-east1-b"
      longitude   = -23.56
      latitude    = -46.64
      seller_name = "google"
    }

    "google.saopaulo.3" = {
      zone        = "southamerica-east1-c"
      region      = "southamerica-east1"
      native_name = "southamerica-east1-c"
      longitude   = -23.56
      latitude    = -46.64
      seller_name = "google"
    }

    "google.santiago.1" = {
      zone        = "southamerica-west1-a"
      region      = "southamerica-west1"
      native_name = "southamerica-west1-a"
      longitude   = -33.45
      latitude    = -70.67
      seller_name = "google"
    }

    "google.santiago.2" = {
      zone        = "southamerica-west1-b"
      region      = "southamerica-west1"
      native_name = "southamerica-west1-b"
      longitude   = -33.45
      latitude    = -70.67
      seller_name = "google"
    }

    "google.santiago.3" = {
      zone        = "southamerica-west1-c"
      region      = "southamerica-west1"
      native_name = "southamerica-west1-c"
      longitude   = -33.45
      latitude    = -70.67
      seller_name = "google"
    }

    "google.iowa.1" = {
      zone        = "us-central1-a"
      region      = "us-central1"
      native_name = "us-central1-a"
      longitude   = 41.88
      latitude    = -93.10
      seller_name = "google"
    }

    "google.iowa.2" = {
      zone        = "us-central1-b"
      region      = "us-central1"
      native_name = "us-central1-b"
      longitude   = 41.88
      latitude    = -93.10
      seller_name = "google"
    }

    "google.iowa.3" = {
      zone        = "us-central1-c"
      region      = "us-central1"
      native_name = "us-central1-c"
      longitude   = 41.88
      latitude    = -93.10
      seller_name = "google"
    }

    "google.iowa.6" = {
      zone        = "us-central1-f"
      region      = "us-central1"
      native_name = "us-central1-f"
      longitude   = 41.88
      latitude    = -93.10
      seller_name = "google"
    }

    "google.southcarolina.2" = {
      zone        = "us-east1-b"
      region      = "us-east1"
      native_name = "us-east1-b"
      longitude   = 33.84
      latitude    = -81.16
      seller_name = "google"
    }

    "google.southcarolina.3" = {
      zone        = "us-east1-c"
      region      = "us-east1"
      native_name = "us-east1-c"
      longitude   = 33.84
      latitude    = -81.16
      seller_name = "google"
    }

    "google.southcarolina.4" = {
      zone        = "us-east1-d"
      region      = "us-east1"
      native_name = "us-east1-d"
      longitude   = 33.84
      latitude    = -81.16
      seller_name = "google"
    }

    "google.virginia.1" = {
      zone        = "us-east4-a"
      region      = "us-east4"
      native_name = "us-east4-a"
      longitude   = 37.43
      latitude    = -78.66
      seller_name = "google"
    }

    "google.virginia.2" = {
      zone        = "us-east4-b"
      region      = "us-east4"
      native_name = "us-east4-b"
      longitude   = 37.43
      latitude    = -78.66
      seller_name = "google"
    }

    "google.virginia.3" = {
      zone        = "us-east4-c"
      region      = "us-east4"
      native_name = "us-east4-c"
      longitude   = 37.43
      latitude    = -78.66
      seller_name = "google"
    }

    "google.ohio.1" = {
      zone        = "us-east5-a"
      region      = "us-east5"
      native_name = "us-east5-a"
      longitude   = 39.96
      latitude    = -83.00
      seller_name = "google"
    }

    "google.ohio.2" = {
      zone        = "us-east5-b"
      region      = "us-east5"
      native_name = "us-east5-b"
      longitude   = 39.96
      latitude    = -83.00
      seller_name = "google"
    }

    "google.ohio.3" = {
      zone        = "us-east5-c"
      region      = "us-east5"
      native_name = "us-east5-c"
      longitude   = 39.96
      latitude    = -83.00
      seller_name = "google"
    }

    "google.dallas.1" = {
      zone        = "us-south1-a"
      region      = "us-south1"
      native_name = "us-south1-a"
      longitude   = 32.78
      latitude    = -96.80
      seller_name = "google"
    }

    "google.dallas.2" = {
      zone        = "us-south1-b"
      region      = "us-south1"
      native_name = "us-south1-b"
      longitude   = 32.78
      latitude    = -96.80
      seller_name = "google"
    }

    "google.dallas.3" = {
      zone        = "us-south1-c"
      region      = "us-south1"
      native_name = "us-south1-c"
      longitude   = 32.78
      latitude    = -96.80
      seller_name = "google"
    }

    "google.oregon.1" = {
      zone        = "us-west1-a"
      region      = "us-west1"
      native_name = "us-west1-a"
      longitude   = 45.59
      latitude    = -121.18
      seller_name = "google"
    }

    "google.oregon.2" = {
      zone        = "us-west1-b"
      region      = "us-west1"
      native_name = "us-west1-b"
      longitude   = 45.59
      latitude    = -121.18
      seller_name = "google"
    }

    "google.oregon.3" = {
      zone        = "us-west1-c"
      region      = "us-west1"
      native_name = "us-west1-c"
      longitude   = 45.59
      latitude    = -121.18
      seller_name = "google"
    }

    "google.losangeles.1" = {
      zone        = "us-west2-a"
      region      = "us-west2"
      native_name = "us-west2-a"
      longitude   = 34.05
      latitude    = -118.24
      seller_name = "google"
    }

    "google.losangeles.2" = {
      zone        = "us-west2-b"
      region      = "us-west2"
      native_name = "us-west2-b"
      longitude   = 34.05
      latitude    = -118.24
      seller_name = "google"
    }

    "google.losangeles.3" = {
      zone        = "us-west2-c"
      region      = "us-west2"
      native_name = "us-west2-c"
      longitude   = 34.05
      latitude    = -118.24
      seller_name = "google"
    }

    "google.saltlakecity.1" = {
      zone        = "us-west3-a"
      region      = "us-west3"
      native_name = "us-west3-a"
      longitude   = 40.76
      latitude    = -111.89
      seller_name = "google"
    }

    "google.saltlakecity.2" = {
      zone        = "us-west3-b"
      region      = "us-west3"
      native_name = "us-west3-b"
      longitude   = 40.76
      latitude    = -111.89
      seller_name = "google"
    }

    "google.saltlakecity.3" = {
      zone        = "us-west3-c"
      region      = "us-west3"
      native_name = "us-west3-c"
      longitude   = 40.76
      latitude    = -111.89
      seller_name = "google"
    }

    "google.lasvegas.1" = {
      zone        = "us-west4-a"
      region      = "us-west4"
      native_name = "us-west4-a"
      longitude   = 36.17
      latitude    = -115.14
      seller_name = "google"
    }

    "google.lasvegas.2" = {
      zone        = "us-west4-b"
      region      = "us-west4"
      native_name = "us-west4-b"
      longitude   = 36.17
      latitude    = -115.14
      seller_name = "google"
    }

    "google.lasvegas.3" = {
      zone        = "us-west4-c"
      region      = "us-west4"
      native_name = "us-west4-c"
      longitude   = 36.17
      latitude    = -115.14
      seller_name = "google"
    }

  }

}
