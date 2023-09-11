
INSERT INTO route_shaders(route_shader_name) VALUES('test');

INSERT INTO buyers
(
	buyer_name,
	buyer_code,
	live,
	public_key_base64, 
	route_shader_id
) 
VALUES(
	'Test',
	'test',
	true,
	'leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==',
	(select route_shader_id from route_shaders where route_shader_name = 'test')
);

INSERT INTO sellers(seller_name, seller_code) VALUES('Test', 'test');

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.0',
	-2.00,
	33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.1',
	41.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.2',
	-34.00,
	-138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.3',
	-39.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.4',
	68.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.5',
	53.00,
	97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.6',
	-45.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.7',
	-34.00,
	151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.8',
	41.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.9',
	9.00,
	101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.10',
	-68.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.11',
	13.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.12',
	-4.00,
	149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.13',
	70.00,
	-64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.14',
	41.00,
	90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.15',
	31.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.16',
	-19.00,
	170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.17',
	-58.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.18',
	81.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.19',
	-55.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.20',
	18.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.21',
	-73.00,
	23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.22',
	58.00,
	106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.23',
	-31.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.24',
	-59.00,
	-63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.25',
	8.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.26',
	5.00,
	67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.27',
	-57.00,
	-102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.28',
	25.00,
	-138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.29',
	-71.00,
	-139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.30',
	-53.00,
	32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.31',
	54.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.32',
	-27.00,
	-120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.33',
	20.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.34',
	-21.00,
	-107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.35',
	46.00,
	71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.36',
	-19.00,
	90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.37',
	23.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.38',
	90.00,
	-88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.39',
	-29.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.40',
	-10.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.41',
	74.00,
	42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.42',
	22.00,
	23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.43',
	-7.00,
	155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.44',
	-30.00,
	-106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.45',
	60.00,
	64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.46',
	-13.00,
	-60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.47',
	34.00,
	-129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.48',
	-10.00,
	-62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.49',
	81.00,
	-161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.50',
	86.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.51',
	37.00,
	-66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.52',
	80.00,
	149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.53',
	-27.00,
	-19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.54',
	-14.00,
	-62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.55',
	-49.00,
	-10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.56',
	7.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.57',
	10.00,
	-146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.58',
	12.00,
	120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.59',
	75.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.60',
	64.00,
	75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.61',
	-31.00,
	88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.62',
	-2.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.63',
	-49.00,
	95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.64',
	89.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.65',
	49.00,
	9.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.66',
	18.00,
	113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.67',
	9.00,
	-6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.68',
	46.00,
	-41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.69',
	-24.00,
	-168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.70',
	40.00,
	-3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.71',
	10.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.72',
	79.00,
	-146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.73',
	-61.00,
	42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.74',
	-26.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.75',
	-73.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.76',
	69.00,
	-2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.77',
	34.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.78',
	-34.00,
	41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.79',
	-16.00,
	-141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.80',
	13.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.81',
	13.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.82',
	52.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.83',
	-5.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.84',
	-5.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.85',
	-36.00,
	47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.86',
	18.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.87',
	-90.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.88',
	-84.00,
	-30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.89',
	14.00,
	143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.90',
	-43.00,
	22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.91',
	-52.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.92',
	66.00,
	-122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.93',
	83.00,
	-160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.94',
	90.00,
	-68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.95',
	-22.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.96',
	43.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.97',
	-74.00,
	-136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.98',
	-5.00,
	130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.99',
	58.00,
	-106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.100',
	2.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.101',
	84.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.102',
	-47.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.103',
	-25.00,
	-160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.104',
	26.00,
	-121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.105',
	-64.00,
	57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.106',
	-51.00,
	89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.107',
	83.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.108',
	-59.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.109',
	-80.00,
	90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.110',
	-17.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.111',
	54.00,
	14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.112',
	-72.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.113',
	-70.00,
	124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.114',
	47.00,
	-178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.115',
	19.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.116',
	43.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.117',
	-10.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.118',
	6.00,
	52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.119',
	-31.00,
	-54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.120',
	-25.00,
	-35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.121',
	-28.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.122',
	73.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.123',
	-11.00,
	-145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.124',
	72.00,
	158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.125',
	0.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.126',
	-62.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.127',
	-85.00,
	-24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.128',
	84.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.129',
	-28.00,
	-151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.130',
	70.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.131',
	-76.00,
	4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.132',
	70.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.133',
	-11.00,
	-146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.134',
	44.00,
	51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.135',
	8.00,
	16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.136',
	45.00,
	19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.137',
	-20.00,
	127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.138',
	12.00,
	-76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.139',
	-65.00,
	-155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.140',
	-3.00,
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.141',
	-39.00,
	22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.142',
	53.00,
	-54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.143',
	21.00,
	73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.144',
	-83.00,
	-175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.145',
	-48.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.146',
	-45.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.147',
	-63.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.148',
	51.00,
	51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.149',
	82.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.150',
	-12.00,
	68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.151',
	-48.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.152',
	85.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.153',
	43.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.154',
	-80.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.155',
	-17.00,
	-107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.156',
	59.00,
	-167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.157',
	17.00,
	-90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.158',
	58.00,
	160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.159',
	82.00,
	34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.160',
	-7.00,
	-79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.161',
	-76.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.162',
	85.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.163',
	49.00,
	151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.164',
	73.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.165',
	0.00,
	-37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.166',
	-84.00,
	148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.167',
	11.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.168',
	18.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.169',
	6.00,
	28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.170',
	-5.00,
	68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.171',
	-53.00,
	-167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.172',
	37.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.173',
	4.00,
	33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.174',
	89.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.175',
	-62.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.176',
	-82.00,
	-140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.177',
	-8.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.178',
	-59.00,
	-142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.179',
	-32.00,
	53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.180',
	-65.00,
	-115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.181',
	35.00,
	151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.182',
	-30.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.183',
	-24.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.184',
	54.00,
	-138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.185',
	82.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.186',
	46.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.187',
	8.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.188',
	53.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.189',
	-50.00,
	-122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.190',
	61.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.191',
	7.00,
	-178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.192',
	80.00,
	-98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.193',
	8.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.194',
	-60.00,
	-102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.195',
	-35.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.196',
	-66.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.197',
	27.00,
	131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.198',
	-62.00,
	-139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.199',
	-13.00,
	-57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.200',
	-36.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.201',
	-61.00,
	-146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.202',
	40.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.203',
	-77.00,
	-180.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.204',
	-18.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.205',
	-81.00,
	-141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.206',
	-29.00,
	146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.207',
	-86.00,
	-95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.208',
	55.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.209',
	36.00,
	41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.210',
	-17.00,
	-175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.211',
	-11.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.212',
	-20.00,
	-71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.213',
	-33.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.214',
	17.00,
	-126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.215',
	-68.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.216',
	5.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.217',
	9.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.218',
	44.00,
	128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.219',
	9.00,
	-176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.220',
	7.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.221',
	74.00,
	-168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.222',
	-35.00,
	49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.223',
	-27.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.224',
	35.00,
	-19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.225',
	-83.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.226',
	36.00,
	158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.227',
	78.00,
	-177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.228',
	51.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.229',
	-30.00,
	-18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.230',
	-72.00,
	86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.231',
	-1.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.232',
	-66.00,
	95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.233',
	-54.00,
	94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.234',
	15.00,
	32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.235',
	-39.00,
	159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.236',
	-45.00,
	-32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.237',
	50.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.238',
	63.00,
	-82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.239',
	49.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.240',
	-67.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.241',
	57.00,
	-120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.242',
	-12.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.243',
	64.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.244',
	-86.00,
	-166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.245',
	-63.00,
	3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.246',
	82.00,
	-53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.247',
	37.00,
	3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.248',
	89.00,
	34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.249',
	-21.00,
	-42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.250',
	-86.00,
	49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.251',
	-15.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.252',
	-62.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.253',
	57.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.254',
	27.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.255',
	-57.00,
	30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.256',
	42.00,
	-24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.257',
	-22.00,
	38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.258',
	-72.00,
	79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.259',
	-85.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.260',
	18.00,
	137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.261',
	29.00,
	-145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.262',
	41.00,
	89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.263',
	21.00,
	-117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.264',
	-60.00,
	-141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.265',
	-70.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.266',
	51.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.267',
	-35.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.268',
	3.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.269',
	69.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.270',
	14.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.271',
	88.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.272',
	39.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.273',
	63.00,
	111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.274',
	64.00,
	-33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.275',
	86.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.276',
	-32.00,
	-3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.277',
	-84.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.278',
	-21.00,
	21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.279',
	48.00,
	26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.280',
	82.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.281',
	-22.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.282',
	77.00,
	-58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.283',
	-15.00,
	-40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.284',
	27.00,
	-11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.285',
	38.00,
	-110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.286',
	-7.00,
	59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.287',
	74.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.288',
	56.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.289',
	63.00,
	-142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.290',
	-88.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.291',
	69.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.292',
	-27.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.293',
	-44.00,
	110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.294',
	-30.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.295',
	50.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.296',
	-54.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.297',
	-39.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.298',
	-85.00,
	104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.299',
	73.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.300',
	-80.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.301',
	19.00,
	17.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.302',
	-77.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.303',
	60.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.304',
	-58.00,
	2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.305',
	-53.00,
	-121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.306',
	72.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.307',
	-73.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.308',
	-16.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.309',
	-47.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.310',
	65.00,
	28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.311',
	-4.00,
	-5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.312',
	78.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.313',
	-39.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.314',
	80.00,
	176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.315',
	-21.00,
	94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.316',
	-48.00,
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.317',
	11.00,
	-68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.318',
	-33.00,
	49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.319',
	33.00,
	136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.320',
	47.00,
	-148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.321',
	39.00,
	-28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.322',
	-16.00,
	175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.323',
	-78.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.324',
	-37.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.325',
	4.00,
	-136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.326',
	82.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.327',
	-40.00,
	-133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.328',
	28.00,
	-106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.329',
	-82.00,
	35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.330',
	-34.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.331',
	17.00,
	178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.332',
	-89.00,
	-39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.333',
	-30.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.334',
	69.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.335',
	36.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.336',
	86.00,
	161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.337',
	-33.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.338',
	21.00,
	-115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.339',
	-83.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.340',
	-15.00,
	-167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.341',
	-18.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.342',
	38.00,
	-160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.343',
	-28.00,
	-166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.344',
	-51.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.345',
	-7.00,
	-5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.346',
	61.00,
	51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.347',
	53.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.348',
	-82.00,
	107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.349',
	38.00,
	112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.350',
	51.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.351',
	-90.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.352',
	-15.00,
	23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.353',
	-13.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.354',
	57.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.355',
	60.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.356',
	32.00,
	74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.357',
	-44.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.358',
	-4.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.359',
	64.00,
	-53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.360',
	-82.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.361',
	-7.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.362',
	-24.00,
	-173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.363',
	19.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.364',
	-35.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.365',
	-5.00,
	57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.366',
	-5.00,
	-3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.367',
	9.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.368',
	-90.00,
	26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.369',
	-72.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.370',
	66.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.371',
	3.00,
	-155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.372',
	74.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.373',
	60.00,
	-161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.374',
	50.00,
	46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.375',
	54.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.376',
	-56.00,
	150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.377',
	78.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.378',
	-22.00,
	69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.379',
	84.00,
	86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.380',
	28.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.381',
	-60.00,
	163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.382',
	-39.00,
	149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.383',
	35.00,
	-162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.384',
	-18.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.385',
	-25.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.386',
	50.00,
	151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.387',
	-11.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.388',
	47.00,
	-153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.389',
	-66.00,
	99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.390',
	-56.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.391',
	4.00,
	-32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.392',
	-86.00,
	-86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.393',
	8.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.394',
	-51.00,
	-125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.395',
	35.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.396',
	-34.00,
	167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.397',
	62.00,
	176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.398',
	-10.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.399',
	-70.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.400',
	54.00,
	-82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.401',
	9.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.402',
	-49.00,
	155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.403',
	5.00,
	102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.404',
	9.00,
	-72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.405',
	14.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.406',
	88.00,
	-64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.407',
	-84.00,
	49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.408',
	-23.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.409',
	-17.00,
	-5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.410',
	-47.00,
	-154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.411',
	21.00,
	-28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.412',
	-50.00,
	-7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.413',
	-24.00,
	-179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.414',
	38.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.415',
	-13.00,
	8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.416',
	78.00,
	-99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.417',
	85.00,
	-49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.418',
	51.00,
	-82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.419',
	-43.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.420',
	-28.00,
	-151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.421',
	-49.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.422',
	-54.00,
	0.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.423',
	67.00,
	-47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.424',
	-4.00,
	115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.425',
	28.00,
	-106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.426',
	20.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.427',
	-85.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.428',
	79.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.429',
	56.00,
	41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.430',
	40.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.431',
	-59.00,
	92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.432',
	-12.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.433',
	64.00,
	146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.434',
	34.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.435',
	56.00,
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.436',
	-47.00,
	-175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.437',
	71.00,
	8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.438',
	51.00,
	-5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.439',
	10.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.440',
	44.00,
	-121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.441',
	-53.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.442',
	80.00,
	-33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.443',
	24.00,
	63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.444',
	-58.00,
	-138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.445',
	49.00,
	94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.446',
	1.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.447',
	38.00,
	163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.448',
	-89.00,
	104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.449',
	8.00,
	144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.450',
	-56.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.451',
	26.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.452',
	-74.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.453',
	-70.00,
	107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.454',
	87.00,
	-40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.455',
	-90.00,
	-16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.456',
	-83.00,
	110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.457',
	30.00,
	-95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.458',
	82.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.459',
	-7.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.460',
	4.00,
	-131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.461',
	7.00,
	95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.462',
	79.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.463',
	23.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.464',
	-72.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.465',
	-55.00,
	-162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.466',
	38.00,
	-58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.467',
	-1.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.468',
	55.00,
	-98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.469',
	-62.00,
	-6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.470',
	36.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.471',
	50.00,
	-3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.472',
	5.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.473',
	-72.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.474',
	83.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.475',
	-23.00,
	-178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.476',
	50.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.477',
	-31.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.478',
	57.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.479',
	36.00,
	149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.480',
	-52.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.481',
	-56.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.482',
	4.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.483',
	-33.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.484',
	76.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.485',
	-31.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.486',
	16.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.487',
	89.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.488',
	-77.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.489',
	-45.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.490',
	-2.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.491',
	-3.00,
	54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.492',
	27.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.493',
	-7.00,
	61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.494',
	-55.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.495',
	81.00,
	-162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.496',
	13.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.497',
	-23.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.498',
	-26.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.499',
	56.00,
	97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.500',
	-65.00,
	146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.501',
	24.00,
	-58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.502',
	-79.00,
	-65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.503',
	84.00,
	-30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.504',
	-24.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.505',
	-31.00,
	97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.506',
	-87.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.507',
	-21.00,
	-81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.508',
	78.00,
	163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.509',
	-36.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.510',
	39.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.511',
	-75.00,
	-147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.512',
	57.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.513',
	15.00,
	-137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.514',
	-88.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.515',
	-49.00,
	-121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.516',
	54.00,
	-177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.517',
	-52.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.518',
	-41.00,
	143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.519',
	48.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.520',
	-40.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.521',
	25.00,
	46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.522',
	-71.00,
	-68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.523',
	-49.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.524',
	-38.00,
	-74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.525',
	-85.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.526',
	-12.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.527',
	41.00,
	19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.528',
	57.00,
	-171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.529',
	-5.00,
	131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.530',
	38.00,
	75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.531',
	17.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.532',
	53.00,
	-141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.533',
	23.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.534',
	-22.00,
	8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.535',
	-57.00,
	-16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.536',
	-81.00,
	107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.537',
	24.00,
	24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.538',
	24.00,
	15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.539',
	19.00,
	-172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.540',
	-38.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.541',
	4.00,
	-24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.542',
	15.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.543',
	-53.00,
	137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.544',
	35.00,
	3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.545',
	-33.00,
	-52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.546',
	36.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.547',
	-46.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.548',
	-76.00,
	44.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.549',
	1.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.550',
	-41.00,
	-41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.551',
	0.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.552',
	-51.00,
	-66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.553',
	-47.00,
	-39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.554',
	80.00,
	178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.555',
	71.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.556',
	60.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.557',
	28.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.558',
	64.00,
	-171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.559',
	-25.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.560',
	-75.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.561',
	15.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.562',
	-42.00,
	-131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.563',
	83.00,
	-71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.564',
	-61.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.565',
	6.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.566',
	-7.00,
	32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.567',
	-82.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.568',
	89.00,
	178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.569',
	14.00,
	-117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.570',
	-88.00,
	-86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.571',
	89.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.572',
	69.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.573',
	-7.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.574',
	-56.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.575',
	-35.00,
	-161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.576',
	90.00,
	-47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.577',
	56.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.578',
	-27.00,
	-146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.579',
	-41.00,
	59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.580',
	-36.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.581',
	52.00,
	64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.582',
	-14.00,
	101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.583',
	-8.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.584',
	3.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.585',
	72.00,
	-134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.586',
	83.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.587',
	-31.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.588',
	10.00,
	41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.589',
	86.00,
	-3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.590',
	-76.00,
	154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.591',
	-79.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.592',
	-73.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.593',
	-80.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.594',
	43.00,
	37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.595',
	-78.00,
	26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.596',
	-25.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.597',
	52.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.598',
	78.00,
	-74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.599',
	-66.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.600',
	11.00,
	59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.601',
	27.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.602',
	49.00,
	176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.603',
	-69.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.604',
	-38.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.605',
	-31.00,
	161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.606',
	41.00,
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.607',
	12.00,
	-87.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.608',
	53.00,
	32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.609',
	-46.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.610',
	12.00,
	135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.611',
	-50.00,
	54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.612',
	-57.00,
	-88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.613',
	-7.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.614',
	51.00,
	162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.615',
	7.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.616',
	20.00,
	-82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.617',
	13.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.618',
	-32.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.619',
	-33.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.620',
	65.00,
	26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.621',
	4.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.622',
	8.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.623',
	52.00,
	-179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.624',
	-54.00,
	112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.625',
	65.00,
	-115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.626',
	3.00,
	-84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.627',
	-34.00,
	-35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.628',
	49.00,
	151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.629',
	-16.00,
	-60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.630',
	-69.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.631',
	-42.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.632',
	-7.00,
	-75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.633',
	-18.00,
	-170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.634',
	15.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.635',
	17.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.636',
	-21.00,
	115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.637',
	-50.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.638',
	15.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.639',
	24.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.640',
	75.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.641',
	22.00,
	98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.642',
	44.00,
	-139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.643',
	51.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.644',
	35.00,
	72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.645',
	12.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.646',
	28.00,
	-15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.647',
	84.00,
	107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.648',
	-62.00,
	58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.649',
	-54.00,
	41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.650',
	51.00,
	73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.651',
	34.00,
	-44.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.652',
	59.00,
	-57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.653',
	1.00,
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.654',
	38.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.655',
	64.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.656',
	74.00,
	-63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.657',
	47.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.658',
	74.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.659',
	21.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.660',
	-57.00,
	119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.661',
	-15.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.662',
	-83.00,
	-53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.663',
	4.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.664',
	-1.00,
	63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.665',
	10.00,
	-34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.666',
	-80.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.667',
	-15.00,
	102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.668',
	0.00,
	143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.669',
	-69.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.670',
	-5.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.671',
	74.00,
	23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.672',
	-11.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.673',
	-9.00,
	-156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.674',
	-61.00,
	-95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.675',
	88.00,
	-137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.676',
	-86.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.677',
	79.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.678',
	66.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.679',
	-1.00,
	-50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.680',
	-6.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.681',
	69.00,
	-58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.682',
	-37.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.683',
	65.00,
	72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.684',
	-8.00,
	-146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.685',
	49.00,
	19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.686',
	-21.00,
	36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.687',
	23.00,
	-136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.688',
	45.00,
	127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.689',
	77.00,
	132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.690',
	0.00,
	155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.691',
	-66.00,
	19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.692',
	-9.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.693',
	48.00,
	35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.694',
	44.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.695',
	-85.00,
	-166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.696',
	40.00,
	88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.697',
	-21.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.698',
	-38.00,
	-82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.699',
	82.00,
	51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.700',
	-71.00,
	106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.701',
	-62.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.702',
	54.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.703',
	-6.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.704',
	83.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.705',
	36.00,
	-137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.706',
	-62.00,
	-65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.707',
	-61.00,
	-148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.708',
	-60.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.709',
	-66.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.710',
	-6.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.711',
	9.00,
	-18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.712',
	-28.00,
	-154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.713',
	78.00,
	-24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.714',
	24.00,
	79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.715',
	-54.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.716',
	-83.00,
	77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.717',
	-77.00,
	-129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.718',
	8.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.719',
	-40.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.720',
	66.00,
	57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.721',
	-86.00,
	10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.722',
	-43.00,
	-131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.723',
	31.00,
	-125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.724',
	59.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.725',
	14.00,
	110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.726',
	-21.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.727',
	51.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.728',
	44.00,
	-63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.729',
	-50.00,
	-47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.730',
	-87.00,
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.731',
	-37.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.732',
	51.00,
	-43.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.733',
	21.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.734',
	15.00,
	16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.735',
	6.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.736',
	-44.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.737',
	5.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.738',
	4.00,
	-151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.739',
	-41.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.740',
	-87.00,
	74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.741',
	51.00,
	-176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.742',
	-4.00,
	-121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.743',
	-31.00,
	-155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.744',
	13.00,
	-57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.745',
	-4.00,
	94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.746',
	-74.00,
	-106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.747',
	32.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.748',
	4.00,
	-90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.749',
	69.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.750',
	17.00,
	-98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.751',
	-12.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.752',
	83.00,
	-17.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.753',
	-6.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.754',
	30.00,
	-10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.755',
	40.00,
	110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.756',
	15.00,
	-37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.757',
	60.00,
	26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.758',
	-63.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.759',
	61.00,
	0.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.760',
	-2.00,
	-174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.761',
	-81.00,
	77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.762',
	22.00,
	115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.763',
	56.00,
	178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.764',
	75.00,
	-15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.765',
	-50.00,
	-159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.766',
	-10.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.767',
	-59.00,
	-3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.768',
	-32.00,
	34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.769',
	-12.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.770',
	-44.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.771',
	-79.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.772',
	-34.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.773',
	-45.00,
	10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.774',
	-10.00,
	-90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.775',
	-36.00,
	-10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.776',
	5.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.777',
	-1.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.778',
	28.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.779',
	-23.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.780',
	34.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.781',
	-12.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.782',
	-86.00,
	26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.783',
	-9.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.784',
	44.00,
	-86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.785',
	55.00,
	179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.786',
	60.00,
	-56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.787',
	33.00,
	167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.788',
	58.00,
	176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.789',
	6.00,
	-87.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.790',
	61.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.791',
	-69.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.792',
	82.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.793',
	-70.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.794',
	-76.00,
	170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.795',
	-20.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.796',
	49.00,
	15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.797',
	-7.00,
	-141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.798',
	85.00,
	58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.799',
	81.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.800',
	58.00,
	-69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.801',
	-49.00,
	59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.802',
	-44.00,
	-11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.803',
	-60.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.804',
	-6.00,
	-24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.805',
	-30.00,
	-73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.806',
	4.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.807',
	81.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.808',
	-35.00,
	-69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.809',
	46.00,
	34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.810',
	72.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.811',
	-72.00,
	144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.812',
	-88.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.813',
	54.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.814',
	-37.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.815',
	48.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.816',
	6.00,
	-151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.817',
	19.00,
	-156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.818',
	-89.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.819',
	-14.00,
	10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.820',
	28.00,
	-69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.821',
	-56.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.822',
	11.00,
	54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.823',
	67.00,
	143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.824',
	76.00,
	161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.825',
	-28.00,
	162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.826',
	-24.00,
	9.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.827',
	-77.00,
	-167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.828',
	82.00,
	-40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.829',
	-8.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.830',
	-18.00,
	130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.831',
	-38.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.832',
	67.00,
	-102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.833',
	-35.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.834',
	-31.00,
	61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.835',
	11.00,
	73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.836',
	13.00,
	106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.837',
	-57.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.838',
	27.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.839',
	-35.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.840',
	-47.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.841',
	53.00,
	108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.842',
	30.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.843',
	-14.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.844',
	-30.00,
	67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.845',
	-49.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.846',
	70.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.847',
	-62.00,
	-24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.848',
	-68.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.849',
	-70.00,
	104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.850',
	87.00,
	9.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.851',
	-89.00,
	42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.852',
	17.00,
	-175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.853',
	4.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.854',
	-19.00,
	35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.855',
	-36.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.856',
	41.00,
	-27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.857',
	70.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.858',
	-30.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.859',
	-27.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.860',
	-76.00,
	-16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.861',
	51.00,
	-82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.862',
	-16.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.863',
	62.00,
	77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.864',
	-17.00,
	14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.865',
	-21.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.866',
	74.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.867',
	-81.00,
	-7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.868',
	80.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.869',
	62.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.870',
	25.00,
	74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.871',
	31.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.872',
	-7.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.873',
	-54.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.874',
	-46.00,
	-133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.875',
	-9.00,
	-147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.876',
	-76.00,
	143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.877',
	39.00,
	-172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.878',
	0.00,
	108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.879',
	39.00,
	99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.880',
	-86.00,
	51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.881',
	-12.00,
	35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.882',
	-59.00,
	88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.883',
	-7.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.884',
	38.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.885',
	-50.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.886',
	-47.00,
	64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.887',
	-90.00,
	-178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.888',
	-54.00,
	-155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.889',
	59.00,
	87.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.890',
	42.00,
	-18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.891',
	-34.00,
	-173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.892',
	21.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.893',
	51.00,
	63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.894',
	-81.00,
	-161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.895',
	-11.00,
	136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.896',
	-63.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.897',
	10.00,
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.898',
	-11.00,
	-126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.899',
	-46.00,
	175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.900',
	-27.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.901',
	65.00,
	73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.902',
	-40.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.903',
	-73.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.904',
	54.00,
	-125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.905',
	-51.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.906',
	-62.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.907',
	67.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.908',
	24.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.909',
	-66.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.910',
	90.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.911',
	-63.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.912',
	-22.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.913',
	60.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.914',
	-58.00,
	-74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.915',
	72.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.916',
	-10.00,
	94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.917',
	-69.00,
	-118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.918',
	56.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.919',
	26.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.920',
	8.00,
	167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.921',
	8.00,
	16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.922',
	-10.00,
	-10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.923',
	-26.00,
	30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.924',
	-30.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.925',
	-42.00,
	154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.926',
	-9.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.927',
	-60.00,
	-74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.928',
	80.00,
	143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.929',
	54.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.930',
	31.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.931',
	-70.00,
	-18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.932',
	82.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.933',
	-50.00,
	-57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.934',
	67.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.935',
	-54.00,
	92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.936',
	-79.00,
	55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.937',
	-45.00,
	131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.938',
	-22.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.939',
	72.00,
	54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.940',
	-17.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.941',
	-56.00,
	-14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.942',
	56.00,
	-101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.943',
	65.00,
	146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.944',
	-54.00,
	111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.945',
	76.00,
	146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.946',
	-85.00,
	16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.947',
	-13.00,
	117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.948',
	-79.00,
	-172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.949',
	20.00,
	42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.950',
	-53.00,
	99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.951',
	-75.00,
	-9.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.952',
	49.00,
	69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.953',
	29.00,
	174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.954',
	19.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.955',
	-16.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.956',
	-55.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.957',
	46.00,
	-52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.958',
	-8.00,
	51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.959',
	-85.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.960',
	-49.00,
	4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.961',
	-54.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.962',
	-20.00,
	-43.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.963',
	-29.00,
	58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.964',
	79.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.965',
	41.00,
	53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.966',
	-52.00,
	-115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.967',
	-87.00,
	78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.968',
	26.00,
	-145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.969',
	-64.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.970',
	-85.00,
	-65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.971',
	85.00,
	136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.972',
	88.00,
	79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.973',
	74.00,
	-133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.974',
	25.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.975',
	19.00,
	-119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.976',
	-77.00,
	-33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.977',
	78.00,
	144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.978',
	-70.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.979',
	48.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.980',
	-30.00,
	-145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.981',
	-90.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.982',
	-87.00,
	-148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.983',
	-19.00,
	148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.984',
	59.00,
	-75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.985',
	60.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.986',
	90.00,
	-153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.987',
	60.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.988',
	-43.00,
	-161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.989',
	-23.00,
	68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.990',
	45.00,
	21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.991',
	23.00,
	-18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.992',
	67.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.993',
	-29.00,
	-121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.994',
	60.00,
	-153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.995',
	-88.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.996',
	0.00,
	-35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.997',
	-13.00,
	-168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.998',
	49.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.999',
	77.00,
	-115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.0',
	'127.0.0.1',
	10000,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.1',
	'127.0.0.1',
	10001,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.2',
	'127.0.0.1',
	10002,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.3',
	'127.0.0.1',
	10003,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.4',
	'127.0.0.1',
	10004,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.5',
	'127.0.0.1',
	10005,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.6',
	'127.0.0.1',
	10006,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.7',
	'127.0.0.1',
	10007,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.8',
	'127.0.0.1',
	10008,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.9',
	'127.0.0.1',
	10009,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.10',
	'127.0.0.1',
	10010,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.11',
	'127.0.0.1',
	10011,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.12',
	'127.0.0.1',
	10012,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.13',
	'127.0.0.1',
	10013,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.14',
	'127.0.0.1',
	10014,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.15',
	'127.0.0.1',
	10015,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.16',
	'127.0.0.1',
	10016,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.17',
	'127.0.0.1',
	10017,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.18',
	'127.0.0.1',
	10018,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.19',
	'127.0.0.1',
	10019,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.20',
	'127.0.0.1',
	10020,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.21',
	'127.0.0.1',
	10021,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.22',
	'127.0.0.1',
	10022,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.23',
	'127.0.0.1',
	10023,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.24',
	'127.0.0.1',
	10024,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.25',
	'127.0.0.1',
	10025,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.26',
	'127.0.0.1',
	10026,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.27',
	'127.0.0.1',
	10027,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.28',
	'127.0.0.1',
	10028,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.29',
	'127.0.0.1',
	10029,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.30',
	'127.0.0.1',
	10030,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.31',
	'127.0.0.1',
	10031,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.32',
	'127.0.0.1',
	10032,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.33',
	'127.0.0.1',
	10033,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.34',
	'127.0.0.1',
	10034,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.35',
	'127.0.0.1',
	10035,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.36',
	'127.0.0.1',
	10036,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.37',
	'127.0.0.1',
	10037,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.38',
	'127.0.0.1',
	10038,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.39',
	'127.0.0.1',
	10039,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.40',
	'127.0.0.1',
	10040,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.41',
	'127.0.0.1',
	10041,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.42',
	'127.0.0.1',
	10042,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.43',
	'127.0.0.1',
	10043,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.44',
	'127.0.0.1',
	10044,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.45',
	'127.0.0.1',
	10045,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.46',
	'127.0.0.1',
	10046,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.47',
	'127.0.0.1',
	10047,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.48',
	'127.0.0.1',
	10048,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.49',
	'127.0.0.1',
	10049,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.50',
	'127.0.0.1',
	10050,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.51',
	'127.0.0.1',
	10051,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.52',
	'127.0.0.1',
	10052,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.53',
	'127.0.0.1',
	10053,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.54',
	'127.0.0.1',
	10054,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.55',
	'127.0.0.1',
	10055,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.56',
	'127.0.0.1',
	10056,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.57',
	'127.0.0.1',
	10057,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.58',
	'127.0.0.1',
	10058,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.59',
	'127.0.0.1',
	10059,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.60',
	'127.0.0.1',
	10060,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.61',
	'127.0.0.1',
	10061,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.62',
	'127.0.0.1',
	10062,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.63',
	'127.0.0.1',
	10063,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.64',
	'127.0.0.1',
	10064,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.65',
	'127.0.0.1',
	10065,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.66',
	'127.0.0.1',
	10066,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.67',
	'127.0.0.1',
	10067,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.68',
	'127.0.0.1',
	10068,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.69',
	'127.0.0.1',
	10069,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.70',
	'127.0.0.1',
	10070,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.71',
	'127.0.0.1',
	10071,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.72',
	'127.0.0.1',
	10072,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.73',
	'127.0.0.1',
	10073,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.74',
	'127.0.0.1',
	10074,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.75',
	'127.0.0.1',
	10075,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.76',
	'127.0.0.1',
	10076,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.77',
	'127.0.0.1',
	10077,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.78',
	'127.0.0.1',
	10078,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.79',
	'127.0.0.1',
	10079,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.80',
	'127.0.0.1',
	10080,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.81',
	'127.0.0.1',
	10081,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.82',
	'127.0.0.1',
	10082,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.83',
	'127.0.0.1',
	10083,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.84',
	'127.0.0.1',
	10084,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.85',
	'127.0.0.1',
	10085,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.86',
	'127.0.0.1',
	10086,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.87',
	'127.0.0.1',
	10087,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.88',
	'127.0.0.1',
	10088,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.89',
	'127.0.0.1',
	10089,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.90',
	'127.0.0.1',
	10090,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.91',
	'127.0.0.1',
	10091,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.92',
	'127.0.0.1',
	10092,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.93',
	'127.0.0.1',
	10093,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.94',
	'127.0.0.1',
	10094,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.95',
	'127.0.0.1',
	10095,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.96',
	'127.0.0.1',
	10096,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.97',
	'127.0.0.1',
	10097,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.98',
	'127.0.0.1',
	10098,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.99',
	'127.0.0.1',
	10099,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.100',
	'127.0.0.1',
	10100,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.101',
	'127.0.0.1',
	10101,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.102',
	'127.0.0.1',
	10102,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.103',
	'127.0.0.1',
	10103,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.104',
	'127.0.0.1',
	10104,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.105',
	'127.0.0.1',
	10105,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.106',
	'127.0.0.1',
	10106,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.107',
	'127.0.0.1',
	10107,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.108',
	'127.0.0.1',
	10108,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.109',
	'127.0.0.1',
	10109,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.110',
	'127.0.0.1',
	10110,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.111',
	'127.0.0.1',
	10111,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.112',
	'127.0.0.1',
	10112,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.113',
	'127.0.0.1',
	10113,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.114',
	'127.0.0.1',
	10114,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.115',
	'127.0.0.1',
	10115,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.116',
	'127.0.0.1',
	10116,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.117',
	'127.0.0.1',
	10117,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.118',
	'127.0.0.1',
	10118,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.119',
	'127.0.0.1',
	10119,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.120',
	'127.0.0.1',
	10120,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.121',
	'127.0.0.1',
	10121,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.122',
	'127.0.0.1',
	10122,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.123',
	'127.0.0.1',
	10123,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.124',
	'127.0.0.1',
	10124,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.125',
	'127.0.0.1',
	10125,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.126',
	'127.0.0.1',
	10126,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.127',
	'127.0.0.1',
	10127,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.128',
	'127.0.0.1',
	10128,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.129',
	'127.0.0.1',
	10129,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.130',
	'127.0.0.1',
	10130,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.131',
	'127.0.0.1',
	10131,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.132',
	'127.0.0.1',
	10132,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.133',
	'127.0.0.1',
	10133,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.134',
	'127.0.0.1',
	10134,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.135',
	'127.0.0.1',
	10135,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.136',
	'127.0.0.1',
	10136,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.137',
	'127.0.0.1',
	10137,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.138',
	'127.0.0.1',
	10138,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.139',
	'127.0.0.1',
	10139,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.140',
	'127.0.0.1',
	10140,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.141',
	'127.0.0.1',
	10141,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.142',
	'127.0.0.1',
	10142,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.143',
	'127.0.0.1',
	10143,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.144',
	'127.0.0.1',
	10144,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.145',
	'127.0.0.1',
	10145,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.146',
	'127.0.0.1',
	10146,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.147',
	'127.0.0.1',
	10147,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.148',
	'127.0.0.1',
	10148,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.149',
	'127.0.0.1',
	10149,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.150',
	'127.0.0.1',
	10150,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.151',
	'127.0.0.1',
	10151,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.152',
	'127.0.0.1',
	10152,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.153',
	'127.0.0.1',
	10153,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.154',
	'127.0.0.1',
	10154,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.155',
	'127.0.0.1',
	10155,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.156',
	'127.0.0.1',
	10156,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.157',
	'127.0.0.1',
	10157,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.158',
	'127.0.0.1',
	10158,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.159',
	'127.0.0.1',
	10159,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.160',
	'127.0.0.1',
	10160,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.161',
	'127.0.0.1',
	10161,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.162',
	'127.0.0.1',
	10162,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.163',
	'127.0.0.1',
	10163,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.164',
	'127.0.0.1',
	10164,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.165',
	'127.0.0.1',
	10165,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.166',
	'127.0.0.1',
	10166,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.167',
	'127.0.0.1',
	10167,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.168',
	'127.0.0.1',
	10168,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.169',
	'127.0.0.1',
	10169,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.170',
	'127.0.0.1',
	10170,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.171',
	'127.0.0.1',
	10171,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.172',
	'127.0.0.1',
	10172,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.173',
	'127.0.0.1',
	10173,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.174',
	'127.0.0.1',
	10174,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.175',
	'127.0.0.1',
	10175,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.176',
	'127.0.0.1',
	10176,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.177',
	'127.0.0.1',
	10177,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.178',
	'127.0.0.1',
	10178,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.179',
	'127.0.0.1',
	10179,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.180',
	'127.0.0.1',
	10180,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.181',
	'127.0.0.1',
	10181,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.182',
	'127.0.0.1',
	10182,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.183',
	'127.0.0.1',
	10183,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.184',
	'127.0.0.1',
	10184,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.185',
	'127.0.0.1',
	10185,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.186',
	'127.0.0.1',
	10186,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.187',
	'127.0.0.1',
	10187,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.188',
	'127.0.0.1',
	10188,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.189',
	'127.0.0.1',
	10189,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.190',
	'127.0.0.1',
	10190,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.191',
	'127.0.0.1',
	10191,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.192',
	'127.0.0.1',
	10192,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.193',
	'127.0.0.1',
	10193,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.194',
	'127.0.0.1',
	10194,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.195',
	'127.0.0.1',
	10195,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.196',
	'127.0.0.1',
	10196,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.197',
	'127.0.0.1',
	10197,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.198',
	'127.0.0.1',
	10198,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.199',
	'127.0.0.1',
	10199,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.200',
	'127.0.0.1',
	10200,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.201',
	'127.0.0.1',
	10201,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.202',
	'127.0.0.1',
	10202,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.203',
	'127.0.0.1',
	10203,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.204',
	'127.0.0.1',
	10204,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.205',
	'127.0.0.1',
	10205,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.206',
	'127.0.0.1',
	10206,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.207',
	'127.0.0.1',
	10207,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.208',
	'127.0.0.1',
	10208,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.209',
	'127.0.0.1',
	10209,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.210',
	'127.0.0.1',
	10210,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.211',
	'127.0.0.1',
	10211,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.212',
	'127.0.0.1',
	10212,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.213',
	'127.0.0.1',
	10213,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.214',
	'127.0.0.1',
	10214,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.215',
	'127.0.0.1',
	10215,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.216',
	'127.0.0.1',
	10216,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.217',
	'127.0.0.1',
	10217,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.218',
	'127.0.0.1',
	10218,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.219',
	'127.0.0.1',
	10219,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.220',
	'127.0.0.1',
	10220,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.221',
	'127.0.0.1',
	10221,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.222',
	'127.0.0.1',
	10222,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.223',
	'127.0.0.1',
	10223,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.224',
	'127.0.0.1',
	10224,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.225',
	'127.0.0.1',
	10225,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.226',
	'127.0.0.1',
	10226,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.227',
	'127.0.0.1',
	10227,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.228',
	'127.0.0.1',
	10228,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.229',
	'127.0.0.1',
	10229,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.230',
	'127.0.0.1',
	10230,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.231',
	'127.0.0.1',
	10231,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.232',
	'127.0.0.1',
	10232,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.233',
	'127.0.0.1',
	10233,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.234',
	'127.0.0.1',
	10234,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.235',
	'127.0.0.1',
	10235,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.236',
	'127.0.0.1',
	10236,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.237',
	'127.0.0.1',
	10237,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.238',
	'127.0.0.1',
	10238,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.239',
	'127.0.0.1',
	10239,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.240',
	'127.0.0.1',
	10240,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.241',
	'127.0.0.1',
	10241,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.242',
	'127.0.0.1',
	10242,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.243',
	'127.0.0.1',
	10243,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.244',
	'127.0.0.1',
	10244,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.245',
	'127.0.0.1',
	10245,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.246',
	'127.0.0.1',
	10246,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.247',
	'127.0.0.1',
	10247,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.248',
	'127.0.0.1',
	10248,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.249',
	'127.0.0.1',
	10249,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.250',
	'127.0.0.1',
	10250,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.251',
	'127.0.0.1',
	10251,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.252',
	'127.0.0.1',
	10252,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.253',
	'127.0.0.1',
	10253,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.254',
	'127.0.0.1',
	10254,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.255',
	'127.0.0.1',
	10255,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.256',
	'127.0.0.1',
	10256,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.257',
	'127.0.0.1',
	10257,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.258',
	'127.0.0.1',
	10258,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.259',
	'127.0.0.1',
	10259,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.260',
	'127.0.0.1',
	10260,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.261',
	'127.0.0.1',
	10261,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.262',
	'127.0.0.1',
	10262,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.263',
	'127.0.0.1',
	10263,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.264',
	'127.0.0.1',
	10264,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.265',
	'127.0.0.1',
	10265,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.266',
	'127.0.0.1',
	10266,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.267',
	'127.0.0.1',
	10267,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.268',
	'127.0.0.1',
	10268,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.269',
	'127.0.0.1',
	10269,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.270',
	'127.0.0.1',
	10270,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.271',
	'127.0.0.1',
	10271,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.272',
	'127.0.0.1',
	10272,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.273',
	'127.0.0.1',
	10273,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.274',
	'127.0.0.1',
	10274,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.275',
	'127.0.0.1',
	10275,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.276',
	'127.0.0.1',
	10276,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.277',
	'127.0.0.1',
	10277,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.278',
	'127.0.0.1',
	10278,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.279',
	'127.0.0.1',
	10279,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.280',
	'127.0.0.1',
	10280,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.281',
	'127.0.0.1',
	10281,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.282',
	'127.0.0.1',
	10282,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.283',
	'127.0.0.1',
	10283,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.284',
	'127.0.0.1',
	10284,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.285',
	'127.0.0.1',
	10285,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.286',
	'127.0.0.1',
	10286,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.287',
	'127.0.0.1',
	10287,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.288',
	'127.0.0.1',
	10288,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.289',
	'127.0.0.1',
	10289,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.290',
	'127.0.0.1',
	10290,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.291',
	'127.0.0.1',
	10291,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.292',
	'127.0.0.1',
	10292,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.293',
	'127.0.0.1',
	10293,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.294',
	'127.0.0.1',
	10294,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.295',
	'127.0.0.1',
	10295,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.296',
	'127.0.0.1',
	10296,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.297',
	'127.0.0.1',
	10297,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.298',
	'127.0.0.1',
	10298,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.299',
	'127.0.0.1',
	10299,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.300',
	'127.0.0.1',
	10300,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.301',
	'127.0.0.1',
	10301,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.302',
	'127.0.0.1',
	10302,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.303',
	'127.0.0.1',
	10303,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.304',
	'127.0.0.1',
	10304,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.305',
	'127.0.0.1',
	10305,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.306',
	'127.0.0.1',
	10306,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.307',
	'127.0.0.1',
	10307,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.308',
	'127.0.0.1',
	10308,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.309',
	'127.0.0.1',
	10309,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.310',
	'127.0.0.1',
	10310,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.311',
	'127.0.0.1',
	10311,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.312',
	'127.0.0.1',
	10312,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.313',
	'127.0.0.1',
	10313,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.314',
	'127.0.0.1',
	10314,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.315',
	'127.0.0.1',
	10315,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.316',
	'127.0.0.1',
	10316,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.317',
	'127.0.0.1',
	10317,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.318',
	'127.0.0.1',
	10318,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.319',
	'127.0.0.1',
	10319,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.320',
	'127.0.0.1',
	10320,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.321',
	'127.0.0.1',
	10321,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.322',
	'127.0.0.1',
	10322,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.323',
	'127.0.0.1',
	10323,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.324',
	'127.0.0.1',
	10324,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.325',
	'127.0.0.1',
	10325,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.326',
	'127.0.0.1',
	10326,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.327',
	'127.0.0.1',
	10327,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.328',
	'127.0.0.1',
	10328,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.329',
	'127.0.0.1',
	10329,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.330',
	'127.0.0.1',
	10330,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.331',
	'127.0.0.1',
	10331,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.332',
	'127.0.0.1',
	10332,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.333',
	'127.0.0.1',
	10333,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.334',
	'127.0.0.1',
	10334,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.335',
	'127.0.0.1',
	10335,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.336',
	'127.0.0.1',
	10336,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.337',
	'127.0.0.1',
	10337,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.338',
	'127.0.0.1',
	10338,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.339',
	'127.0.0.1',
	10339,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.340',
	'127.0.0.1',
	10340,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.341',
	'127.0.0.1',
	10341,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.342',
	'127.0.0.1',
	10342,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.343',
	'127.0.0.1',
	10343,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.344',
	'127.0.0.1',
	10344,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.345',
	'127.0.0.1',
	10345,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.346',
	'127.0.0.1',
	10346,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.347',
	'127.0.0.1',
	10347,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.348',
	'127.0.0.1',
	10348,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.349',
	'127.0.0.1',
	10349,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.350',
	'127.0.0.1',
	10350,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.351',
	'127.0.0.1',
	10351,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.352',
	'127.0.0.1',
	10352,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.353',
	'127.0.0.1',
	10353,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.354',
	'127.0.0.1',
	10354,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.355',
	'127.0.0.1',
	10355,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.356',
	'127.0.0.1',
	10356,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.357',
	'127.0.0.1',
	10357,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.358',
	'127.0.0.1',
	10358,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.359',
	'127.0.0.1',
	10359,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.360',
	'127.0.0.1',
	10360,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.361',
	'127.0.0.1',
	10361,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.362',
	'127.0.0.1',
	10362,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.363',
	'127.0.0.1',
	10363,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.364',
	'127.0.0.1',
	10364,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.365',
	'127.0.0.1',
	10365,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.366',
	'127.0.0.1',
	10366,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.367',
	'127.0.0.1',
	10367,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.368',
	'127.0.0.1',
	10368,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.369',
	'127.0.0.1',
	10369,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.370',
	'127.0.0.1',
	10370,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.371',
	'127.0.0.1',
	10371,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.372',
	'127.0.0.1',
	10372,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.373',
	'127.0.0.1',
	10373,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.374',
	'127.0.0.1',
	10374,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.375',
	'127.0.0.1',
	10375,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.376',
	'127.0.0.1',
	10376,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.377',
	'127.0.0.1',
	10377,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.378',
	'127.0.0.1',
	10378,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.379',
	'127.0.0.1',
	10379,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.380',
	'127.0.0.1',
	10380,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.381',
	'127.0.0.1',
	10381,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.382',
	'127.0.0.1',
	10382,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.383',
	'127.0.0.1',
	10383,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.384',
	'127.0.0.1',
	10384,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.385',
	'127.0.0.1',
	10385,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.386',
	'127.0.0.1',
	10386,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.387',
	'127.0.0.1',
	10387,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.388',
	'127.0.0.1',
	10388,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.389',
	'127.0.0.1',
	10389,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.390',
	'127.0.0.1',
	10390,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.391',
	'127.0.0.1',
	10391,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.392',
	'127.0.0.1',
	10392,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.393',
	'127.0.0.1',
	10393,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.394',
	'127.0.0.1',
	10394,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.395',
	'127.0.0.1',
	10395,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.396',
	'127.0.0.1',
	10396,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.397',
	'127.0.0.1',
	10397,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.398',
	'127.0.0.1',
	10398,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.399',
	'127.0.0.1',
	10399,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.400',
	'127.0.0.1',
	10400,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.401',
	'127.0.0.1',
	10401,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.402',
	'127.0.0.1',
	10402,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.403',
	'127.0.0.1',
	10403,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.404',
	'127.0.0.1',
	10404,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.405',
	'127.0.0.1',
	10405,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.406',
	'127.0.0.1',
	10406,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.407',
	'127.0.0.1',
	10407,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.408',
	'127.0.0.1',
	10408,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.409',
	'127.0.0.1',
	10409,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.410',
	'127.0.0.1',
	10410,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.411',
	'127.0.0.1',
	10411,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.412',
	'127.0.0.1',
	10412,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.413',
	'127.0.0.1',
	10413,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.414',
	'127.0.0.1',
	10414,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.415',
	'127.0.0.1',
	10415,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.416',
	'127.0.0.1',
	10416,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.417',
	'127.0.0.1',
	10417,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.418',
	'127.0.0.1',
	10418,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.419',
	'127.0.0.1',
	10419,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.420',
	'127.0.0.1',
	10420,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.421',
	'127.0.0.1',
	10421,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.422',
	'127.0.0.1',
	10422,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.423',
	'127.0.0.1',
	10423,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.424',
	'127.0.0.1',
	10424,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.425',
	'127.0.0.1',
	10425,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.426',
	'127.0.0.1',
	10426,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.427',
	'127.0.0.1',
	10427,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.428',
	'127.0.0.1',
	10428,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.429',
	'127.0.0.1',
	10429,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.430',
	'127.0.0.1',
	10430,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.431',
	'127.0.0.1',
	10431,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.432',
	'127.0.0.1',
	10432,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.433',
	'127.0.0.1',
	10433,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.434',
	'127.0.0.1',
	10434,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.435',
	'127.0.0.1',
	10435,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.436',
	'127.0.0.1',
	10436,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.437',
	'127.0.0.1',
	10437,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.438',
	'127.0.0.1',
	10438,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.439',
	'127.0.0.1',
	10439,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.440',
	'127.0.0.1',
	10440,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.441',
	'127.0.0.1',
	10441,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.442',
	'127.0.0.1',
	10442,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.443',
	'127.0.0.1',
	10443,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.444',
	'127.0.0.1',
	10444,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.445',
	'127.0.0.1',
	10445,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.446',
	'127.0.0.1',
	10446,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.447',
	'127.0.0.1',
	10447,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.448',
	'127.0.0.1',
	10448,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.449',
	'127.0.0.1',
	10449,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.450',
	'127.0.0.1',
	10450,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.451',
	'127.0.0.1',
	10451,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.452',
	'127.0.0.1',
	10452,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.453',
	'127.0.0.1',
	10453,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.454',
	'127.0.0.1',
	10454,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.455',
	'127.0.0.1',
	10455,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.456',
	'127.0.0.1',
	10456,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.457',
	'127.0.0.1',
	10457,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.458',
	'127.0.0.1',
	10458,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.459',
	'127.0.0.1',
	10459,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.460',
	'127.0.0.1',
	10460,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.461',
	'127.0.0.1',
	10461,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.462',
	'127.0.0.1',
	10462,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.463',
	'127.0.0.1',
	10463,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.464',
	'127.0.0.1',
	10464,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.465',
	'127.0.0.1',
	10465,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.466',
	'127.0.0.1',
	10466,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.467',
	'127.0.0.1',
	10467,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.468',
	'127.0.0.1',
	10468,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.469',
	'127.0.0.1',
	10469,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.470',
	'127.0.0.1',
	10470,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.471',
	'127.0.0.1',
	10471,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.472',
	'127.0.0.1',
	10472,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.473',
	'127.0.0.1',
	10473,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.474',
	'127.0.0.1',
	10474,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.475',
	'127.0.0.1',
	10475,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.476',
	'127.0.0.1',
	10476,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.477',
	'127.0.0.1',
	10477,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.478',
	'127.0.0.1',
	10478,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.479',
	'127.0.0.1',
	10479,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.480',
	'127.0.0.1',
	10480,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.481',
	'127.0.0.1',
	10481,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.482',
	'127.0.0.1',
	10482,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.483',
	'127.0.0.1',
	10483,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.484',
	'127.0.0.1',
	10484,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.485',
	'127.0.0.1',
	10485,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.486',
	'127.0.0.1',
	10486,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.487',
	'127.0.0.1',
	10487,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.488',
	'127.0.0.1',
	10488,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.489',
	'127.0.0.1',
	10489,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.490',
	'127.0.0.1',
	10490,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.491',
	'127.0.0.1',
	10491,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.492',
	'127.0.0.1',
	10492,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.493',
	'127.0.0.1',
	10493,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.494',
	'127.0.0.1',
	10494,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.495',
	'127.0.0.1',
	10495,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.496',
	'127.0.0.1',
	10496,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.497',
	'127.0.0.1',
	10497,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.498',
	'127.0.0.1',
	10498,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.499',
	'127.0.0.1',
	10499,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.500',
	'127.0.0.1',
	10500,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.501',
	'127.0.0.1',
	10501,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.502',
	'127.0.0.1',
	10502,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.503',
	'127.0.0.1',
	10503,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.504',
	'127.0.0.1',
	10504,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.505',
	'127.0.0.1',
	10505,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.506',
	'127.0.0.1',
	10506,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.507',
	'127.0.0.1',
	10507,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.508',
	'127.0.0.1',
	10508,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.509',
	'127.0.0.1',
	10509,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.510',
	'127.0.0.1',
	10510,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.511',
	'127.0.0.1',
	10511,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.512',
	'127.0.0.1',
	10512,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.513',
	'127.0.0.1',
	10513,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.514',
	'127.0.0.1',
	10514,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.515',
	'127.0.0.1',
	10515,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.516',
	'127.0.0.1',
	10516,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.517',
	'127.0.0.1',
	10517,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.518',
	'127.0.0.1',
	10518,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.519',
	'127.0.0.1',
	10519,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.520',
	'127.0.0.1',
	10520,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.521',
	'127.0.0.1',
	10521,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.522',
	'127.0.0.1',
	10522,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.523',
	'127.0.0.1',
	10523,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.524',
	'127.0.0.1',
	10524,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.525',
	'127.0.0.1',
	10525,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.526',
	'127.0.0.1',
	10526,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.527',
	'127.0.0.1',
	10527,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.528',
	'127.0.0.1',
	10528,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.529',
	'127.0.0.1',
	10529,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.530',
	'127.0.0.1',
	10530,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.531',
	'127.0.0.1',
	10531,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.532',
	'127.0.0.1',
	10532,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.533',
	'127.0.0.1',
	10533,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.534',
	'127.0.0.1',
	10534,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.535',
	'127.0.0.1',
	10535,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.536',
	'127.0.0.1',
	10536,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.537',
	'127.0.0.1',
	10537,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.538',
	'127.0.0.1',
	10538,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.539',
	'127.0.0.1',
	10539,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.540',
	'127.0.0.1',
	10540,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.541',
	'127.0.0.1',
	10541,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.542',
	'127.0.0.1',
	10542,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.543',
	'127.0.0.1',
	10543,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.544',
	'127.0.0.1',
	10544,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.545',
	'127.0.0.1',
	10545,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.546',
	'127.0.0.1',
	10546,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.547',
	'127.0.0.1',
	10547,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.548',
	'127.0.0.1',
	10548,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.549',
	'127.0.0.1',
	10549,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.550',
	'127.0.0.1',
	10550,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.551',
	'127.0.0.1',
	10551,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.552',
	'127.0.0.1',
	10552,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.553',
	'127.0.0.1',
	10553,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.554',
	'127.0.0.1',
	10554,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.555',
	'127.0.0.1',
	10555,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.556',
	'127.0.0.1',
	10556,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.557',
	'127.0.0.1',
	10557,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.558',
	'127.0.0.1',
	10558,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.559',
	'127.0.0.1',
	10559,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.560',
	'127.0.0.1',
	10560,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.561',
	'127.0.0.1',
	10561,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.562',
	'127.0.0.1',
	10562,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.563',
	'127.0.0.1',
	10563,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.564',
	'127.0.0.1',
	10564,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.565',
	'127.0.0.1',
	10565,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.566',
	'127.0.0.1',
	10566,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.567',
	'127.0.0.1',
	10567,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.568',
	'127.0.0.1',
	10568,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.569',
	'127.0.0.1',
	10569,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.570',
	'127.0.0.1',
	10570,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.571',
	'127.0.0.1',
	10571,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.572',
	'127.0.0.1',
	10572,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.573',
	'127.0.0.1',
	10573,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.574',
	'127.0.0.1',
	10574,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.575',
	'127.0.0.1',
	10575,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.576',
	'127.0.0.1',
	10576,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.577',
	'127.0.0.1',
	10577,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.578',
	'127.0.0.1',
	10578,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.579',
	'127.0.0.1',
	10579,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.580',
	'127.0.0.1',
	10580,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.581',
	'127.0.0.1',
	10581,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.582',
	'127.0.0.1',
	10582,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.583',
	'127.0.0.1',
	10583,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.584',
	'127.0.0.1',
	10584,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.585',
	'127.0.0.1',
	10585,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.586',
	'127.0.0.1',
	10586,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.587',
	'127.0.0.1',
	10587,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.588',
	'127.0.0.1',
	10588,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.589',
	'127.0.0.1',
	10589,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.590',
	'127.0.0.1',
	10590,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.591',
	'127.0.0.1',
	10591,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.592',
	'127.0.0.1',
	10592,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.593',
	'127.0.0.1',
	10593,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.594',
	'127.0.0.1',
	10594,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.595',
	'127.0.0.1',
	10595,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.596',
	'127.0.0.1',
	10596,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.597',
	'127.0.0.1',
	10597,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.598',
	'127.0.0.1',
	10598,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.599',
	'127.0.0.1',
	10599,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.600',
	'127.0.0.1',
	10600,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.601',
	'127.0.0.1',
	10601,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.602',
	'127.0.0.1',
	10602,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.603',
	'127.0.0.1',
	10603,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.604',
	'127.0.0.1',
	10604,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.605',
	'127.0.0.1',
	10605,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.606',
	'127.0.0.1',
	10606,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.607',
	'127.0.0.1',
	10607,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.608',
	'127.0.0.1',
	10608,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.609',
	'127.0.0.1',
	10609,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.610',
	'127.0.0.1',
	10610,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.611',
	'127.0.0.1',
	10611,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.612',
	'127.0.0.1',
	10612,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.613',
	'127.0.0.1',
	10613,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.614',
	'127.0.0.1',
	10614,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.615',
	'127.0.0.1',
	10615,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.616',
	'127.0.0.1',
	10616,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.617',
	'127.0.0.1',
	10617,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.618',
	'127.0.0.1',
	10618,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.619',
	'127.0.0.1',
	10619,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.620',
	'127.0.0.1',
	10620,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.621',
	'127.0.0.1',
	10621,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.622',
	'127.0.0.1',
	10622,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.623',
	'127.0.0.1',
	10623,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.624',
	'127.0.0.1',
	10624,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.625',
	'127.0.0.1',
	10625,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.626',
	'127.0.0.1',
	10626,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.627',
	'127.0.0.1',
	10627,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.628',
	'127.0.0.1',
	10628,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.629',
	'127.0.0.1',
	10629,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.630',
	'127.0.0.1',
	10630,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.631',
	'127.0.0.1',
	10631,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.632',
	'127.0.0.1',
	10632,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.633',
	'127.0.0.1',
	10633,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.634',
	'127.0.0.1',
	10634,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.635',
	'127.0.0.1',
	10635,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.636',
	'127.0.0.1',
	10636,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.637',
	'127.0.0.1',
	10637,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.638',
	'127.0.0.1',
	10638,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.639',
	'127.0.0.1',
	10639,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.640',
	'127.0.0.1',
	10640,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.641',
	'127.0.0.1',
	10641,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.642',
	'127.0.0.1',
	10642,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.643',
	'127.0.0.1',
	10643,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.644',
	'127.0.0.1',
	10644,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.645',
	'127.0.0.1',
	10645,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.646',
	'127.0.0.1',
	10646,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.647',
	'127.0.0.1',
	10647,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.648',
	'127.0.0.1',
	10648,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.649',
	'127.0.0.1',
	10649,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.650',
	'127.0.0.1',
	10650,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.651',
	'127.0.0.1',
	10651,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.652',
	'127.0.0.1',
	10652,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.653',
	'127.0.0.1',
	10653,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.654',
	'127.0.0.1',
	10654,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.655',
	'127.0.0.1',
	10655,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.656',
	'127.0.0.1',
	10656,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.657',
	'127.0.0.1',
	10657,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.658',
	'127.0.0.1',
	10658,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.659',
	'127.0.0.1',
	10659,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.660',
	'127.0.0.1',
	10660,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.661',
	'127.0.0.1',
	10661,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.662',
	'127.0.0.1',
	10662,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.663',
	'127.0.0.1',
	10663,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.664',
	'127.0.0.1',
	10664,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.665',
	'127.0.0.1',
	10665,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.666',
	'127.0.0.1',
	10666,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.667',
	'127.0.0.1',
	10667,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.668',
	'127.0.0.1',
	10668,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.669',
	'127.0.0.1',
	10669,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.670',
	'127.0.0.1',
	10670,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.671',
	'127.0.0.1',
	10671,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.672',
	'127.0.0.1',
	10672,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.673',
	'127.0.0.1',
	10673,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.674',
	'127.0.0.1',
	10674,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.675',
	'127.0.0.1',
	10675,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.676',
	'127.0.0.1',
	10676,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.677',
	'127.0.0.1',
	10677,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.678',
	'127.0.0.1',
	10678,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.679',
	'127.0.0.1',
	10679,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.680',
	'127.0.0.1',
	10680,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.681',
	'127.0.0.1',
	10681,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.682',
	'127.0.0.1',
	10682,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.683',
	'127.0.0.1',
	10683,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.684',
	'127.0.0.1',
	10684,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.685',
	'127.0.0.1',
	10685,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.686',
	'127.0.0.1',
	10686,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.687',
	'127.0.0.1',
	10687,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.688',
	'127.0.0.1',
	10688,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.689',
	'127.0.0.1',
	10689,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.690',
	'127.0.0.1',
	10690,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.691',
	'127.0.0.1',
	10691,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.692',
	'127.0.0.1',
	10692,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.693',
	'127.0.0.1',
	10693,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.694',
	'127.0.0.1',
	10694,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.695',
	'127.0.0.1',
	10695,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.696',
	'127.0.0.1',
	10696,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.697',
	'127.0.0.1',
	10697,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.698',
	'127.0.0.1',
	10698,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.699',
	'127.0.0.1',
	10699,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.700',
	'127.0.0.1',
	10700,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.701',
	'127.0.0.1',
	10701,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.702',
	'127.0.0.1',
	10702,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.703',
	'127.0.0.1',
	10703,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.704',
	'127.0.0.1',
	10704,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.705',
	'127.0.0.1',
	10705,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.706',
	'127.0.0.1',
	10706,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.707',
	'127.0.0.1',
	10707,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.708',
	'127.0.0.1',
	10708,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.709',
	'127.0.0.1',
	10709,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.710',
	'127.0.0.1',
	10710,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.711',
	'127.0.0.1',
	10711,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.712',
	'127.0.0.1',
	10712,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.713',
	'127.0.0.1',
	10713,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.714',
	'127.0.0.1',
	10714,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.715',
	'127.0.0.1',
	10715,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.716',
	'127.0.0.1',
	10716,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.717',
	'127.0.0.1',
	10717,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.718',
	'127.0.0.1',
	10718,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.719',
	'127.0.0.1',
	10719,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.720',
	'127.0.0.1',
	10720,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.721',
	'127.0.0.1',
	10721,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.722',
	'127.0.0.1',
	10722,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.723',
	'127.0.0.1',
	10723,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.724',
	'127.0.0.1',
	10724,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.725',
	'127.0.0.1',
	10725,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.726',
	'127.0.0.1',
	10726,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.727',
	'127.0.0.1',
	10727,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.728',
	'127.0.0.1',
	10728,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.729',
	'127.0.0.1',
	10729,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.730',
	'127.0.0.1',
	10730,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.731',
	'127.0.0.1',
	10731,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.732',
	'127.0.0.1',
	10732,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.733',
	'127.0.0.1',
	10733,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.734',
	'127.0.0.1',
	10734,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.735',
	'127.0.0.1',
	10735,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.736',
	'127.0.0.1',
	10736,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.737',
	'127.0.0.1',
	10737,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.738',
	'127.0.0.1',
	10738,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.739',
	'127.0.0.1',
	10739,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.740',
	'127.0.0.1',
	10740,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.741',
	'127.0.0.1',
	10741,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.742',
	'127.0.0.1',
	10742,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.743',
	'127.0.0.1',
	10743,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.744',
	'127.0.0.1',
	10744,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.745',
	'127.0.0.1',
	10745,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.746',
	'127.0.0.1',
	10746,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.747',
	'127.0.0.1',
	10747,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.748',
	'127.0.0.1',
	10748,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.749',
	'127.0.0.1',
	10749,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.750',
	'127.0.0.1',
	10750,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.751',
	'127.0.0.1',
	10751,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.752',
	'127.0.0.1',
	10752,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.753',
	'127.0.0.1',
	10753,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.754',
	'127.0.0.1',
	10754,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.755',
	'127.0.0.1',
	10755,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.756',
	'127.0.0.1',
	10756,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.757',
	'127.0.0.1',
	10757,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.758',
	'127.0.0.1',
	10758,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.759',
	'127.0.0.1',
	10759,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.760',
	'127.0.0.1',
	10760,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.761',
	'127.0.0.1',
	10761,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.762',
	'127.0.0.1',
	10762,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.763',
	'127.0.0.1',
	10763,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.764',
	'127.0.0.1',
	10764,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.765',
	'127.0.0.1',
	10765,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.766',
	'127.0.0.1',
	10766,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.767',
	'127.0.0.1',
	10767,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.768',
	'127.0.0.1',
	10768,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.769',
	'127.0.0.1',
	10769,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.770',
	'127.0.0.1',
	10770,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.771',
	'127.0.0.1',
	10771,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.772',
	'127.0.0.1',
	10772,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.773',
	'127.0.0.1',
	10773,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.774',
	'127.0.0.1',
	10774,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.775',
	'127.0.0.1',
	10775,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.776',
	'127.0.0.1',
	10776,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.777',
	'127.0.0.1',
	10777,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.778',
	'127.0.0.1',
	10778,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.779',
	'127.0.0.1',
	10779,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.780',
	'127.0.0.1',
	10780,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.781',
	'127.0.0.1',
	10781,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.782',
	'127.0.0.1',
	10782,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.783',
	'127.0.0.1',
	10783,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.784',
	'127.0.0.1',
	10784,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.785',
	'127.0.0.1',
	10785,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.786',
	'127.0.0.1',
	10786,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.787',
	'127.0.0.1',
	10787,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.788',
	'127.0.0.1',
	10788,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.789',
	'127.0.0.1',
	10789,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.790',
	'127.0.0.1',
	10790,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.791',
	'127.0.0.1',
	10791,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.792',
	'127.0.0.1',
	10792,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.793',
	'127.0.0.1',
	10793,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.794',
	'127.0.0.1',
	10794,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.795',
	'127.0.0.1',
	10795,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.796',
	'127.0.0.1',
	10796,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.797',
	'127.0.0.1',
	10797,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.798',
	'127.0.0.1',
	10798,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.799',
	'127.0.0.1',
	10799,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.800',
	'127.0.0.1',
	10800,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.801',
	'127.0.0.1',
	10801,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.802',
	'127.0.0.1',
	10802,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.803',
	'127.0.0.1',
	10803,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.804',
	'127.0.0.1',
	10804,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.805',
	'127.0.0.1',
	10805,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.806',
	'127.0.0.1',
	10806,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.807',
	'127.0.0.1',
	10807,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.808',
	'127.0.0.1',
	10808,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.809',
	'127.0.0.1',
	10809,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.810',
	'127.0.0.1',
	10810,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.811',
	'127.0.0.1',
	10811,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.812',
	'127.0.0.1',
	10812,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.813',
	'127.0.0.1',
	10813,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.814',
	'127.0.0.1',
	10814,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.815',
	'127.0.0.1',
	10815,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.816',
	'127.0.0.1',
	10816,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.817',
	'127.0.0.1',
	10817,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.818',
	'127.0.0.1',
	10818,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.819',
	'127.0.0.1',
	10819,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.820',
	'127.0.0.1',
	10820,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.821',
	'127.0.0.1',
	10821,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.822',
	'127.0.0.1',
	10822,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.823',
	'127.0.0.1',
	10823,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.824',
	'127.0.0.1',
	10824,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.825',
	'127.0.0.1',
	10825,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.826',
	'127.0.0.1',
	10826,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.827',
	'127.0.0.1',
	10827,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.828',
	'127.0.0.1',
	10828,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.829',
	'127.0.0.1',
	10829,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.830',
	'127.0.0.1',
	10830,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.831',
	'127.0.0.1',
	10831,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.832',
	'127.0.0.1',
	10832,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.833',
	'127.0.0.1',
	10833,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.834',
	'127.0.0.1',
	10834,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.835',
	'127.0.0.1',
	10835,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.836',
	'127.0.0.1',
	10836,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.837',
	'127.0.0.1',
	10837,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.838',
	'127.0.0.1',
	10838,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.839',
	'127.0.0.1',
	10839,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.840',
	'127.0.0.1',
	10840,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.841',
	'127.0.0.1',
	10841,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.842',
	'127.0.0.1',
	10842,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.843',
	'127.0.0.1',
	10843,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.844',
	'127.0.0.1',
	10844,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.845',
	'127.0.0.1',
	10845,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.846',
	'127.0.0.1',
	10846,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.847',
	'127.0.0.1',
	10847,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.848',
	'127.0.0.1',
	10848,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.849',
	'127.0.0.1',
	10849,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.850',
	'127.0.0.1',
	10850,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.851',
	'127.0.0.1',
	10851,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.852',
	'127.0.0.1',
	10852,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.853',
	'127.0.0.1',
	10853,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.854',
	'127.0.0.1',
	10854,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.855',
	'127.0.0.1',
	10855,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.856',
	'127.0.0.1',
	10856,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.857',
	'127.0.0.1',
	10857,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.858',
	'127.0.0.1',
	10858,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.859',
	'127.0.0.1',
	10859,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.860',
	'127.0.0.1',
	10860,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.861',
	'127.0.0.1',
	10861,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.862',
	'127.0.0.1',
	10862,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.863',
	'127.0.0.1',
	10863,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.864',
	'127.0.0.1',
	10864,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.865',
	'127.0.0.1',
	10865,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.866',
	'127.0.0.1',
	10866,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.867',
	'127.0.0.1',
	10867,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.868',
	'127.0.0.1',
	10868,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.869',
	'127.0.0.1',
	10869,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.870',
	'127.0.0.1',
	10870,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.871',
	'127.0.0.1',
	10871,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.872',
	'127.0.0.1',
	10872,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.873',
	'127.0.0.1',
	10873,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.874',
	'127.0.0.1',
	10874,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.875',
	'127.0.0.1',
	10875,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.876',
	'127.0.0.1',
	10876,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.877',
	'127.0.0.1',
	10877,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.878',
	'127.0.0.1',
	10878,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.879',
	'127.0.0.1',
	10879,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.880',
	'127.0.0.1',
	10880,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.881',
	'127.0.0.1',
	10881,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.882',
	'127.0.0.1',
	10882,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.883',
	'127.0.0.1',
	10883,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.884',
	'127.0.0.1',
	10884,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.885',
	'127.0.0.1',
	10885,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.886',
	'127.0.0.1',
	10886,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.887',
	'127.0.0.1',
	10887,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.888',
	'127.0.0.1',
	10888,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.889',
	'127.0.0.1',
	10889,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.890',
	'127.0.0.1',
	10890,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.891',
	'127.0.0.1',
	10891,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.892',
	'127.0.0.1',
	10892,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.893',
	'127.0.0.1',
	10893,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.894',
	'127.0.0.1',
	10894,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.895',
	'127.0.0.1',
	10895,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.896',
	'127.0.0.1',
	10896,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.897',
	'127.0.0.1',
	10897,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.898',
	'127.0.0.1',
	10898,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.899',
	'127.0.0.1',
	10899,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.900',
	'127.0.0.1',
	10900,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.901',
	'127.0.0.1',
	10901,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.902',
	'127.0.0.1',
	10902,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.903',
	'127.0.0.1',
	10903,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.904',
	'127.0.0.1',
	10904,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.905',
	'127.0.0.1',
	10905,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.906',
	'127.0.0.1',
	10906,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.907',
	'127.0.0.1',
	10907,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.908',
	'127.0.0.1',
	10908,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.909',
	'127.0.0.1',
	10909,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.910',
	'127.0.0.1',
	10910,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.911',
	'127.0.0.1',
	10911,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.912',
	'127.0.0.1',
	10912,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.913',
	'127.0.0.1',
	10913,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.914',
	'127.0.0.1',
	10914,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.915',
	'127.0.0.1',
	10915,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.916',
	'127.0.0.1',
	10916,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.917',
	'127.0.0.1',
	10917,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.918',
	'127.0.0.1',
	10918,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.919',
	'127.0.0.1',
	10919,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.920',
	'127.0.0.1',
	10920,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.921',
	'127.0.0.1',
	10921,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.922',
	'127.0.0.1',
	10922,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.923',
	'127.0.0.1',
	10923,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.924',
	'127.0.0.1',
	10924,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.925',
	'127.0.0.1',
	10925,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.926',
	'127.0.0.1',
	10926,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.927',
	'127.0.0.1',
	10927,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.928',
	'127.0.0.1',
	10928,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.929',
	'127.0.0.1',
	10929,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.930',
	'127.0.0.1',
	10930,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.931',
	'127.0.0.1',
	10931,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.932',
	'127.0.0.1',
	10932,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.933',
	'127.0.0.1',
	10933,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.934',
	'127.0.0.1',
	10934,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.935',
	'127.0.0.1',
	10935,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.936',
	'127.0.0.1',
	10936,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.937',
	'127.0.0.1',
	10937,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.938',
	'127.0.0.1',
	10938,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.939',
	'127.0.0.1',
	10939,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.940',
	'127.0.0.1',
	10940,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.941',
	'127.0.0.1',
	10941,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.942',
	'127.0.0.1',
	10942,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.943',
	'127.0.0.1',
	10943,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.944',
	'127.0.0.1',
	10944,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.945',
	'127.0.0.1',
	10945,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.946',
	'127.0.0.1',
	10946,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.947',
	'127.0.0.1',
	10947,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.948',
	'127.0.0.1',
	10948,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.949',
	'127.0.0.1',
	10949,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.950',
	'127.0.0.1',
	10950,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.951',
	'127.0.0.1',
	10951,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.952',
	'127.0.0.1',
	10952,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.953',
	'127.0.0.1',
	10953,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.954',
	'127.0.0.1',
	10954,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.955',
	'127.0.0.1',
	10955,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.956',
	'127.0.0.1',
	10956,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.957',
	'127.0.0.1',
	10957,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.958',
	'127.0.0.1',
	10958,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.959',
	'127.0.0.1',
	10959,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.960',
	'127.0.0.1',
	10960,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.961',
	'127.0.0.1',
	10961,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.962',
	'127.0.0.1',
	10962,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.963',
	'127.0.0.1',
	10963,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.964',
	'127.0.0.1',
	10964,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.965',
	'127.0.0.1',
	10965,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.966',
	'127.0.0.1',
	10966,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.967',
	'127.0.0.1',
	10967,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.968',
	'127.0.0.1',
	10968,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.969',
	'127.0.0.1',
	10969,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.970',
	'127.0.0.1',
	10970,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.971',
	'127.0.0.1',
	10971,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.972',
	'127.0.0.1',
	10972,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.973',
	'127.0.0.1',
	10973,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.974',
	'127.0.0.1',
	10974,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.975',
	'127.0.0.1',
	10975,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.976',
	'127.0.0.1',
	10976,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.977',
	'127.0.0.1',
	10977,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.978',
	'127.0.0.1',
	10978,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.979',
	'127.0.0.1',
	10979,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.980',
	'127.0.0.1',
	10980,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.981',
	'127.0.0.1',
	10981,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.982',
	'127.0.0.1',
	10982,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.983',
	'127.0.0.1',
	10983,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.984',
	'127.0.0.1',
	10984,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.985',
	'127.0.0.1',
	10985,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.986',
	'127.0.0.1',
	10986,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.987',
	'127.0.0.1',
	10987,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.988',
	'127.0.0.1',
	10988,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.989',
	'127.0.0.1',
	10989,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.990',
	'127.0.0.1',
	10990,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.991',
	'127.0.0.1',
	10991,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.992',
	'127.0.0.1',
	10992,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.993',
	'127.0.0.1',
	10993,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.994',
	'127.0.0.1',
	10994,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.995',
	'127.0.0.1',
	10995,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.996',
	'127.0.0.1',
	10996,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.997',
	'127.0.0.1',
	10997,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.998',
	'127.0.0.1',
	10998,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.999',
	'127.0.0.1',
	10999,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.0'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.100'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.200'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.300'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.400'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.500'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.600'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.700'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.800'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.900'),
	true
);
