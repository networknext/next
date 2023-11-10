
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
	'AN+VWuqgAQfd5d+QeT6apNf+Gbb9rqFBMtk5M+GoMvwS1Eqz764X/A==',
	(select route_shader_id from route_shaders where route_shader_name = 'test')
);

INSERT INTO sellers(seller_name, seller_code) VALUES('Test', 'test');

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.000',
	5.00,
	24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.001',
	81.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.002',
	60.00,
	-32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.003',
	-42.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.004',
	-17.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.005',
	28.00,
	120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.006',
	82.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.007',
	-25.00,
	-76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.008',
	75.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.009',
	-20.00,
	-75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.010',
	88.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.011',
	-1.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.012',
	-26.00,
	-133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.013',
	-55.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.014',
	-82.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.015',
	-3.00,
	-119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.016',
	-50.00,
	-138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.017',
	57.00,
	61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.018',
	64.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.019',
	39.00,
	-40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.020',
	-87.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.021',
	-74.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.022',
	38.00,
	102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.023',
	86.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.024',
	-20.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.025',
	-11.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.026',
	13.00,
	37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.027',
	35.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.028',
	-6.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.029',
	-72.00,
	-81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.030',
	-3.00,
	-72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.031',
	-46.00,
	-102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.032',
	50.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.033',
	-70.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.034',
	-24.00,
	-6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.035',
	64.00,
	-110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.036',
	16.00,
	155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.037',
	-6.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.038',
	-72.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.039',
	-31.00,
	28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.040',
	-72.00,
	99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.041',
	6.00,
	-137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.042',
	26.00,
	34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.043',
	3.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.044',
	65.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.045',
	-22.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.046',
	-14.00,
	132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.047',
	28.00,
	112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.048',
	84.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.049',
	-18.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.050',
	-16.00,
	-91.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.051',
	56.00,
	-95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.052',
	82.00,
	113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.053',
	24.00,
	161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.054',
	71.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.055',
	89.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.056',
	38.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.057',
	88.00,
	128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.058',
	6.00,
	-95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.059',
	22.00,
	167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.060',
	-61.00,
	-9.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.061',
	-18.00,
	-172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.062',
	2.00,
	22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.063',
	-17.00,
	161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.064',
	0.00,
	-106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.065',
	-7.00,
	-2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.066',
	-36.00,
	24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.067',
	56.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.068',
	60.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.069',
	47.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.070',
	60.00,
	-29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.071',
	24.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.072',
	55.00,
	-2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.073',
	23.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.074',
	-78.00,
	41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.075',
	-42.00,
	78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.076',
	30.00,
	-50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.077',
	-33.00,
	-49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.078',
	-22.00,
	88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.079',
	67.00,
	-126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.080',
	-41.00,
	2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.081',
	19.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.082',
	-21.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.083',
	34.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.084',
	-68.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.085',
	-12.00,
	24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.086',
	-11.00,
	119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.087',
	-67.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.088',
	-5.00,
	-56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.089',
	67.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.090',
	69.00,
	136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.091',
	-69.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.092',
	39.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.093',
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
	'test.094',
	-58.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.095',
	-53.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.096',
	39.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.097',
	-83.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.098',
	60.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.099',
	-49.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.100',
	-19.00,
	-84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.101',
	-28.00,
	-37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.102',
	-45.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.103',
	61.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.104',
	10.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.105',
	-18.00,
	-114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.106',
	67.00,
	160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.107',
	6.00,
	77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.108',
	10.00,
	1.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.109',
	78.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.110',
	-30.00,
	-65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.111',
	78.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.112',
	87.00,
	37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.113',
	36.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.114',
	-61.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.115',
	55.00,
	-37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.116',
	-89.00,
	159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.117',
	79.00,
	-104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.118',
	7.00,
	-73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.119',
	79.00,
	84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.120',
	37.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.121',
	-66.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.122',
	-88.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.123',
	30.00,
	79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.124',
	-1.00,
	19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.125',
	77.00,
	-168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.126',
	58.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.127',
	-25.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.128',
	-2.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.129',
	-61.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.130',
	80.00,
	144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.131',
	7.00,
	-53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.132',
	-3.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.133',
	41.00,
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.134',
	-76.00,
	-19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.135',
	0.00,
	-66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.136',
	-1.00,
	178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.137',
	-5.00,
	67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.138',
	37.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.139',
	-20.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.140',
	71.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.141',
	10.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.142',
	-45.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.143',
	-31.00,
	135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.144',
	-46.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.145',
	8.00,
	37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.146',
	2.00,
	-121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.147',
	53.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.148',
	-48.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.149',
	-45.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.150',
	18.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.151',
	77.00,
	-38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.152',
	-78.00,
	131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.153',
	60.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.154',
	-67.00,
	130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.155',
	-48.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.156',
	-23.00,
	89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.157',
	-52.00,
	-125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.158',
	-52.00,
	34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.159',
	-43.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.160',
	28.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.161',
	82.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.162',
	-84.00,
	-28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.163',
	78.00,
	124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.164',
	8.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.165',
	-53.00,
	-119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.166',
	90.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.167',
	-84.00,
	-114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.168',
	-4.00,
	119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.169',
	38.00,
	68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.170',
	72.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.171',
	-37.00,
	-161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.172',
	-64.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.173',
	36.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.174',
	69.00,
	131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.175',
	-67.00,
	17.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.176',
	-57.00,
	-127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.177',
	-71.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.178',
	6.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.179',
	-49.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.180',
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
	'test.181',
	-66.00,
	-146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.182',
	-53.00,
	-65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.183',
	-28.00,
	52.00,
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
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.185',
	-79.00,
	-3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.186',
	-22.00,
	115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.187',
	45.00,
	-11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.188',
	40.00,
	68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.189',
	66.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.190',
	74.00,
	21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.191',
	63.00,
	9.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.192',
	24.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.193',
	72.00,
	-103.00,
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
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.195',
	-81.00,
	53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.196',
	53.00,
	86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.197',
	35.00,
	36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.198',
	-66.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.199',
	56.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.200',
	1.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.201',
	-84.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.202',
	-63.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.203',
	58.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.204',
	68.00,
	33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.205',
	90.00,
	-38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.206',
	-47.00,
	170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.207',
	87.00,
	-160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.208',
	-49.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.209',
	-42.00,
	-66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.210',
	-63.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.211',
	21.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.212',
	-44.00,
	15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.213',
	-86.00,
	135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.214',
	57.00,
	-5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.215',
	-32.00,
	-90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.216',
	-40.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.217',
	62.00,
	-60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.218',
	-20.00,
	167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.219',
	-22.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.220',
	70.00,
	63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.221',
	-21.00,
	-58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.222',
	-55.00,
	124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.223',
	82.00,
	-76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.224',
	-76.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.225',
	67.00,
	28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.226',
	23.00,
	74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.227',
	-49.00,
	179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.228',
	86.00,
	178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.229',
	-10.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.230',
	30.00,
	-13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.231',
	-18.00,
	-50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.232',
	-71.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.233',
	-49.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.234',
	-9.00,
	-154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.235',
	7.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.236',
	-82.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.237',
	-17.00,
	-139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.238',
	-57.00,
	78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.239',
	-73.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.240',
	-21.00,
	-11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.241',
	35.00,
	24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.242',
	66.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.243',
	65.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.244',
	87.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.245',
	77.00,
	-145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.246',
	70.00,
	67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.247',
	12.00,
	-53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.248',
	-28.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.249',
	23.00,
	119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.250',
	8.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.251',
	-45.00,
	179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.252',
	-24.00,
	98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.253',
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
	'test.254',
	-12.00,
	-14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.255',
	82.00,
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.256',
	-81.00,
	-127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.257',
	64.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.258',
	3.00,
	80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.259',
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
	'test.260',
	25.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.261',
	-80.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.262',
	-79.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.263',
	-8.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.264',
	42.00,
	-39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.265',
	-26.00,
	89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.266',
	-58.00,
	8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.267',
	-40.00,
	10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.268',
	-47.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.269',
	90.00,
	-131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.270',
	19.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.271',
	-47.00,
	23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.272',
	7.00,
	19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.273',
	79.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.274',
	-60.00,
	-42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.275',
	-84.00,
	117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.276',
	23.00,
	-104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.277',
	32.00,
	153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.278',
	43.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.279',
	42.00,
	-3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.280',
	-84.00,
	-160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.281',
	-20.00,
	-54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.282',
	34.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.283',
	85.00,
	-133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.284',
	3.00,
	143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.285',
	63.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.286',
	-2.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.287',
	-59.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.288',
	-81.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.289',
	2.00,
	-140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.290',
	-22.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.291',
	76.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.292',
	-69.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.293',
	35.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.294',
	49.00,
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.295',
	44.00,
	180.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.296',
	-75.00,
	15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.297',
	38.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.298',
	-4.00,
	-7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.299',
	-50.00,
	110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.300',
	-40.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.301',
	-50.00,
	-28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.302',
	51.00,
	159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.303',
	88.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.304',
	1.00,
	14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.305',
	-48.00,
	-10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.306',
	51.00,
	43.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.307',
	42.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.308',
	-29.00,
	77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.309',
	43.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.310',
	77.00,
	-63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.311',
	62.00,
	28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.312',
	63.00,
	-117.00,
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
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.314',
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
	'test.315',
	-59.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.316',
	-70.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.317',
	-87.00,
	-118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.318',
	88.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.319',
	-29.00,
	-42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.320',
	33.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.321',
	1.00,
	149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.322',
	-40.00,
	-10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.323',
	-54.00,
	-52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.324',
	-54.00,
	128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.325',
	-23.00,
	106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.326',
	-59.00,
	-99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.327',
	-25.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.328',
	4.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.329',
	-25.00,
	67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.330',
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
	'test.331',
	34.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.332',
	-43.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.333',
	5.00,
	-121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.334',
	42.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.335',
	-55.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.336',
	-72.00,
	80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.337',
	45.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.338',
	36.00,
	-52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.339',
	-59.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.340',
	31.00,
	87.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.341',
	8.00,
	-84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.342',
	61.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.343',
	30.00,
	-127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.344',
	84.00,
	35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.345',
	-66.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.346',
	72.00,
	19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.347',
	-35.00,
	-179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.348',
	47.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.349',
	78.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.350',
	-34.00,
	28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.351',
	65.00,
	-82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.352',
	63.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.353',
	8.00,
	98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.354',
	65.00,
	52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.355',
	90.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.356',
	33.00,
	-162.00,
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
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.358',
	-36.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.359',
	-24.00,
	-71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.360',
	-59.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.361',
	-31.00,
	-1.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.362',
	-87.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.363',
	82.00,
	68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.364',
	-57.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.365',
	41.00,
	-66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.366',
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
	'test.367',
	-12.00,
	150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.368',
	-37.00,
	177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.369',
	-19.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.370',
	85.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.371',
	-19.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.372',
	-53.00,
	-119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.373',
	-74.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.374',
	23.00,
	-75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.375',
	-33.00,
	-60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.376',
	85.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.377',
	90.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.378',
	46.00,
	-120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.379',
	-66.00,
	117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.380',
	-32.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.381',
	2.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.382',
	-23.00,
	49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.383',
	-26.00,
	52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.384',
	-50.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.385',
	-13.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.386',
	32.00,
	2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.387',
	-66.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.388',
	29.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.389',
	-37.00,
	-11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.390',
	64.00,
	-40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.391',
	44.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.392',
	38.00,
	54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.393',
	-87.00,
	-173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.394',
	-84.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.395',
	-66.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.396',
	-26.00,
	108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.397',
	-68.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.398',
	16.00,
	112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.399',
	12.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.400',
	41.00,
	64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.401',
	-68.00,
	42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.402',
	-27.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.403',
	35.00,
	91.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.404',
	55.00,
	-88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.405',
	55.00,
	-72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.406',
	-26.00,
	-106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.407',
	-10.00,
	-52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.408',
	-69.00,
	-29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.409',
	-25.00,
	-14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.410',
	36.00,
	-69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.411',
	26.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.412',
	66.00,
	59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.413',
	34.00,
	1.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.414',
	90.00,
	-27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.415',
	-64.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.416',
	-69.00,
	146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.417',
	41.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.418',
	-9.00,
	148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.419',
	-80.00,
	130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.420',
	45.00,
	78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.421',
	-66.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.422',
	42.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.423',
	-42.00,
	108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.424',
	-30.00,
	-171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.425',
	12.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.426',
	-79.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.427',
	29.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.428',
	-51.00,
	-178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.429',
	49.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.430',
	90.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.431',
	-15.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.432',
	80.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.433',
	11.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.434',
	-78.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.435',
	-45.00,
	146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.436',
	-38.00,
	-180.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.437',
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
	'test.438',
	76.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.439',
	35.00,
	-134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.440',
	-59.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.441',
	-88.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.442',
	-31.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.443',
	-56.00,
	4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.444',
	87.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.445',
	-21.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.446',
	-23.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.447',
	-6.00,
	-14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.448',
	-54.00,
	75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.449',
	63.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.450',
	-20.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.451',
	-48.00,
	-87.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.452',
	45.00,
	-173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.453',
	1.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.454',
	-80.00,
	145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.455',
	-37.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.456',
	36.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.457',
	76.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.458',
	6.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.459',
	-13.00,
	-33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.460',
	-8.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.461',
	-60.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.462',
	54.00,
	-108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.463',
	86.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.464',
	-33.00,
	15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.465',
	31.00,
	43.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.466',
	66.00,
	26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.467',
	69.00,
	-134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.468',
	-53.00,
	2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.469',
	4.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.470',
	-42.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.471',
	-5.00,
	44.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.472',
	-19.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.473',
	-88.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.474',
	-11.00,
	-19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.475',
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
	'test.476',
	79.00,
	106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.477',
	-72.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.478',
	-48.00,
	-129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.479',
	-15.00,
	144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.480',
	56.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.481',
	48.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.482',
	24.00,
	-177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.483',
	67.00,
	79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.484',
	39.00,
	-141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.485',
	-5.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.486',
	73.00,
	-156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.487',
	28.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.488',
	79.00,
	145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.489',
	-75.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.490',
	-12.00,
	95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.491',
	42.00,
	43.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.492',
	35.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.493',
	30.00,
	-171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.494',
	32.00,
	-76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.495',
	-62.00,
	-15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.496',
	-25.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.497',
	6.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.498',
	-1.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.499',
	70.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.500',
	-22.00,
	-129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.501',
	50.00,
	-177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.502',
	-37.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.503',
	-1.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.504',
	-30.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.505',
	28.00,
	-137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.506',
	67.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.507',
	-83.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.508',
	-89.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.509',
	37.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.510',
	18.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.511',
	27.00,
	-9.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.512',
	7.00,
	-79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.513',
	-45.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.514',
	-84.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.515',
	-31.00,
	-88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.516',
	46.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.517',
	61.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.518',
	8.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.519',
	82.00,
	47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.520',
	6.00,
	59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.521',
	-2.00,
	74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.522',
	49.00,
	-153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.523',
	-35.00,
	-35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.524',
	45.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.525',
	19.00,
	57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.526',
	-49.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.527',
	-51.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.528',
	-68.00,
	36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.529',
	81.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.530',
	82.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.531',
	-23.00,
	153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.532',
	-52.00,
	55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.533',
	-29.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.534',
	65.00,
	128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.535',
	54.00,
	0.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.536',
	9.00,
	-98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.537',
	-61.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.538',
	7.00,
	-72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.539',
	-55.00,
	-154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.540',
	-24.00,
	-180.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.541',
	-82.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.542',
	-16.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.543',
	-90.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.544',
	-12.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.545',
	-39.00,
	55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.546',
	-88.00,
	-139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.547',
	-38.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.548',
	-6.00,
	-81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.549',
	50.00,
	51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.550',
	68.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.551',
	-39.00,
	-19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.552',
	13.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.553',
	-45.00,
	-40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.554',
	58.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.555',
	24.00,
	-18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.556',
	-43.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.557',
	-28.00,
	124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.558',
	76.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.559',
	69.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.560',
	52.00,
	-98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.561',
	45.00,
	-136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.562',
	43.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.563',
	62.00,
	146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.564',
	-49.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.565',
	37.00,
	-140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.566',
	-61.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.567',
	-43.00,
	97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.568',
	-51.00,
	162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.569',
	39.00,
	-178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.570',
	-80.00,
	38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.571',
	-51.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.572',
	-25.00,
	179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.573',
	-81.00,
	80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.574',
	40.00,
	23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.575',
	-22.00,
	95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.576',
	88.00,
	-19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.577',
	-35.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.578',
	-43.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.579',
	80.00,
	-139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.580',
	-11.00,
	-129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.581',
	16.00,
	154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.582',
	51.00,
	-18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.583',
	-88.00,
	-58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.584',
	-89.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.585',
	-28.00,
	-140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.586',
	21.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.587',
	83.00,
	158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.588',
	77.00,
	-178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.589',
	2.00,
	40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.590',
	87.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.591',
	85.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.592',
	18.00,
	-146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.593',
	-23.00,
	4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.594',
	-55.00,
	169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.595',
	44.00,
	-159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.596',
	-74.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.597',
	43.00,
	-151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.598',
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
	'test.599',
	-47.00,
	-91.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.600',
	-18.00,
	33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.601',
	77.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.602',
	-6.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.603',
	-21.00,
	10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.604',
	26.00,
	58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.605',
	-48.00,
	-44.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.606',
	-25.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.607',
	90.00,
	-179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.608',
	37.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.609',
	90.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.610',
	86.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.611',
	5.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.612',
	50.00,
	-49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.613',
	-11.00,
	-153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.614',
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
	'test.615',
	48.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.616',
	-31.00,
	-147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.617',
	-69.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.618',
	17.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.619',
	-65.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.620',
	-23.00,
	78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.621',
	-69.00,
	-30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.622',
	59.00,
	72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.623',
	53.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.624',
	7.00,
	-171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.625',
	-16.00,
	-15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.626',
	42.00,
	108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.627',
	75.00,
	-127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.628',
	40.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.629',
	-13.00,
	130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.630',
	68.00,
	158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.631',
	-71.00,
	-91.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.632',
	-23.00,
	-129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.633',
	42.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.634',
	12.00,
	-64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.635',
	82.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.636',
	22.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.637',
	-86.00,
	-106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.638',
	-31.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.639',
	-59.00,
	-13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.640',
	-41.00,
	61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.641',
	-71.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.642',
	63.00,
	4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.643',
	-25.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.644',
	21.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.645',
	59.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.646',
	-34.00,
	-60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.647',
	-67.00,
	33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.648',
	0.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.649',
	-41.00,
	-174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.650',
	77.00,
	159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.651',
	69.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.652',
	26.00,
	117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.653',
	8.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.654',
	20.00,
	-16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.655',
	43.00,
	63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.656',
	-59.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.657',
	-85.00,
	175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.658',
	84.00,
	53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.659',
	10.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.660',
	-70.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.661',
	54.00,
	38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.662',
	-80.00,
	-172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.663',
	-75.00,
	-50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.664',
	-17.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.665',
	75.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.666',
	51.00,
	174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.667',
	37.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.668',
	37.00,
	-159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.669',
	5.00,
	38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.670',
	-31.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.671',
	51.00,
	-155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.672',
	66.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.673',
	73.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.674',
	-65.00,
	145.00,
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
	-13.00,
	153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.677',
	88.00,
	88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.678',
	48.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.679',
	11.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.680',
	17.00,
	-179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.681',
	-39.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.682',
	64.00,
	-101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.683',
	70.00,
	172.00,
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
	-50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.685',
	-44.00,
	57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.686',
	10.00,
	10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.687',
	-70.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.688',
	-16.00,
	0.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.689',
	13.00,
	-71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.690',
	-88.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.691',
	86.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.692',
	18.00,
	58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.693',
	-59.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.694',
	-68.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.695',
	-82.00,
	59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.696',
	-27.00,
	-71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.697',
	58.00,
	113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.698',
	83.00,
	-98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.699',
	-81.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.700',
	-31.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.701',
	66.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.702',
	72.00,
	-88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.703',
	-76.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.704',
	61.00,
	-72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.705',
	8.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.706',
	3.00,
	120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.707',
	-88.00,
	-13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.708',
	-26.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.709',
	-70.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.710',
	3.00,
	46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.711',
	0.00,
	55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.712',
	-76.00,
	-98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.713',
	47.00,
	-32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.714',
	-49.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.715',
	-38.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.716',
	88.00,
	34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.717',
	71.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.718',
	78.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.719',
	-20.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.720',
	-26.00,
	-136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.721',
	-44.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.722',
	-18.00,
	-74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.723',
	52.00,
	-81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.724',
	26.00,
	101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.725',
	-8.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.726',
	-57.00,
	2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.727',
	-28.00,
	37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.728',
	-7.00,
	-54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.729',
	-9.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.730',
	-59.00,
	-117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.731',
	-85.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.732',
	81.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.733',
	76.00,
	-180.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.734',
	4.00,
	-29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.735',
	-32.00,
	-57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.736',
	81.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.737',
	53.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.738',
	58.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.739',
	-72.00,
	-148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.740',
	40.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.741',
	27.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.742',
	68.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.743',
	30.00,
	160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.744',
	-37.00,
	-81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.745',
	-76.00,
	-30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.746',
	6.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.747',
	-30.00,
	-64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.748',
	5.00,
	-44.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.749',
	-77.00,
	92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.750',
	-12.00,
	-148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.751',
	-13.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.752',
	29.00,
	-50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.753',
	5.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.754',
	-41.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.755',
	-68.00,
	84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.756',
	31.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.757',
	-23.00,
	90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.758',
	-2.00,
	-35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.759',
	56.00,
	-16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.760',
	-77.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.761',
	-45.00,
	-149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.762',
	35.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.763',
	35.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.764',
	14.00,
	-34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.765',
	68.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.766',
	53.00,
	-140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.767',
	-40.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.768',
	42.00,
	108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.769',
	51.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.770',
	-78.00,
	92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.771',
	19.00,
	-60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.772',
	85.00,
	-50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.773',
	11.00,
	-14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.774',
	70.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.775',
	-8.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.776',
	90.00,
	-162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.777',
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
	'test.778',
	73.00,
	-107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.779',
	-78.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.780',
	42.00,
	130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.781',
	-39.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.782',
	16.00,
	-102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.783',
	59.00,
	154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.784',
	-42.00,
	-133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.785',
	59.00,
	4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.786',
	-51.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.787',
	36.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.788',
	-65.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.789',
	9.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.790',
	-55.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.791',
	87.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.792',
	-31.00,
	-59.00,
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
	24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.794',
	8.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.795',
	7.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.796',
	-65.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.797',
	-72.00,
	-108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.798',
	-24.00,
	150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.799',
	22.00,
	-176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.800',
	59.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.801',
	-68.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.802',
	83.00,
	1.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.803',
	-53.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.804',
	-66.00,
	112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.805',
	87.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.806',
	-26.00,
	-118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.807',
	27.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.808',
	-72.00,
	161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.809',
	-58.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.810',
	90.00,
	75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.811',
	-30.00,
	-90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.812',
	-65.00,
	-79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.813',
	50.00,
	-129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.814',
	11.00,
	17.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.815',
	29.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.816',
	49.00,
	54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.817',
	-28.00,
	145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.818',
	-58.00,
	145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.819',
	-61.00,
	-84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.820',
	59.00,
	77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.821',
	4.00,
	-84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.822',
	13.00,
	-133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.823',
	11.00,
	174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.824',
	-79.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.825',
	-1.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.826',
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
	'test.827',
	-70.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.828',
	32.00,
	-172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.829',
	68.00,
	101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.830',
	9.00,
	148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.831',
	55.00,
	-91.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.832',
	8.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.833',
	17.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.834',
	49.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.835',
	-73.00,
	-45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.836',
	-79.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.837',
	-20.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.838',
	43.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.839',
	87.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.840',
	17.00,
	-66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.841',
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
	'test.842',
	89.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.843',
	-81.00,
	59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.844',
	11.00,
	-174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.845',
	50.00,
	-127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.846',
	77.00,
	55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.847',
	-47.00,
	-84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.848',
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
	'test.849',
	-71.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.850',
	39.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.851',
	78.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.852',
	-73.00,
	-62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.853',
	-77.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.854',
	-23.00,
	-139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.855',
	60.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.856',
	-22.00,
	99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.857',
	-46.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.858',
	-78.00,
	-49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.859',
	-39.00,
	-153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.860',
	24.00,
	-127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.861',
	45.00,
	-74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.862',
	18.00,
	94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.863',
	78.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.864',
	-21.00,
	120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.865',
	-34.00,
	17.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.866',
	27.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.867',
	68.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.868',
	-13.00,
	-27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.869',
	24.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.870',
	-29.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.871',
	-42.00,
	104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.872',
	55.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.873',
	44.00,
	-129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.874',
	-76.00,
	120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.875',
	64.00,
	38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.876',
	-32.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.877',
	21.00,
	32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.878',
	-69.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.879',
	-8.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.880',
	-64.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.881',
	12.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.882',
	41.00,
	-127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.883',
	27.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.884',
	-76.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.885',
	23.00,
	-42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.886',
	1.00,
	174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.887',
	71.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.888',
	-48.00,
	22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.889',
	-63.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.890',
	-85.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.891',
	-46.00,
	178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.892',
	23.00,
	72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.893',
	-77.00,
	98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.894',
	-4.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.895',
	-74.00,
	-18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.896',
	31.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.897',
	28.00,
	-81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.898',
	-64.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.899',
	32.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.900',
	0.00,
	-166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.901',
	44.00,
	-30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.902',
	36.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.903',
	-85.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.904',
	69.00,
	-75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.905',
	25.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.906',
	57.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.907',
	43.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.908',
	-9.00,
	-95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.909',
	-84.00,
	101.00,
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
	-176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.911',
	-43.00,
	-64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.912',
	-60.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.913',
	5.00,
	-114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.914',
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
	'test.915',
	72.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.916',
	60.00,
	169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.917',
	50.00,
	-121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.918',
	22.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.919',
	-13.00,
	17.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.920',
	-4.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.921',
	-37.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.922',
	-52.00,
	110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.923',
	-29.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.924',
	-43.00,
	47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.925',
	-78.00,
	-107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.926',
	29.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.927',
	-63.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.928',
	-81.00,
	80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.929',
	-22.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.930',
	-46.00,
	69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.931',
	29.00,
	-166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.932',
	84.00,
	47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.933',
	-71.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.934',
	-17.00,
	-122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.935',
	-15.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.936',
	-6.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.937',
	-27.00,
	-14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.938',
	20.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.939',
	-49.00,
	40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.940',
	-58.00,
	-88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.941',
	62.00,
	-78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.942',
	-50.00,
	-170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.943',
	11.00,
	84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.944',
	84.00,
	-57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.945',
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
	'test.946',
	-90.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.947',
	57.00,
	-54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.948',
	60.00,
	126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.949',
	77.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.950',
	-60.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.951',
	36.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.952',
	9.00,
	-38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.953',
	64.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.954',
	4.00,
	17.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.955',
	86.00,
	-155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.956',
	-44.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.957',
	48.00,
	67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.958',
	-23.00,
	3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.959',
	5.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.960',
	-42.00,
	-13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.961',
	-29.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.962',
	-13.00,
	61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.963',
	-81.00,
	-126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.964',
	-81.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.965',
	-31.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.966',
	63.00,
	-154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.967',
	-47.00,
	-163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.968',
	-11.00,
	135.00,
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
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.970',
	-57.00,
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.971',
	-54.00,
	-176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.972',
	85.00,
	153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.973',
	2.00,
	35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.974',
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
	'test.975',
	9.00,
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.976',
	-46.00,
	-107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.977',
	39.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.978',
	74.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.979',
	7.00,
	-148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.980',
	-28.00,
	160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.981',
	-75.00,
	-32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.982',
	-18.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.983',
	25.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.984',
	83.00,
	24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.985',
	-17.00,
	168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.986',
	-83.00,
	-140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.987',
	62.00,
	9.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.988',
	-40.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.989',
	45.00,
	-115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.990',
	-19.00,
	94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.991',
	-83.00,
	-42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.992',
	78.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.993',
	-72.00,
	163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.994',
	84.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.995',
	-30.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.996',
	85.00,
	75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.997',
	52.00,
	3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.998',
	-13.00,
	-161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.999',
	-7.00,
	50.00,
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
	'test.000',
	'127.0.0.1',
	10000,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.000')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.001',
	'127.0.0.1',
	10001,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.001')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.002',
	'127.0.0.1',
	10002,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.002')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.003',
	'127.0.0.1',
	10003,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.003')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.004',
	'127.0.0.1',
	10004,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.004')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.005',
	'127.0.0.1',
	10005,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.005')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.006',
	'127.0.0.1',
	10006,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.006')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.007',
	'127.0.0.1',
	10007,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.007')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.008',
	'127.0.0.1',
	10008,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.008')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.009',
	'127.0.0.1',
	10009,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.009')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.010',
	'127.0.0.1',
	10010,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.010')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.011',
	'127.0.0.1',
	10011,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.011')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.012',
	'127.0.0.1',
	10012,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.012')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.013',
	'127.0.0.1',
	10013,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.013')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.014',
	'127.0.0.1',
	10014,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.014')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.015',
	'127.0.0.1',
	10015,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.015')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.016',
	'127.0.0.1',
	10016,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.016')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.017',
	'127.0.0.1',
	10017,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.017')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.018',
	'127.0.0.1',
	10018,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.018')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.019',
	'127.0.0.1',
	10019,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.019')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.020',
	'127.0.0.1',
	10020,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.020')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.021',
	'127.0.0.1',
	10021,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.021')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.022',
	'127.0.0.1',
	10022,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.022')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.023',
	'127.0.0.1',
	10023,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.023')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.024',
	'127.0.0.1',
	10024,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.024')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.025',
	'127.0.0.1',
	10025,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.025')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.026',
	'127.0.0.1',
	10026,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.026')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.027',
	'127.0.0.1',
	10027,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.027')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.028',
	'127.0.0.1',
	10028,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.028')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.029',
	'127.0.0.1',
	10029,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.029')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.030',
	'127.0.0.1',
	10030,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.030')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.031',
	'127.0.0.1',
	10031,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.031')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.032',
	'127.0.0.1',
	10032,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.032')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.033',
	'127.0.0.1',
	10033,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.033')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.034',
	'127.0.0.1',
	10034,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.034')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.035',
	'127.0.0.1',
	10035,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.035')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.036',
	'127.0.0.1',
	10036,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.036')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.037',
	'127.0.0.1',
	10037,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.037')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.038',
	'127.0.0.1',
	10038,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.038')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.039',
	'127.0.0.1',
	10039,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.039')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.040',
	'127.0.0.1',
	10040,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.040')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.041',
	'127.0.0.1',
	10041,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.041')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.042',
	'127.0.0.1',
	10042,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.042')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.043',
	'127.0.0.1',
	10043,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.043')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.044',
	'127.0.0.1',
	10044,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.044')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.045',
	'127.0.0.1',
	10045,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.045')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.046',
	'127.0.0.1',
	10046,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.046')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.047',
	'127.0.0.1',
	10047,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.047')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.048',
	'127.0.0.1',
	10048,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.048')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.049',
	'127.0.0.1',
	10049,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.049')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.050',
	'127.0.0.1',
	10050,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.050')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.051',
	'127.0.0.1',
	10051,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.051')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.052',
	'127.0.0.1',
	10052,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.052')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.053',
	'127.0.0.1',
	10053,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.053')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.054',
	'127.0.0.1',
	10054,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.054')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.055',
	'127.0.0.1',
	10055,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.055')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.056',
	'127.0.0.1',
	10056,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.056')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.057',
	'127.0.0.1',
	10057,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.057')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.058',
	'127.0.0.1',
	10058,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.058')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.059',
	'127.0.0.1',
	10059,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.059')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.060',
	'127.0.0.1',
	10060,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.060')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.061',
	'127.0.0.1',
	10061,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.061')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.062',
	'127.0.0.1',
	10062,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.062')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.063',
	'127.0.0.1',
	10063,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.063')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.064',
	'127.0.0.1',
	10064,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.064')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.065',
	'127.0.0.1',
	10065,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.065')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.066',
	'127.0.0.1',
	10066,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.066')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.067',
	'127.0.0.1',
	10067,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.067')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.068',
	'127.0.0.1',
	10068,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.068')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.069',
	'127.0.0.1',
	10069,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.069')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.070',
	'127.0.0.1',
	10070,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.070')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.071',
	'127.0.0.1',
	10071,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.071')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.072',
	'127.0.0.1',
	10072,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.072')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.073',
	'127.0.0.1',
	10073,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.073')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.074',
	'127.0.0.1',
	10074,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.074')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.075',
	'127.0.0.1',
	10075,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.075')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.076',
	'127.0.0.1',
	10076,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.076')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.077',
	'127.0.0.1',
	10077,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.077')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.078',
	'127.0.0.1',
	10078,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.078')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.079',
	'127.0.0.1',
	10079,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.079')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.080',
	'127.0.0.1',
	10080,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.080')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.081',
	'127.0.0.1',
	10081,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.081')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.082',
	'127.0.0.1',
	10082,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.082')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.083',
	'127.0.0.1',
	10083,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.083')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.084',
	'127.0.0.1',
	10084,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.084')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.085',
	'127.0.0.1',
	10085,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.085')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.086',
	'127.0.0.1',
	10086,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.086')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.087',
	'127.0.0.1',
	10087,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.087')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.088',
	'127.0.0.1',
	10088,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.088')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.089',
	'127.0.0.1',
	10089,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.089')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.090',
	'127.0.0.1',
	10090,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.090')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.091',
	'127.0.0.1',
	10091,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.091')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.092',
	'127.0.0.1',
	10092,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.092')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.093',
	'127.0.0.1',
	10093,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.093')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.094',
	'127.0.0.1',
	10094,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.094')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.095',
	'127.0.0.1',
	10095,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.095')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.096',
	'127.0.0.1',
	10096,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.096')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.097',
	'127.0.0.1',
	10097,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.097')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.098',
	'127.0.0.1',
	10098,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.098')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.099',
	'127.0.0.1',
	10099,
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.099')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.100')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.101')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.102')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.103')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.104')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.105')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.106')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.107')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.108')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.109')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.110')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.111')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.112')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.113')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.114')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.115')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.116')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.117')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.118')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.119')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.120')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.121')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.122')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.123')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.124')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.125')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.126')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.127')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.128')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.129')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.130')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.131')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.132')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.133')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.134')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.135')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.136')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.137')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.138')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.139')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.140')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.141')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.142')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.143')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.144')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.145')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.146')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.147')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.148')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.149')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.150')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.151')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.152')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.153')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.154')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.155')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.156')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.157')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.158')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.159')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.160')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.161')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.162')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.163')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.164')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.165')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.166')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.167')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.168')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.169')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.170')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.171')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.172')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.173')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.174')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.175')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.176')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.177')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.178')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.179')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.180')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.181')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.182')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.183')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.184')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.185')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.186')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.187')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.188')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.189')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.190')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.191')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.192')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.193')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.194')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.195')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.196')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.197')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.198')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.199')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.200')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.201')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.202')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.203')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.204')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.205')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.206')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.207')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.208')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.209')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.210')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.211')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.212')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.213')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.214')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.215')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.216')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.217')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.218')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.219')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.220')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.221')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.222')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.223')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.224')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.225')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.226')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.227')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.228')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.229')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.230')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.231')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.232')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.233')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.234')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.235')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.236')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.237')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.238')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.239')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.240')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.241')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.242')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.243')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.244')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.245')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.246')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.247')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.248')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.249')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.250')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.251')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.252')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.253')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.254')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.255')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.256')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.257')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.258')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.259')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.260')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.261')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.262')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.263')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.264')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.265')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.266')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.267')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.268')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.269')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.270')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.271')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.272')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.273')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.274')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.275')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.276')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.277')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.278')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.279')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.280')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.281')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.282')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.283')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.284')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.285')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.286')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.287')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.288')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.289')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.290')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.291')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.292')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.293')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.294')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.295')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.296')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.297')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.298')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.299')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.300')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.301')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.302')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.303')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.304')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.305')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.306')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.307')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.308')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.309')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.310')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.311')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.312')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.313')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.314')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.315')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.316')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.317')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.318')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.319')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.320')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.321')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.322')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.323')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.324')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.325')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.326')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.327')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.328')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.329')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.330')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.331')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.332')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.333')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.334')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.335')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.336')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.337')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.338')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.339')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.340')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.341')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.342')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.343')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.344')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.345')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.346')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.347')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.348')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.349')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.350')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.351')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.352')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.353')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.354')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.355')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.356')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.357')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.358')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.359')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.360')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.361')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.362')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.363')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.364')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.365')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.366')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.367')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.368')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.369')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.370')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.371')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.372')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.373')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.374')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.375')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.376')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.377')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.378')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.379')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.380')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.381')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.382')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.383')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.384')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.385')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.386')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.387')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.388')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.389')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.390')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.391')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.392')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.393')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.394')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.395')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.396')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.397')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.398')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.399')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.400')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.401')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.402')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.403')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.404')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.405')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.406')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.407')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.408')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.409')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.410')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.411')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.412')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.413')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.414')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.415')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.416')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.417')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.418')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.419')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.420')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.421')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.422')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.423')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.424')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.425')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.426')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.427')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.428')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.429')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.430')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.431')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.432')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.433')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.434')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.435')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.436')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.437')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.438')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.439')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.440')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.441')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.442')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.443')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.444')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.445')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.446')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.447')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.448')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.449')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.450')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.451')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.452')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.453')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.454')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.455')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.456')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.457')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.458')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.459')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.460')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.461')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.462')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.463')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.464')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.465')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.466')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.467')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.468')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.469')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.470')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.471')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.472')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.473')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.474')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.475')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.476')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.477')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.478')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.479')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.480')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.481')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.482')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.483')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.484')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.485')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.486')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.487')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.488')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.489')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.490')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.491')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.492')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.493')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.494')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.495')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.496')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.497')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.498')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.499')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.500')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.501')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.502')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.503')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.504')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.505')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.506')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.507')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.508')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.509')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.510')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.511')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.512')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.513')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.514')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.515')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.516')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.517')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.518')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.519')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.520')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.521')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.522')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.523')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.524')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.525')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.526')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.527')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.528')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.529')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.530')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.531')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.532')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.533')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.534')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.535')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.536')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.537')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.538')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.539')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.540')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.541')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.542')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.543')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.544')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.545')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.546')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.547')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.548')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.549')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.550')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.551')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.552')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.553')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.554')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.555')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.556')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.557')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.558')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.559')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.560')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.561')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.562')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.563')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.564')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.565')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.566')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.567')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.568')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.569')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.570')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.571')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.572')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.573')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.574')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.575')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.576')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.577')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.578')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.579')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.580')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.581')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.582')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.583')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.584')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.585')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.586')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.587')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.588')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.589')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.590')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.591')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.592')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.593')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.594')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.595')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.596')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.597')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.598')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.599')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.600')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.601')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.602')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.603')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.604')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.605')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.606')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.607')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.608')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.609')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.610')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.611')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.612')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.613')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.614')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.615')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.616')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.617')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.618')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.619')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.620')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.621')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.622')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.623')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.624')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.625')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.626')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.627')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.628')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.629')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.630')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.631')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.632')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.633')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.634')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.635')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.636')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.637')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.638')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.639')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.640')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.641')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.642')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.643')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.644')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.645')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.646')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.647')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.648')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.649')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.650')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.651')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.652')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.653')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.654')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.655')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.656')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.657')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.658')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.659')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.660')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.661')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.662')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.663')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.664')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.665')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.666')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.667')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.668')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.669')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.670')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.671')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.672')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.673')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.674')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.675')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.676')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.677')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.678')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.679')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.680')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.681')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.682')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.683')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.684')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.685')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.686')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.687')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.688')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.689')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.690')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.691')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.692')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.693')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.694')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.695')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.696')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.697')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.698')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.699')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.700')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.701')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.702')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.703')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.704')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.705')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.706')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.707')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.708')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.709')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.710')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.711')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.712')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.713')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.714')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.715')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.716')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.717')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.718')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.719')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.720')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.721')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.722')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.723')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.724')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.725')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.726')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.727')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.728')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.729')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.730')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.731')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.732')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.733')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.734')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.735')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.736')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.737')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.738')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.739')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.740')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.741')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.742')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.743')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.744')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.745')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.746')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.747')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.748')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.749')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.750')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.751')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.752')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.753')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.754')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.755')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.756')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.757')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.758')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.759')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.760')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.761')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.762')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.763')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.764')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.765')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.766')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.767')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.768')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.769')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.770')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.771')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.772')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.773')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.774')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.775')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.776')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.777')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.778')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.779')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.780')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.781')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.782')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.783')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.784')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.785')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.786')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.787')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.788')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.789')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.790')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.791')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.792')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.793')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.794')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.795')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.796')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.797')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.798')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.799')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.800')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.801')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.802')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.803')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.804')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.805')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.806')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.807')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.808')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.809')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.810')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.811')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.812')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.813')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.814')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.815')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.816')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.817')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.818')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.819')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.820')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.821')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.822')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.823')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.824')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.825')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.826')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.827')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.828')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.829')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.830')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.831')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.832')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.833')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.834')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.835')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.836')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.837')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.838')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.839')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.840')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.841')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.842')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.843')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.844')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.845')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.846')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.847')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.848')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.849')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.850')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.851')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.852')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.853')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.854')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.855')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.856')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.857')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.858')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.859')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.860')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.861')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.862')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.863')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.864')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.865')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.866')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.867')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.868')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.869')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.870')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.871')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.872')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.873')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.874')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.875')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.876')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.877')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.878')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.879')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.880')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.881')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.882')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.883')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.884')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.885')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.886')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.887')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.888')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.889')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.890')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.891')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.892')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.893')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.894')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.895')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.896')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.897')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.898')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.899')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.900')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.901')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.902')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.903')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.904')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.905')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.906')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.907')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.908')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.909')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.910')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.911')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.912')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.913')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.914')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.915')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.916')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.917')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.918')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.919')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.920')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.921')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.922')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.923')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.924')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.925')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.926')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.927')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.928')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.929')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.930')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.931')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.932')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.933')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.934')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.935')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.936')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.937')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.938')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.939')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.940')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.941')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.942')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.943')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.944')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.945')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.946')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.947')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.948')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.949')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.950')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.951')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.952')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.953')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.954')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.955')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.956')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.957')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.958')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.959')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.960')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.961')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.962')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.963')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.964')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.965')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.966')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.967')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.968')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.969')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.970')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.971')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.972')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.973')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.974')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.975')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.976')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.977')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.978')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.979')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.980')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.981')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.982')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.983')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.984')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.985')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.986')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.987')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.988')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.989')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.990')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.991')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.992')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.993')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.994')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.995')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.996')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.997')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.998')
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
	'swNl9YuThlOMQ4jDcbxQYt2uvmv08OZqrgMRzrXtriA=',
	'2h7RT4KEtPSA9z+L5iSOvAWOtb9LBDSThYO0pHzzQC8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.999')
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.000'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.001'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.002'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.003'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.004'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.005'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.006'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.007'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.008'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.009'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.010'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.011'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.012'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.013'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.014'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.015'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.016'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.017'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.018'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.019'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.020'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.021'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.022'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.023'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.024'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.025'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.026'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.027'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.028'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.029'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.030'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.031'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.032'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.033'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.034'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.035'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.036'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.037'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.038'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.039'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.040'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.041'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.042'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.043'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.044'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.045'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.046'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.047'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.048'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.049'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.050'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.051'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.052'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.053'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.054'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.055'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.056'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.057'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.058'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.059'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.060'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.061'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.062'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.063'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.064'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.065'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.066'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.067'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.068'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.069'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.070'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.071'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.072'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.073'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.074'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.075'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.076'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.077'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.078'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.079'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.080'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.081'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.082'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.083'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.084'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.085'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.086'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.087'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.088'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.089'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.090'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.091'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.092'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.093'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.094'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.095'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.096'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.097'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.098'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.099'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.100'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.101'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.102'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.103'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.104'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.105'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.106'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.107'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.108'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.109'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.110'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.111'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.112'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.113'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.114'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.115'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.116'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.117'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.118'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.119'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.120'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.121'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.122'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.123'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.124'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.125'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.126'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.127'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.128'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.129'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.130'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.131'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.132'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.133'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.134'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.135'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.136'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.137'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.138'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.139'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.140'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.141'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.142'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.143'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.144'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.145'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.146'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.147'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.148'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.149'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.150'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.151'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.152'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.153'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.154'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.155'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.156'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.157'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.158'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.159'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.160'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.161'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.162'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.163'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.164'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.165'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.166'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.167'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.168'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.169'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.170'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.171'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.172'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.173'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.174'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.175'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.176'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.177'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.178'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.179'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.180'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.181'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.182'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.183'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.184'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.185'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.186'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.187'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.188'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.189'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.190'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.191'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.192'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.193'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.194'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.195'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.196'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.197'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.198'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.199'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.200'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.201'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.202'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.203'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.204'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.205'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.206'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.207'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.208'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.209'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.210'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.211'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.212'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.213'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.214'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.215'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.216'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.217'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.218'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.219'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.220'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.221'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.222'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.223'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.224'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.225'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.226'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.227'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.228'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.229'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.230'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.231'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.232'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.233'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.234'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.235'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.236'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.237'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.238'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.239'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.240'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.241'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.242'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.243'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.244'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.245'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.246'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.247'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.248'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.249'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.250'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.251'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.252'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.253'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.254'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.255'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.256'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.257'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.258'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.259'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.260'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.261'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.262'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.263'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.264'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.265'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.266'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.267'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.268'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.269'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.270'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.271'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.272'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.273'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.274'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.275'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.276'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.277'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.278'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.279'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.280'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.281'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.282'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.283'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.284'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.285'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.286'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.287'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.288'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.289'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.290'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.291'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.292'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.293'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.294'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.295'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.296'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.297'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.298'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.299'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.300'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.301'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.302'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.303'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.304'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.305'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.306'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.307'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.308'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.309'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.310'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.311'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.312'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.313'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.314'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.315'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.316'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.317'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.318'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.319'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.320'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.321'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.322'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.323'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.324'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.325'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.326'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.327'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.328'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.329'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.330'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.331'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.332'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.333'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.334'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.335'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.336'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.337'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.338'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.339'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.340'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.341'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.342'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.343'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.344'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.345'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.346'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.347'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.348'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.349'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.350'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.351'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.352'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.353'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.354'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.355'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.356'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.357'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.358'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.359'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.360'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.361'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.362'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.363'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.364'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.365'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.366'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.367'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.368'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.369'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.370'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.371'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.372'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.373'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.374'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.375'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.376'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.377'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.378'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.379'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.380'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.381'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.382'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.383'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.384'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.385'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.386'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.387'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.388'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.389'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.390'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.391'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.392'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.393'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.394'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.395'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.396'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.397'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.398'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.399'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.400'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.401'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.402'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.403'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.404'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.405'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.406'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.407'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.408'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.409'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.410'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.411'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.412'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.413'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.414'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.415'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.416'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.417'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.418'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.419'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.420'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.421'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.422'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.423'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.424'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.425'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.426'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.427'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.428'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.429'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.430'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.431'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.432'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.433'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.434'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.435'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.436'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.437'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.438'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.439'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.440'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.441'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.442'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.443'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.444'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.445'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.446'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.447'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.448'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.449'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.450'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.451'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.452'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.453'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.454'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.455'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.456'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.457'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.458'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.459'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.460'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.461'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.462'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.463'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.464'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.465'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.466'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.467'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.468'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.469'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.470'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.471'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.472'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.473'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.474'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.475'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.476'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.477'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.478'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.479'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.480'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.481'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.482'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.483'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.484'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.485'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.486'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.487'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.488'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.489'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.490'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.491'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.492'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.493'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.494'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.495'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.496'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.497'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.498'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.499'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.500'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.501'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.502'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.503'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.504'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.505'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.506'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.507'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.508'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.509'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.510'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.511'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.512'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.513'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.514'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.515'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.516'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.517'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.518'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.519'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.520'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.521'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.522'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.523'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.524'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.525'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.526'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.527'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.528'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.529'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.530'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.531'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.532'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.533'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.534'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.535'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.536'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.537'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.538'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.539'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.540'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.541'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.542'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.543'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.544'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.545'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.546'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.547'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.548'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.549'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.550'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.551'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.552'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.553'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.554'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.555'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.556'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.557'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.558'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.559'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.560'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.561'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.562'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.563'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.564'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.565'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.566'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.567'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.568'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.569'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.570'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.571'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.572'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.573'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.574'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.575'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.576'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.577'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.578'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.579'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.580'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.581'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.582'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.583'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.584'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.585'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.586'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.587'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.588'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.589'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.590'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.591'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.592'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.593'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.594'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.595'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.596'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.597'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.598'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.599'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.600'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.601'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.602'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.603'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.604'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.605'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.606'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.607'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.608'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.609'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.610'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.611'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.612'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.613'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.614'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.615'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.616'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.617'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.618'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.619'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.620'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.621'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.622'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.623'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.624'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.625'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.626'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.627'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.628'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.629'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.630'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.631'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.632'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.633'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.634'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.635'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.636'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.637'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.638'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.639'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.640'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.641'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.642'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.643'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.644'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.645'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.646'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.647'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.648'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.649'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.650'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.651'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.652'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.653'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.654'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.655'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.656'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.657'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.658'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.659'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.660'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.661'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.662'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.663'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.664'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.665'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.666'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.667'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.668'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.669'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.670'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.671'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.672'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.673'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.674'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.675'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.676'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.677'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.678'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.679'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.680'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.681'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.682'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.683'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.684'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.685'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.686'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.687'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.688'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.689'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.690'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.691'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.692'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.693'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.694'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.695'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.696'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.697'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.698'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.699'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.700'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.701'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.702'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.703'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.704'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.705'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.706'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.707'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.708'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.709'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.710'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.711'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.712'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.713'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.714'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.715'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.716'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.717'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.718'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.719'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.720'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.721'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.722'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.723'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.724'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.725'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.726'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.727'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.728'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.729'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.730'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.731'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.732'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.733'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.734'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.735'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.736'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.737'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.738'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.739'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.740'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.741'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.742'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.743'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.744'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.745'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.746'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.747'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.748'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.749'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.750'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.751'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.752'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.753'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.754'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.755'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.756'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.757'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.758'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.759'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.760'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.761'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.762'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.763'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.764'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.765'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.766'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.767'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.768'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.769'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.770'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.771'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.772'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.773'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.774'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.775'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.776'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.777'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.778'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.779'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.780'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.781'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.782'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.783'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.784'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.785'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.786'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.787'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.788'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.789'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.790'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.791'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.792'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.793'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.794'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.795'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.796'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.797'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.798'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.799'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.800'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.801'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.802'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.803'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.804'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.805'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.806'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.807'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.808'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.809'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.810'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.811'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.812'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.813'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.814'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.815'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.816'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.817'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.818'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.819'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.820'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.821'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.822'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.823'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.824'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.825'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.826'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.827'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.828'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.829'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.830'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.831'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.832'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.833'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.834'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.835'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.836'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.837'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.838'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.839'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.840'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.841'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.842'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.843'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.844'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.845'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.846'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.847'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.848'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.849'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.850'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.851'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.852'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.853'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.854'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.855'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.856'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.857'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.858'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.859'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.860'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.861'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.862'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.863'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.864'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.865'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.866'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.867'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.868'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.869'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.870'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.871'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.872'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.873'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.874'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.875'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.876'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.877'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.878'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.879'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.880'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.881'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.882'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.883'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.884'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.885'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.886'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.887'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.888'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.889'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.890'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.891'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.892'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.893'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.894'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.895'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.896'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.897'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.898'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.899'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.900'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.901'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.902'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.903'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.904'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.905'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.906'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.907'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.908'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.909'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.910'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.911'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.912'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.913'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.914'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.915'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.916'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.917'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.918'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.919'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.920'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.921'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.922'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.923'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.924'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.925'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.926'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.927'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.928'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.929'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.930'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.931'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.932'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.933'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.934'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.935'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.936'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.937'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.938'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.939'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.940'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.941'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.942'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.943'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.944'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.945'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.946'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.947'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.948'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.949'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.950'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.951'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.952'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.953'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.954'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.955'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.956'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.957'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.958'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.959'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.960'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.961'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.962'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.963'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.964'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.965'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.966'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.967'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.968'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.969'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.970'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.971'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.972'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.973'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.974'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.975'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.976'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.977'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.978'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.979'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.980'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.981'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.982'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.983'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.984'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.985'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.986'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.987'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.988'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.989'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.990'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.991'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.992'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.993'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.994'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.995'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.996'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.997'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.998'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.999'),
	true
);
