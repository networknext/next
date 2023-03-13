
-- amazon datacenters

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.capetown.1',
	'afs1-az1',
	-33.9249,
	18.4241,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.capetown.2',
	'afs1-az2',
	-33.9249,
	18.4241,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.capetown.3',
	'afs1-az3',
	-33.9249,
	18.4241,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.hongkong.1',
	'ape1-az1',
	22.3193, 
	114.1694,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.hongkong.2',
	'ape1-az2',
	22.3193, 
	114.1694,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.hongkong.3',
	'ape1-az3',
	22.3193, 
	114.1694,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.tokyo.1',
	'apne1-az1',
	35.6762,
	139.6503,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.tokyo.2',
	'apne1-az2',
	35.6762,
	139.6503,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.tokyo.3',
	'apne1-az4',
	35.6762,
	139.6503,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.seoul.1',
	'apne2-az1',
	37.5665,
	126.9780,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.seoul.2',
	'apne2-az2',
	37.5665,
	126.9780,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.seoul.3',
	'apne2-az3',
	37.5665,
	126.9780,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.seoul.4',
	'apne2-az4',
	37.5665,
	126.9780,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.osaka.1',
	'apne3-az1',
	34.6937,
	135.5023,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.osaka.2',
	'apne3-az2',
	34.6937,
	135.5023,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.osaka.3',
	'apne3-az3',
	34.6937,
	135.5023,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.mumbai.1',
	'aps1-az1',
	19.0760,
	72.8777,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.mumbai.2',
	'aps1-az2',
	19.0760,
	72.8777,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.mumbai.3',
	'aps1-az3',
	19.0760,
	72.8777,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.hyderabad.1',
	'aps2-az1',
	19.0760,
	72.8777,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.hyderabad.2',
	'aps2-az2',
	19.0760,
	72.8777,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.hyderabad.3',
	'aps2-az3',
	19.0760,
	72.8777,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.singapore.1',
	'apse1-az1',
	1.3521,
	103.8198,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.singapore.2',
	'apse1-az2',
	1.3521,
	103.8198,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.singapore.3',
	'apse1-az3',
	1.3521,
	103.8198,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.sydney.1',
	'apse2-az1',
	-33.8688,
	151.2093,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.sydney.2',
	'apse2-az2',
	-33.8688,
	151.2093,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.sydney.3',
	'apse2-az3',
	-33.8688,
	151.2093,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.jakarta.1',
	'apse3-az1',
	-6.2088,
	106.8456,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.jakarta.2',
	'apse3-az2',
	-6.2088,
	106.8456,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.jakarta.3',
	'apse3-az3',
	-6.2088,
	106.8456,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.melbourne.1',
	'apse4-az1',
	-37.8136,
	144.9631,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.melbourne.2',
	'apse4-az2',
	-37.8136,
	144.9631,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.melbourne.3',
	'apse4-az3',
	-37.8136,
	144.9631,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.montreal.1',
	'cac1-az1',
	45.5019,
	-73.5674,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.montreal.2',
	'cac1-az2',
	45.5019,
	-73.5674,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.montreal.3',
	'cac1-az4',
	45.5019,
	-73.5674,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.frankfurt.1',
	'euc1-az1',
	50.1109,
	8.6821,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.frankfurt.2',
	'euc1-az2',
	50.1109,
	8.6821,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.frankfurt.3',
	'euc1-az3',
	50.1109,
	8.6821,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.zurich.1',
	'euc2-az1',
	47.3769,
	8.5417,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.zurich.2',
	'euc2-az2',
	47.3769,
	8.5417,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.zurich.3',
	'euc2-az3',
	47.3769,
	8.5417,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.stockholm.1',
	'eun1-az1',
	59.3293,
	18.0686,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.stockholm.2',
	'eun1-az2',
	59.3293,
	18.0686,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.stockholm.3',
	'eun1-az3',
	59.3293,
	18.0686,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.milan.1',
	'eus1-az1',
	45.4642,
	9.1900,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.milan.2',
	'eus1-az2',
	45.4642,
	9.1900,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.milan.3',
	'eus1-az3',
	45.4642,
	9.1900,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.spain.1',
	'eus2-az1',
	41.5976,
	-0.9057,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.spain.2',
	'eus2-az2',
	41.5976,
	-0.9057,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.spain.3',
	'eus2-az3',
	41.5976,
	-0.9057,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.ireland.1',
	'euw1-az1',
	53.7798,
	-7.3055,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.ireland.2',
	'euw1-az2',
	53.7798,
	-7.3055,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.ireland.3',
	'euw1-az3',
	53.7798,
	-7.3055,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.london.1',
	'euw2-az1',
	51.5072,
	-0.1276,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.london.2',
	'euw2-az2',
	51.5072,
	-0.1276,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.london.3',
	'euw2-az3',
	51.5072,
	-0.1276,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.paris.1',
	'euw3-az1',
	48.8566,
	2.3522,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.paris.2',
	'euw3-az2',
	48.8566,
	2.3522,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.paris.3',
	'euw3-az3',
	48.8566,
	2.3522,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.uae.1',
	'mec1-az1',
	23.4241,
	53.8478,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.uae.2',
	'mec1-az2',
	23.4241,
	53.8478,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.uae.3',
	'mec1-az3',
	23.4241,
	53.8478,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.bahrain.1',
	'mes1-az1',
	26.0667,
	50.5577,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.bahrain.2',
	'mes1-az2',
	26.0667,
	50.5577,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.bahrain.3',
	'mes1-az3',
	26.0667,
	50.5577,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.saopaulo.1',
	'sae1-az1',
	-23.5558,
	-46.6396,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.saopaulo.2',
	'sae1-az1',
	-23.5558,
	-46.6396,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.saopaulo.3',
	'sae1-az1',
	-23.5558,
	-46.6396,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.virginia.1',
	'use1-az1',
	39.0438,
	-77.4874,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.virginia.2',
	'use1-az2',
	39.0438,
	-77.4874,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.virginia.3',
	'use1-az3',
	39.0438,
	-77.4874,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.virginia.4',
	'use1-az4',
	39.0438,
	-77.4874,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.virginia.5',
	'use1-az5',
	39.0438,
	-77.4874,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.virginia.6',
	'use1-az6',
	39.0438,
	-77.4874,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.ohio.1',
	'use2-az1',
	40.4173, 
	-82.9071,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.ohio.2',
	'use2-az2',
	40.4173, 
	-82.9071,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.ohio.3',
	'use2-az3',
	40.4173, 
	-82.9071,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.sanjose.1',
	'usw1-az1',
	37.3387,
	-121.8853,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.sanjose.2',
	'usw1-az3',
	37.3387,
	-121.8853,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.oregon.1',
	'usw2-az1',
	45.8399,
	-119.7006,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.oregon.2',
	'usw2-az2',
	45.8399,
	-119.7006,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.oregon.3',
	'usw2-az3',
	45.8399,
	-119.7006,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.oregon.4',
	'usw2-az4',
	45.8399,
	-119.7006,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.atlanta.1',
	'use1-atl1-az1',
	33.7488,
	-84.3877,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.boston.1',
	'use1-bos1-az1',
	42.3601,
	-71.0589,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.buenosaires.1',
	'use1-bue1-az1',
	42.3601,
	-71.0589,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.chicago.1',
	'use1-chi1-az1',
	41.8781,
	-87.6298,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.dallas.1',
	'use1-dfw1-az1',
	32.7767,
	-96.7970,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.houston.1',
	'use1-iah1-az1',
	29.7604,
	-95.3698,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.lima.1',
	'use1-lim1-az1',
	-12.0464,
	-77.0428,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.kansas.1',
	'use1-mci1-az1',
	39.0997,
	-94.5786,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.miami.1',
	'use1-mia1-az1',
	25.7617,
	-80.1918,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.minneapolis.1',
	'use1-msp1-az1',
	44.9778,
	-93.2650,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.newyork.1',
	'use1-nyc1-az1',
	40.7128,
	-74.0060,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.philadelphia.1',
	'use1-phl1-az1',
	39.9526,
	-75.1652,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.queretaro.1',
	'use1-qro1-az1',
	20.5888,
	-100.3899,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	native_name,
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.santiago.1',
	'use1-scl1-az1',
	-33.4489,
	-70.6693,
	(select seller_id from sellers where seller_name = 'amazon')
);
