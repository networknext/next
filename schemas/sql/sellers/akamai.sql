
-- akamai datacenters

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.southcarolina.2',
   'us-east1-b',
   33.836102,
   -81.163696,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.southcarolina.3',
   'us-east1-c',
   33.836102,
   -81.163696,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.southcarolina.4',
   'us-east1-d',
   33.836102,
   -81.163696,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.virginia.3',
   'us-east4-c',
   37.431599,
   -78.656898,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.virginia.2',
   'us-east4-b',
   37.431599,
   -78.656898,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.virginia.1',
   'us-east4-a',
   37.431599,
   -78.656898,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.iowa.3',
   'us-central1-c',
   41.877998,
   -93.097702,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.iowa.1',
   'us-central1-a',
   41.877998,
   -93.097702,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.iowa.6',
   'us-central1-f',
   41.877998,
   -93.097702,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.iowa.2',
   'us-central1-b',
   41.877998,
   -93.097702,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.oregon.2',
   'us-west1-b',
   45.594601,
   -121.178703,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.oregon.3',
   'us-west1-c',
   45.594601,
   -121.178703,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.oregon.1',
   'us-west1-a',
   45.594601,
   -121.178703,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.netherlands.1',
   'europe-west4-a',
   53.438599,
   6.835500,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.netherlands.2',
   'europe-west4-b',
   53.438599,
   6.835500,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.netherlands.3',
   'europe-west4-c',
   53.438599,
   6.835500,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.belgium.2',
   'europe-west1-b',
   50.470600,
   3.817000,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.belgium.4',
   'europe-west1-d',
   50.470600,
   3.817000,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.belgium.3',
   'europe-west1-c',
   50.470600,
   3.817000,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.frankfurt.3',
   'europe-west3-c',
   50.110901,
   8.682100,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.frankfurt.1',
   'europe-west3-a',
   50.110901,
   8.682100,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.frankfurt.2',
   'europe-west3-b',
   50.110901,
   8.682100,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.london.3',
   'europe-west2-c',
   51.507198,
   -0.127600,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.london.2',
   'europe-west2-b',
   51.507198,
   -0.127600,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.london.1',
   'europe-west2-a',
   51.507198,
   -0.127600,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.taiwan.2',
   'asia-east1-b',
   25.105497,
   121.597366,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.taiwan.1',
   'asia-east1-a',
   25.105497,
   121.597366,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.taiwan.3',
   'asia-east1-c',
   25.105497,
   121.597366,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.singapore.2',
   'asia-southeast1-b',
   1.352100,
   103.819801,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.singapore.1',
   'asia-southeast1-a',
   1.352100,
   103.819801,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.singapore.3',
   'asia-southeast1-c',
   1.352100,
   103.819801,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.tokyo.2',
   'asia-northeast1-b',
   35.676201,
   139.650299,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.tokyo.3',
   'asia-northeast1-c',
   35.676201,
   139.650299,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.tokyo.1',
   'asia-northeast1-a',
   35.676201,
   139.650299,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.mumbai.3',
   'asia-south1-c',
   19.076000,
   72.877701,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.mumbai.2',
   'asia-south1-b',
   19.076000,
   72.877701,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.mumbai.1',
   'asia-south1-a',
   19.076000,
   72.877701,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.sydney.2',
   'australia-southeast1-b',
   -33.868801,
   151.209305,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.sydney.3',
   'australia-southeast1-c',
   -33.868801,
   151.209305,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.sydney.1',
   'australia-southeast1-a',
   -33.868801,
   151.209305,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.saopaulo.2',
   'southamerica-east1-b',
   -23.555799,
   -46.639599,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.saopaulo.3',
   'southamerica-east1-c',
   -23.555799,
   -46.639599,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.saopaulo.1',
   'southamerica-east1-a',
   -23.555799,
   -46.639599,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.hongkong.1',
   'asia-east2-a',
   22.319300,
   114.169403,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.hongkong.2',
   'asia-east2-b',
   22.319300,
   114.169403,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.hongkong.3',
   'asia-east2-c',
   22.319300,
   114.169403,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.osaka.1',
   'asia-northeast2-a',
   34.693699,
   135.502304,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.osaka.2',
   'asia-northeast2-b',
   34.693699,
   135.502304,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.osaka.3',
   'asia-northeast2-c',
   34.693699,
   135.502304,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.seoul.1',
   'asia-northeast3-a',
   37.566502,
   126.977997,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.seoul.2',
   'asia-northeast3-b',
   37.566502,
   126.977997,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.seoul.3',
   'asia-northeast3-c',
   37.566502,
   126.977997,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.delhi.1',
   'asia-south2-a',
   28.704100,
   77.102501,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.delhi.2',
   'asia-south2-b',
   28.704100,
   77.102501,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.delhi.3',
   'asia-south2-c',
   28.704100,
   77.102501,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.jakarta.1',
   'asia-southeast2-a',
   6.208800,
   106.845596,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.jakarta.2',
   'asia-southeast2-b',
   6.208800,
   106.845596,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.jakarta.3',
   'asia-southeast2-c',
   6.208800,
   106.845596,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.melbourne.1',
   'australia-southeast2-a',
   -37.813599,
   144.963104,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.melbourne.2',
   'australia-southeast2-b',
   -37.813599,
   144.963104,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.melbourne.3',
   'australia-southeast2-c',
   -37.813599,
   144.963104,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.warsaw.1',
   'europe-central2-a',
   52.229698,
   21.012199,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.warsaw.2',
   'europe-central2-b',
   52.229698,
   21.012199,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.warsaw.3',
   'europe-central2-c',
   52.229698,
   21.012199,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.finland.1',
   'europe-north1-a',
   60.569302,
   27.187799,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.finland.2',
   'europe-north1-b',
   60.569302,
   27.187799,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.finland.3',
   'europe-north1-c',
   60.569302,
   27.187799,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.madrid.1',
   'europe-southwest1-a',
   40.416801,
   3.703800,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.madrid.2',
   'europe-southwest1-b',
   40.416801,
   3.703800,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.madrid.3',
   'europe-southwest1-c',
   40.416801,
   3.703800,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.zurich.1',
   'europe-west6-a',
   47.376900,
   8.541700,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.zurich.2',
   'europe-west6-b',
   47.376900,
   8.541700,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.zurich.3',
   'europe-west6-c',
   47.376900,
   8.541700,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.milan.1',
   'europe-west8-a',
   45.464199,
   9.190000,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.milan.2',
   'europe-west8-b',
   45.464199,
   9.190000,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.milan.3',
   'europe-west8-c',
   45.464199,
   9.190000,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.paris.1',
   'europe-west9-a',
   48.856602,
   2.352200,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.paris.2',
   'europe-west9-b',
   48.856602,
   2.352200,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.paris.3',
   'europe-west9-c',
   48.856602,
   2.352200,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.telaviv.1',
   'me-west1-a',
   32.085300,
   34.781799,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.telaviv.2',
   'me-west1-b',
   32.085300,
   34.781799,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.telaviv.3',
   'me-west1-c',
   32.085300,
   34.781799,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.montreal.1',
   'northamerica-northeast1-a',
   45.501900,
   -73.567398,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.montreal.2',
   'northamerica-northeast1-b',
   45.501900,
   -73.567398,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.montreal.3',
   'northamerica-northeast1-c',
   45.501900,
   -73.567398,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.toronto.1',
   'northamerica-northeast2-a',
   43.653198,
   -79.383202,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.toronto.2',
   'northamerica-northeast2-b',
   43.653198,
   -79.383202,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.toronto.3',
   'northamerica-northeast2-c',
   43.653198,
   -79.383202,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.santiago.1',
   'southamerica-west1-a',
   -33.448898,
   -70.669296,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.santiago.2',
   'southamerica-west1-b',
   -33.448898,
   -70.669296,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.santiago.3',
   'southamerica-west1-c',
   -33.448898,
   -70.669296,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.ohio.1',
   'us-east5-a',
   39.961201,
   -82.998802,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.ohio.2',
   'us-east5-b',
   39.961201,
   -82.998802,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.ohio.3',
   'us-east5-c',
   39.961201,
   -82.998802,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.dallas.1',
   'us-south1-a',
   32.776699,
   -96.796997,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.dallas.2',
   'us-south1-b',
   32.776699,
   -96.796997,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.dallas.3',
   'us-south1-c',
   32.776699,
   -96.796997,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.losangeles.1',
   'us-west2-a',
   34.052200,
   -118.243698,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.losangeles.2',
   'us-west2-b',
   34.052200,
   -118.243698,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.losangeles.3',
   'us-west2-c',
   34.052200,
   -118.243698,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.saltlakecity.1',
   'us-west3-a',
   40.760799,
   -111.890999,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.saltlakecity.2',
   'us-west3-b',
   40.760799,
   -111.890999,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.saltlakecity.3',
   'us-west3-c',
   40.760799,
   -111.890999,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.lasvegas.1',
   'us-west4-a',
   36.171600,
   -115.139099,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.lasvegas.2',
   'us-west4-b',
   36.171600,
   -115.139099,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.lasvegas.3',
   'us-west4-c',
   36.171600,
   -115.139099,
   (select seller_id from sellers where seller_name = 'akamai')
);
