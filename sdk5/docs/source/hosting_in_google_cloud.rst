
Hosting in Google Cloud
=======================

Network Next can automatically detect which google datacenter your server is running in.

To use this autodetection, simply pass in a datacenter of "" to *next_server_create* when you run on google cloud, while making sure that NEXT_DATACENTER environment variable is also not set, as it overrides whatever datacenter string you pass in.

The mapping from google cloud zone to datacenter string is as follows:

========================== =========================================
Google Cloud Zone          Network Next Datacenter
-------------------------- -----------------------------------------
northamerica-northeast1-a  google.montreal.1
northamerica-northeast1-b  google.montreal.2
northamerica-northeast1-c  google.montreal.3
southamerica-east1-a       google.saopaulo.1
southamerica-east1-b       google.saopaulo.2
southamerica-east1-c       google.saopaulo.3
us-central1-a              google.iowa.1
us-central1-b              google.iowa.2
us-central1-c              google.iowa.3
us-central1-f              google.iowa.6
us-east1-b                 google.southcarolina.2
us-east1-c                 google.southcarolina.3
us-east1-d                 google.southcarolina.4
us-east4-a                 google.nothernvirginia.1
us-east4-b                 google.nothernvirginia.2
us-east4-c                 google.nothernvirginia.3
us-west1-a                 google.oregon.1
us-west1-b                 google.oregon.2
us-west1-c                 google.oregon.3
us-west2-a                 google.losangeles.1
us-west2-b                 google.losangeles.2
us-west2-c                 google.losangeles.3
us-west3-a                 google.saltlakecity.1
us-west3-b                 google.saltlakecity.2
us-west3-c                 google.saltlakecity.3
us-west4-a                 google.lasvegas.1
us-west4-b                 google.lasvegas.2
us-west4-c                 google.lasvegas.3
europe-north1-a            google.finland.1
europe-north1-b            google.finland.2
europe-north1-c            google.finland.3
europe-west1-b             google.belgium.2
europe-west1-c             google.belgium.3
europe-west1-d             google.belgium.4
europe-west2-a             google.london.1
europe-west2-b             google.london.2
europe-west2-c             google.london.3
europe-west3-a             google.frankfurt.1
europe-west3-b             google.frankfurt.2
europe-west3-c             google.frankfurt.3
europe-west4-a             google.netherlands.1
europe-west4-b             google.netherlands.2
europe-west4-c             google.netherlands.3
europe-west6-a             google.zurich.1
europe-west6-b             google.zurich.2
europe-west6-c             google.zurich.3
asia-east1-a               google.taiwan.1
asia-east1-b               google.taiwan.2
asia-east1-c               google.taiwan.3
asia-east2-a               google.hongkong.1
asia-east2-b               google.hongkong.2
asia-east2-c               google.hongkong.3
asia-northeast1-a          google.tokyo.1
asia-northeast1-b          google.tokyo.2
asia-northeast1-c          google.tokyo.3
asia-northeast2-a          google.osaka.1
asia-northeast2-b          google.osaka.2
asia-northeast2-c          google.osaka.3
asia-northeast3-a          google.seoul.1
asia-northeast3-b          google.seoul.2
asia-northeast3-c          google.seoul.3
asia-south1-a              google.mumbai.1
asia-south1-b              google.mumbai.2
asia-south1-c              google.mumbai.3
asia-southeast1-a          google.singapore.1
asia-southeast1-b          google.singapore.2
asia-southeast1-c          google.singapore.3
asia-southeast2-a          google.jakarta.1
asia-southeast2-b          google.jakarta.2
asia-southeast2-c          google.jakarta.3
australia-southeast1-a     google.sydney.1
australia-southeast1-b     google.sydney.2
australia-southeast1-c     google.sydney.3
========================== =========================================

You can always get the most up to date version of this mapping from this URL:

https://storage.googleapis.com/network-next-sdk/google.txt
