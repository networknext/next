
-- google datacenters

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.taiwan.1',
	'asia-east1-a',
	25.105497,
	121.597366,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.taiwan.2',
	'asia-east1-b',
	25.105497,
	121.597366,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.taiwan.3',
	'asia-east1-c',
	25.105497,
	121.597366,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.hongkong.1',
	'asia-east2-a',
	22.3193,
	114.1694,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.hongkong.2',
	'asia-east2-b',
	22.3193,
	114.1694,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.hongkong.3',
	'asia-east2-c',
	22.3193,
	114.1694,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.tokyo.1',
	'asia-northeast1-a',
	35.6762,
	139.6503,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.tokyo.2',
	'asia-northeast1-b',
	35.6762,
	139.6503,
a	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.tokyo.3',
	'asia-northeast1-c',
	35.6762,
	139.6503,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.osaka.1',
	'asia-northeast2-a',
	34.6937,
	135.5023,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.osaka.2',
	'asia-northeast2-b',
	34.6937,
	135.5023,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.osaka.3',
	'asia-northeast2-c',
	34.6937,
	135.5023,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.seoul.1',
	'asia-northeast3-a',
	37.5665,
	126.9780,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.seoul.2',
	'asia-northeast3-b',
	37.5665,
	126.9780,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.seoul.3',
	'asia-northeast3-c',
	37.5665,
	126.9780,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.mumbai.1',
	'asia-south1-a',
	19.0760,
	72.8777,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.mumbai.2',
	'asia-south1-b',
	19.0760,
	72.8777,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.mumbai.3',
	'asia-south1-c',
	19.0760,
	72.8777,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.delhi.1',
	'asia-south2-a',
	28.7041,
	77.1025,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.delhi.2',
	'asia-south2-b',
	28.7041,
	77.1025,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.delhi.3',
	'asia-south2-c',
	28.7041,
	77.1025,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.singapore.1',
	'asia-southeast1-a',
	1.3521,
	103.8198,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.singapore.2',
	'asia-southeast1-b',
	1.3521,
	103.8198,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.singapore.3',
	'asia-southeast1-c',
	1.3521,
	103.8198,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.jakarta.1',
	'asia-southeast2-a',
	6.2088,
	106.8456,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.jakarta.2',
	'asia-southeast2-b',
	6.2088,
	106.8456,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.jakarta.3',
	'asia-southeast2-c',
	6.2088,
	106.8456,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.sydney.1',
	'australia-southeast1-a',
	-33.8688,
	151.2093,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.sydney.2',
	'australia-southeast1-b',
	-33.8688,
	151.2093,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.sydney.3',
	'australia-southeast1-c',
	-33.8688,
	151.2093,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.melbourne.1',
	'australia-southeast2-a',
	-37.8136,
	144.9631,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.melbourne.2',
	'australia-southeast2-b',
	-37.8136,
	144.9631,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.melbourne.3',
	'australia-southeast2-c',
	-37.8136,
	144.9631,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.warsaw.1',
	'europe-central2-a',
	52.2297,
	21.0122,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.warsaw.2',
	'europe-central2-b',
	52.2297,
	21.0122,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.warsaw.3',
	'europe-central2-c',
	52.2297,
	21.0122,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.finland.1',
	'europe-north1-a',
	60.5693,
	27.1878,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.finland.2',
	'europe-north1-b',
	60.5693,
	27.1878,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.finland.3',
	'europe-north1-c',
	60.5693,
	27.1878,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.madrid.1',
	'europe-southwest1-a',
	40.4168,
	3.7038,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.madrid.2',
	'europe-southwest1-b',
	40.4168,
	3.7038,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.madrid.3',
	'europe-southwest1-c',
	40.4168,
	3.7038,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.belgium.1',
	'europe-west1-b',
	50.4706,
	3.8170,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.belgium.2',
	'europe-west1-c',
	50.4706,
	3.8170,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.belgium.3',
	'europe-west1-d',
	50.4706,
	3.8170,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.london.1',
	'europe-west2-a',
	51.5072,
	-0.1276,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.london.2',
	'europe-west2-b',
	51.5072,
	-0.1276,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.london.3',
	'europe-west2-c',
	51.5072,
	-0.1276,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.frankfurt.1',
	'europe-west3-a',
	50.1109,
	8.6821,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.frankfurt.2',
	'europe-west3-b',
	50.1109,
	8.6821,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.frankfurt.3',
	'europe-west3-c',
	50.1109,
	8.6821,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.netherlands.1',
	'europe-west4-a',
	53.4386,
	6.8355,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.netherlands.2',
	'europe-west4-b',
	53.4386,
	6.8355,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.netherlands.3',
	'europe-west4-c',
	53.4386,
	6.8355,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.zurich.1',
	'europe-west6-a',
	47.3769,
	8.5417,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.zurich.2',
	'europe-west6-b',
	47.3769,
	8.5417,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.zurich.3',
	'europe-west6-c',
	47.3769,
	8.5417,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.milan.1',
	'europe-west8-a',	
	45.4642,
	9.1900,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.milan.2',
	'europe-west8-b',
	45.4642,
	9.1900,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.milan.3',
	'europe-west8-c',
	45.4642,
	9.1900,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.paris.1',
	'europe-west9-a',
	48.8566,
	2.3522,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.paris.2',
	'europe-west9-b',
	48.8566,
	2.3522,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.paris.3',
	'europe-west9-c',
	48.8566,
	2.3522,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.telaviv.1',
	'me-west1-a',
	32.0853,
	34.7818,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.telaviv.2',
	'me-west1-b',
	32.0853,
	34.7818,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.telaviv.3',
	'me-west1-c',
	32.0853,
	34.7818,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.montreal.1',
	'northamerica-northeast1-a',
	45.5019,
	-73.5674,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.montreal.2',
	'northamerica-northeast1-b',
	45.5019,
	-73.5674,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.montreal.3',
	'northamerica-northeast1-c',
	45.5019,
	-73.5674,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.toronto.1',
	'northamerica-northeast1-a',
	43.6532,
	-79.3832,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.toronto.2',
	'northamerica-northeast2-b',
	43.6532,
	-79.3832,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.toronto.3',
	'northamerica-northeast2-c',
	43.6532,
	-79.3832,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saopaulo.1',
	'southamerica-east1-a',
	-23.5558,
	-46.6396,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saopaulo.2',
	'southamerica-east1-b',
	-23.5558,
	-46.6396,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saopaulo.3',
	'southamerica-east1-c',
	-23.5558,
	-46.6396,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.santiago.1',
	'southamerica-west1-a',
	-33.4489,
	-70.6693,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.santiago.2',
	'southamerica-west1-b',
	-33.4489,
	-70.6693,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.santiago.3',
	'southamerica-west1-c',
	-33.4489,
	-70.6693,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.iowa.1',
	'us-central1-a',
	41.8780,
	-93.0977,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.iowa.2',
	'us-central1-b',
	41.8780,
	-93.0977,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.iowa.3',
	'us-central1-c',
	41.8780,
	-93.0977,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.iowa.4',
	'us-central1-f',
	41.8780,
	-93.0977,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.southcarolina.1',
	'us-east1-b',
	33.8361,
	-81.1637,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.southcarolina.2',
	'us-east1-c',
	33.8361,
	-81.1637,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.southcarolina.3',
	'us-east1-d',
	33.8361,
	-81.1637,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.virginia.1',
	'us-east4-a',
	37.4316,
	-78.6569,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.virginia.2',
	'us-east4-b',
	37.4316,
	-78.6569,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.virginia.3',
	'us-east4-c',
	37.4316,
	-78.6569,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.ohio.1',
	'us-east5-a',
	39.9612,
	-82.9988,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.ohio.2',
	'us-east5-b',
	39.9612,
	-82.9988,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.ohio.3',
	'us-east5-c',
	39.9612,
	-82.9988,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.dallas.1',
	'us-south1-a',
	32.7767,
	-96.7970,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.dallas.2',
	'us-south1-b',
	32.7767,
	-96.7970,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.dallas.3',
	'us-south1-c',
	32.7767,
	-96.7970,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.oregon.1',
	'us-west1-a',
	45.5946,
	-121.1787,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.oregon.2',
	'us-west1-b',
	45.5946,
	-121.1787,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.oregon.3',
	'us-west1-c',
	45.5946,
	-121.1787,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.losangeles.1',
	'us-west2-a',
	34.0522,
	-118.2437,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.losangeles.2',
	'us-west2-b',
	34.0522,
	-118.2437,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.losangeles.3',
	'us-west2-c',
	34.0522,
	-118.2437,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saltlakecity.1',
	'us-west3-a',
	40.7608,
	-111.8910,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saltlakecity.2',
	'us-west3-b',
	40.7608,
	-111.8910,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saltlakecity.3',
	'us-west3-c',
	40.7608,
	-111.8910,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.lasvegas.1',
	'us-west4-a',
	36.1716,
	-115.1391,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.lasvegas.2',
	'us-west4-b',
	36.1716,
	-115.1391,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.lasvegas.3',
	'us-west4-c',
	36.1716,
	-115.1391,
	(select seller_id from sellers where seller_name = 'google')
);
