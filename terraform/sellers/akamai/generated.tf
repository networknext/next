
locals {

  datacenter_map = {

    "akamai.amsterdam" = {
      zone        = "nl-ams"
      native_name = "nl-ams"
      latitude    = 52.37
      longitude   = 4.90
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.tokyo.2" = {
      zone        = "jp-tyo-3"
      native_name = "jp-tyo-3"
      latitude    = 35.68
      longitude   = 139.65
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.singapore.2" = {
      zone        = "sg-sin-2"
      native_name = "sg-sin-2"
      latitude    = 1.35
      longitude   = 103.82
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.frankfurt.2" = {
      zone        = "de-fra-2"
      native_name = "de-fra-2"
      latitude    = 50.11
      longitude   = 8.68
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.mumbai.2" = {
      zone        = "in-bom-2"
      native_name = "in-bom-2"
      latitude    = 19.08
      longitude   = 72.88
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.melbourne" = {
      zone        = "au-mel"
      native_name = "au-mel"
      latitude    = -37.81
      longitude   = 144.96
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.london" = {
      zone        = "gb-lon"
      native_name = "gb-lon"
      latitude    = 51.51
      longitude   = -0.13
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.losangeles" = {
      zone        = "us-lax"
      native_name = "us-lax"
      latitude    = 34.05
      longitude   = -118.24
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.jakarta" = {
      zone        = "id-cgk"
      native_name = "id-cgk"
      latitude    = -6.19
      longitude   = 106.82
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.miami" = {
      zone        = "us-mia"
      native_name = "us-mia"
      latitude    = 25.76
      longitude   = -80.19
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.milan" = {
      zone        = "it-mil"
      native_name = "it-mil"
      latitude    = 45.47
      longitude   = 9.18
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.osaka" = {
      zone        = "jp-osa"
      native_name = "jp-osa"
      latitude    = 34.69
      longitude   = 135.50
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.chennai" = {
      zone        = "in-maa"
      native_name = "in-maa"
      latitude    = 13.08
      longitude   = 80.27
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.madrid" = {
      zone        = "es-mad"
      native_name = "es-mad"
      latitude    = 40.42
      longitude   = -3.70
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.stockholm" = {
      zone        = "se-sto"
      native_name = "se-sto"
      latitude    = 59.33
      longitude   = 18.07
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.saopaulo" = {
      zone        = "br-gru"
      native_name = "br-gru"
      latitude    = -23.56
      longitude   = -46.64
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.seattle" = {
      zone        = "us-sea"
      native_name = "us-sea"
      latitude    = 47.61
      longitude   = -122.33
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.paris" = {
      zone        = "fr-par"
      native_name = "fr-par"
      latitude    = 48.86
      longitude   = 2.35
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.chicago" = {
      zone        = "us-ord"
      native_name = "us-ord"
      latitude    = 41.88
      longitude   = -87.63
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.washington" = {
      zone        = "us-iad"
      native_name = "us-iad"
      latitude    = 47.75
      longitude   = -120.74
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.sydney" = {
      zone        = "ap-southeast"
      native_name = "ap-southeast"
      latitude    = -33.87
      longitude   = 151.21
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.toronto" = {
      zone        = "ca-central"
      native_name = "ca-central"
      latitude    = 43.65
      longitude   = -79.38
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.mumbai.1" = {
      zone        = "ap-west"
      native_name = "ap-west"
      latitude    = 19.08
      longitude   = 72.88
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.dallas" = {
      zone        = "us-central"
      native_name = "us-central"
      latitude    = 32.78
      longitude   = -96.80
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.tokyo.1" = {
      zone        = "ap-northeast"
      native_name = "ap-northeast"
      latitude    = 35.68
      longitude   = 139.65
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.frankfurt.1" = {
      zone        = "eu-central"
      native_name = "eu-central"
      latitude    = 50.11
      longitude   = 8.68
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.singapore.1" = {
      zone        = "ap-south"
      native_name = "ap-south"
      latitude    = 1.35
      longitude   = 103.82
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.london" = {
      zone        = "eu-west"
      native_name = "eu-west"
      latitude    = 51.51
      longitude   = -0.13
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.newyork" = {
      zone        = "us-east"
      native_name = "us-east"
      latitude    = 40.71
      longitude   = -74.01
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.atlanta" = {
      zone        = "us-southeast"
      native_name = "us-southeast"
      latitude    = 33.75
      longitude   = -84.39
      seller_name = "Akamai"
      seller_code = "akamai"
    }

    "akamai.fremont" = {
      zone        = "us-west"
      native_name = "us-west"
      latitude    = 37.55
      longitude   = -121.99
      seller_name = "Akamai"
      seller_code = "akamai"
    }

  }

}
