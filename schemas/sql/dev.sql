
INSERT INTO customers 
(
	live,
	debug, 
	customer_name, 
	customer_code
) 
VALUES (
	true,
	true,
	'Raspberry',
	'raspberry'
);

INSERT INTO route_shaders(route_shader_name,force_next,route_select_threshold,route_switch_threshold) VALUES('raspberry', true, 300, 300);

INSERT INTO buyers
(
	buyer_name,
	public_key_base64, 
	customer_id,
	route_shader_id
) 
VALUES(
	'raspberry',
	'UoFYERKJnCt18mU53IsWzlEXD2pYD9yd+TiZiq9+cMF9cHG4kMwRtw==',
	(select customer_id from customers where customer_code = 'raspberry'),
	(select route_shader_id from route_shaders where route_shader_name = 'raspberry')
);

INSERT INTO sellers(seller_name) VALUES('google');
INSERT INTO sellers(seller_name) VALUES('amazon');
INSERT INTO sellers(seller_name) VALUES('vultr');
INSERT INTO sellers(seller_name) VALUES('linode');

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
	'amazon.oregon.2',
	'usw2-az2',
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
	'amazon.oregon.3',
	'usw2-az3',
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
	'amazon.oregon.4',
	'usw2-az4',
	37.3387,
	-121.8853,
	(select seller_id from sellers where seller_name = 'amazon')
);

/*
use1-atl1-az1,amazon.atlanta.1
use1-bos1-az1,amazon.boston.1
use1-bue1-az1,amazon.buenosaires.1
use1-chi1-az1,amazon.chicago.1
use1-dfw1-az1,amazon.dallas.1
use1-iah1-az1,amazon.houston.1
use1-lim1-az1,amazon.lima.1
use1-mci1-az1,amazon.kansas.1
use1-mia1-az1,amazon.miami.1
use1-msp1-az1,amazon.minneapolis.1
use1-nyc1-az1,amazon.newyork.1
use1-phl1-az1,amazon.philadelphia.1
use1-qro1-az1,amazon.queretaro.1
use1-scl1-az1,amazon.santiago.1
*/











INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.oregon.1',
	45.8399,
	-119.7006,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.saopaulo.1',
	-23.5558, 
	-46.6396,
	(select seller_id from sellers where seller_name = 'amazon')
);

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

-- linode datacenters

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.atlanta',
	33.7488,
	-84.3877,
	(select seller_id from sellers where seller_name = 'linode')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.dallas',
	32.7767,
	-96.7970,
	(select seller_id from sellers where seller_name = 'linode')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.fremont',
	37.3387,
	-121.8853,
	(select seller_id from sellers where seller_name = 'linode')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.newark',
	40.7357,
	-74.1724,
	(select seller_id from sellers where seller_name = 'linode')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.toronto',
	43.6532,
	79.3832,
	(select seller_id from sellers where seller_name = 'linode')
);

-- vultr datacenters 

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.atlanta',
	33.7488,
	-84.3877,
	(select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.chicago',
	41.8781,
	-87.6298,
	(select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.dallas',
	32.7767,
	-96.7970,
	(select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.honolulu',
	21.3099,
	-157.8581,
	(select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.losangeles',
	34.0522,
	118.2437,
	(select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.miami',
	25.7617,
	-80.1918,
	(select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.newyork',
	40.7128,
	-74.0060,
	(select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.seattle',
	47.6062,
	-122.3321,
	(select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.siliconvalley',
	37.3387,
	-121.8853,
	(select seller_id from sellers where seller_name = 'vultr')
);

-- enable datacenters for buyers

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_name = 'raspberry'),
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.1'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_name = 'raspberry'),
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.2'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_name = 'raspberry'),
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.3'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_name = 'raspberry'),
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.4'),
	true
);
