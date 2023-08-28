
locals {

  datacenter_map = {

    "vultr.amsterdam" = {
      zone        = "ams"
      native_name = "ams"
      latitude    = 2.37
      longitude   = 4.90
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.atlanta" = {
      zone        = "atl"
      native_name = "atl"
      latitude    = 33.75
      longitude   = -84.39
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.bangalore" = {
      zone        = "blr"
      native_name = "blr"
      latitude    = 12.97
      longitude   = 77.59
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.mumbai" = {
      zone        = "bom"
      native_name = "bom"
      latitude    = 19.08
      longitude   = 72.88
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.paris" = {
      zone        = "cdg"
      native_name = "cdg"
      latitude    = 48.86
      longitude   = 2.35
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.delhi" = {
      zone        = "del"
      native_name = "del"
      latitude    = 28.70
      longitude   = 77.10
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.dallas" = {
      zone        = "dfw"
      native_name = "dfw"
      latitude    = 32.78
      longitude   = -96.80
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.newyork" = {
      zone        = "ewr"
      native_name = "ewr"
      latitude    = 40.71
      longitude   = -74.01
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.frankfurt" = {
      zone        = "fra"
      native_name = "fra"
      latitude    = 50.11
      longitude   = 8.68
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.honolulu" = {
      zone        = "hnl"
      native_name = "hnl"
      latitude    = 21.31
      longitude   = -157.86
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.seoul" = {
      zone        = "icn"
      native_name = "icn"
      latitude    = 37.57
      longitude   = 126.98
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.osaka" = {
      zone        = "itm"
      native_name = "itm"
      latitude    = 34.69
      longitude   = 135.50
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.johannesburg" = {
      zone        = "jnb"
      native_name = "jnb"
      latitude    = -26.20
      longitude   = 28.05
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.losangeles" = {
      zone        = "lax"
      native_name = "lax"
      latitude    = 34.05
      longitude   = -118.24
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.london" = {
      zone        = "lhr"
      native_name = "lhr"
      latitude    = 51.51
      longitude   = -0.13
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.madrid" = {
      zone        = "mad"
      native_name = "mad"
      latitude    = 40.42
      longitude   = -3.70
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.melbourne" = {
      zone        = "mel"
      native_name = "mel"
      latitude    = -37.81
      longitude   = 144.96
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.mexico" = {
      zone        = "mex"
      native_name = "mex"
      latitude    = 19.43
      longitude   = -99.13
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.miami" = {
      zone        = "mia"
      native_name = "mia"
      latitude    = 25.76
      longitude   = -80.19
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.tokyo" = {
      zone        = "nrt"
      native_name = "nrt"
      latitude    = 35.68
      longitude   = 139.65
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.chicago" = {
      zone        = "ord"
      native_name = "ord"
      latitude    = 41.88
      longitude   = -87.63
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.saopaulo" = {
      zone        = "sao"
      native_name = "sao"
      latitude    = -23.56
      longitude   = -46.64
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.santiago" = {
      zone        = "scl"
      native_name = "scl"
      latitude    = -33.45
      longitude   = -70.67
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.seattle" = {
      zone        = "sea"
      native_name = "sea"
      latitude    = 47.61
      longitude   = -122.33
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.singapore" = {
      zone        = "sgp"
      native_name = "sgp"
      latitude    = 1.35
      longitude   = 103.82
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.siliconvalley" = {
      zone        = "sjc"
      native_name = "sjc"
      latitude    = 37.34
      longitude   = -121.89
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.stockholm" = {
      zone        = "sto"
      native_name = "sto"
      latitude    = 59.33
      longitude   = 18.07
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.sydney" = {
      zone        = "syd"
      native_name = "syd"
      latitude    = -33.87
      longitude   = 151.21
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.warsaw" = {
      zone        = "waw"
      native_name = "waw"
      latitude    = 52.23
      longitude   = 21.01
      seller_name = "VULTR"
      seller_code = "vultr"
    }

    "vultr.toronto" = {
      zone        = "yto"
      native_name = "yto"
      latitude    = 43.65
      longitude   = -79.38
      seller_name = "VULTR"
      seller_code = "vultr"
    }

  }

}
