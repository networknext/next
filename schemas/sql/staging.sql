
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
	'5Vr+VZdUXckgQwHdPRftc/8IUWDL7ZftvBOzE/+Zpp+PIjSU0Kxmwg==',
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
	-44.00,
	-7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.001',
	79.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.002',
	-74.00,
	174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.003',
	-56.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.004',
	-29.00,
	98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.005',
	58.00,
	73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.006',
	-68.00,
	-145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.007',
	-30.00,
	69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.008',
	87.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.009',
	70.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.010',
	-12.00,
	-120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.011',
	11.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.012',
	-70.00,
	16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.013',
	-10.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.014',
	-53.00,
	112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.015',
	-13.00,
	84.00,
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
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.017',
	90.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.018',
	-6.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.019',
	-20.00,
	158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.020',
	62.00,
	-166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.021',
	-33.00,
	-17.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.022',
	-56.00,
	44.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.023',
	-26.00,
	127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.024',
	88.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.025',
	52.00,
	-68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.026',
	21.00,
	158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.027',
	-34.00,
	-73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.028',
	15.00,
	104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.029',
	-64.00,
	159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.030',
	-10.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.031',
	47.00,
	148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.032',
	-57.00,
	126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.033',
	-14.00,
	69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.034',
	25.00,
	169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.035',
	42.00,
	-167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.036',
	-25.00,
	75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.037',
	-57.00,
	22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.038',
	51.00,
	36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.039',
	-33.00,
	180.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.040',
	28.00,
	126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.041',
	1.00,
	79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.042',
	-16.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.043',
	-66.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.044',
	-17.00,
	-79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.045',
	24.00,
	-33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.046',
	10.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.047',
	-26.00,
	177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.048',
	39.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.049',
	63.00,
	-114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.050',
	68.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.051',
	-74.00,
	111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.052',
	57.00,
	148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.053',
	13.00,
	-139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.054',
	-3.00,
	-28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.055',
	4.00,
	130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.056',
	-76.00,
	-7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.057',
	-48.00,
	80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.058',
	66.00,
	177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.059',
	48.00,
	-29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.060',
	23.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.061',
	-70.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.062',
	10.00,
	126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.063',
	63.00,
	-71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.064',
	74.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.065',
	-30.00,
	-151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.066',
	-23.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.067',
	-85.00,
	-35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.068',
	-36.00,
	-88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.069',
	14.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.070',
	63.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.071',
	-81.00,
	-168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.072',
	-31.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.073',
	-23.00,
	-162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.074',
	69.00,
	-81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.075',
	23.00,
	180.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.076',
	-61.00,
	-58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.077',
	56.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.078',
	83.00,
	-180.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.079',
	82.00,
	-159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.080',
	-9.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.081',
	-4.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.082',
	-48.00,
	37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.083',
	-21.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.084',
	-71.00,
	-114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.085',
	0.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.086',
	62.00,
	150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.087',
	-58.00,
	-62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.088',
	-59.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.089',
	4.00,
	101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.090',
	75.00,
	-27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.091',
	-5.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.092',
	11.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.093',
	42.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.094',
	64.00,
	19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.095',
	-45.00,
	-49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.096',
	57.00,
	-75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.097',
	32.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.098',
	-44.00,
	-33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.099',
	-65.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.100',
	57.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.101',
	-75.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.102',
	-75.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.103',
	86.00,
	117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.104',
	-21.00,
	-13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.105',
	-65.00,
	10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.106',
	-85.00,
	119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.107',
	55.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.108',
	-39.00,
	175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.109',
	31.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.110',
	-34.00,
	94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.111',
	-37.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.112',
	-11.00,
	-47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.113',
	4.00,
	-125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.114',
	-20.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.115',
	-25.00,
	126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.116',
	-5.00,
	-75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.117',
	23.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.118',
	-70.00,
	-10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.119',
	28.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.120',
	-71.00,
	53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.121',
	16.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.122',
	-29.00,
	53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.123',
	-36.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.124',
	-27.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.125',
	14.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.126',
	1.00,
	-176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.127',
	35.00,
	-126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.128',
	24.00,
	-173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.129',
	76.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.130',
	-3.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.131',
	33.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.132',
	-18.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.133',
	57.00,
	179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.134',
	-74.00,
	-170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.135',
	-86.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.136',
	-79.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.137',
	-87.00,
	-79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.138',
	-68.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.139',
	75.00,
	117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.140',
	-37.00,
	-15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.141',
	-90.00,
	106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.142',
	74.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.143',
	-88.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.144',
	-33.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.145',
	-7.00,
	-40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.146',
	14.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.147',
	-40.00,
	160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.148',
	-34.00,
	155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.149',
	54.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.150',
	83.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.151',
	-38.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.152',
	87.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.153',
	53.00,
	-68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.154',
	32.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.155',
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
	'test.156',
	-54.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.157',
	56.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.158',
	89.00,
	74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.159',
	-40.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.160',
	71.00,
	-91.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.161',
	26.00,
	23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.162',
	9.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.163',
	-63.00,
	152.00,
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
	40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.165',
	4.00,
	79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.166',
	6.00,
	88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.167',
	-34.00,
	-99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.168',
	-45.00,
	-52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.169',
	-11.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.170',
	-66.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.171',
	-76.00,
	64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.172',
	45.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.173',
	83.00,
	-18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.174',
	-6.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.175',
	-49.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.176',
	71.00,
	57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.177',
	29.00,
	77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.178',
	72.00,
	-19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.179',
	-3.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.180',
	-50.00,
	-101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.181',
	-56.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.182',
	-26.00,
	-178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.183',
	60.00,
	-134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.184',
	81.00,
	-142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.185',
	-49.00,
	3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.186',
	3.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.187',
	-58.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.188',
	-63.00,
	-76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.189',
	37.00,
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
	-84.00,
	32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.191',
	73.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.192',
	-30.00,
	162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.193',
	-78.00,
	-34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.194',
	43.00,
	72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.195',
	-42.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.196',
	23.00,
	15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.197',
	-16.00,
	143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.198',
	-45.00,
	-22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.199',
	-6.00,
	-32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.200',
	-20.00,
	-125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.201',
	27.00,
	57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.202',
	0.00,
	86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.203',
	-20.00,
	-11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.204',
	31.00,
	176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.205',
	-30.00,
	-177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.206',
	-9.00,
	49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.207',
	-14.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.208',
	-61.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.209',
	41.00,
	155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.210',
	-42.00,
	-98.00,
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
	174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.212',
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
	'test.213',
	-79.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.214',
	86.00,
	84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.215',
	-24.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.216',
	60.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.217',
	-13.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.218',
	28.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.219',
	48.00,
	-47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.220',
	-48.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.221',
	-72.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.222',
	-28.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.223',
	-54.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.224',
	-64.00,
	98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.225',
	-62.00,
	-147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.226',
	-35.00,
	98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.227',
	49.00,
	-87.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.228',
	-13.00,
	-5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.229',
	13.00,
	79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.230',
	69.00,
	176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.231',
	37.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.232',
	4.00,
	-148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.233',
	42.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.234',
	-56.00,
	42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.235',
	14.00,
	161.00,
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
	-101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.237',
	14.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.238',
	-83.00,
	-173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.239',
	-83.00,
	-147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.240',
	20.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.241',
	67.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.242',
	54.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.243',
	49.00,
	-141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.244',
	46.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.245',
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
	'test.246',
	-14.00,
	51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.247',
	1.00,
	2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.248',
	29.00,
	-107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.249',
	-36.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.250',
	-57.00,
	-136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.251',
	55.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.252',
	-14.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.253',
	-22.00,
	-73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.254',
	-71.00,
	-87.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.255',
	-82.00,
	-19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.256',
	-85.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.257',
	45.00,
	-101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.258',
	-83.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.259',
	3.00,
	145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.260',
	9.00,
	162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.261',
	-35.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.262',
	-6.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.263',
	-49.00,
	130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.264',
	-80.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.265',
	-77.00,
	114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.266',
	78.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.267',
	17.00,
	78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.268',
	-59.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.269',
	-79.00,
	-31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.270',
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
	'test.271',
	31.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.272',
	-15.00,
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
	-86.00,
	37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.274',
	-81.00,
	-56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.275',
	-50.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.276',
	-28.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.277',
	-63.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.278',
	33.00,
	-71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.279',
	-14.00,
	-118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.280',
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
	'test.281',
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
	'test.282',
	60.00,
	90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.283',
	49.00,
	170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.284',
	20.00,
	-71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.285',
	73.00,
	1.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.286',
	-59.00,
	-176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.287',
	-9.00,
	94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.288',
	-3.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.289',
	19.00,
	44.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.290',
	-5.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.291',
	-26.00,
	-177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.292',
	-58.00,
	72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.293',
	0.00,
	-134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.294',
	-60.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.295',
	47.00,
	-62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.296',
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
	'test.297',
	-70.00,
	63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.298',
	59.00,
	-175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.299',
	-56.00,
	-178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.300',
	-59.00,
	40.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.301',
	-59.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.302',
	17.00,
	-122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.303',
	-34.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.304',
	-89.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.305',
	86.00,
	111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.306',
	-55.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.307',
	49.00,
	63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.308',
	17.00,
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
	65.00,
	-9.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.310',
	-11.00,
	73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.311',
	-47.00,
	8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.312',
	-29.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.313',
	-5.00,
	86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.314',
	52.00,
	-28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.315',
	16.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.316',
	79.00,
	-145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.317',
	-65.00,
	72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.318',
	-32.00,
	35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.319',
	-25.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.320',
	35.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.321',
	5.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.322',
	82.00,
	-90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.323',
	-3.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.324',
	8.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.325',
	69.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.326',
	-39.00,
	84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.327',
	13.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.328',
	-52.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.329',
	55.00,
	-54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.330',
	68.00,
	-65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.331',
	33.00,
	-13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.332',
	52.00,
	72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.333',
	42.00,
	104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.334',
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
	'test.335',
	-71.00,
	-179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.336',
	73.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.337',
	82.00,
	-166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.338',
	39.00,
	30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.339',
	-40.00,
	-171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.340',
	53.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.341',
	-29.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.342',
	62.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.343',
	38.00,
	-162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.344',
	-88.00,
	3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.345',
	2.00,
	-24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.346',
	42.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.347',
	31.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.348',
	-33.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.349',
	76.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.350',
	-83.00,
	136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.351',
	27.00,
	-66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.352',
	67.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.353',
	65.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.354',
	0.00,
	-53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.355',
	-9.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.356',
	-30.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.357',
	42.00,
	-7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.358',
	43.00,
	44.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.359',
	-8.00,
	-124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.360',
	-16.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.361',
	-13.00,
	-107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.362',
	60.00,
	-104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.363',
	-87.00,
	47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.364',
	-81.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.365',
	-74.00,
	-24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.366',
	-18.00,
	-59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.367',
	-46.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.368',
	-2.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.369',
	-36.00,
	64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.370',
	-15.00,
	97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.371',
	-36.00,
	-99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.372',
	-30.00,
	-128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.373',
	-73.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.374',
	45.00,
	28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.375',
	81.00,
	-122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.376',
	-45.00,
	-1.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.377',
	32.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.378',
	18.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.379',
	76.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.380',
	21.00,
	38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.381',
	-33.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.382',
	-13.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.383',
	0.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.384',
	88.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.385',
	-36.00,
	-86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.386',
	-70.00,
	154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.387',
	21.00,
	117.00,
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
	161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.389',
	-28.00,
	141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.390',
	-19.00,
	61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.391',
	27.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.392',
	50.00,
	-174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.393',
	-70.00,
	-177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.394',
	-14.00,
	111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.395',
	48.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.396',
	-53.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.397',
	-19.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.398',
	-63.00,
	-30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.399',
	-78.00,
	-65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.400',
	83.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.401',
	-52.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.402',
	76.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.403',
	-41.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.404',
	-25.00,
	158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.405',
	68.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.406',
	-42.00,
	-28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.407',
	60.00,
	149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.408',
	-59.00,
	163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.409',
	-77.00,
	80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.410',
	25.00,
	-41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.411',
	-21.00,
	108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.412',
	48.00,
	-11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.413',
	6.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.414',
	-26.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.415',
	74.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.416',
	31.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.417',
	26.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.418',
	-49.00,
	-101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.419',
	60.00,
	176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.420',
	71.00,
	174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.421',
	50.00,
	-75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.422',
	10.00,
	4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.423',
	21.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.424',
	-8.00,
	163.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.425',
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
	'test.426',
	64.00,
	102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.427',
	-54.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.428',
	27.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.429',
	-90.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.430',
	-63.00,
	18.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.431',
	42.00,
	137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.432',
	11.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.433',
	69.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.434',
	78.00,
	75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.435',
	8.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.436',
	58.00,
	-170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.437',
	-86.00,
	-74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.438',
	15.00,
	-91.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.439',
	-3.00,
	117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.440',
	-35.00,
	137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.441',
	81.00,
	33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.442',
	-46.00,
	-131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.443',
	28.00,
	-119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.444',
	-54.00,
	154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.445',
	-77.00,
	53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.446',
	-28.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.447',
	-86.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.448',
	52.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.449',
	66.00,
	131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.450',
	31.00,
	23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.451',
	-69.00,
	-34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.452',
	-84.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.453',
	-84.00,
	58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.454',
	12.00,
	-37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.455',
	37.00,
	79.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.456',
	45.00,
	-62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.457',
	48.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.458',
	69.00,
	-41.00,
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
	-119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.460',
	-38.00,
	148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.461',
	-46.00,
	112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.462',
	32.00,
	108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.463',
	83.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.464',
	14.00,
	-140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.465',
	34.00,
	113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.466',
	50.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.467',
	-69.00,
	-117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.468',
	-12.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.469',
	-55.00,
	-7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.470',
	47.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.471',
	-2.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.472',
	-71.00,
	170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.473',
	4.00,
	-174.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.474',
	-52.00,
	90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.475',
	54.00,
	36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.476',
	59.00,
	-110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.477',
	46.00,
	168.00,
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
	-27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.479',
	62.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.480',
	70.00,
	-113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.481',
	36.00,
	-58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.482',
	-33.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.483',
	86.00,
	92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.484',
	-10.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.485',
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
	'test.486',
	84.00,
	-170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.487',
	-63.00,
	108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.488',
	57.00,
	-119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.489',
	-36.00,
	68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.490',
	-44.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.491',
	54.00,
	-19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.492',
	-87.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.493',
	75.00,
	58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.494',
	-77.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.495',
	6.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.496',
	57.00,
	-144.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.497',
	78.00,
	-69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.498',
	6.00,
	-14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.499',
	69.00,
	80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.500',
	-48.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.501',
	2.00,
	-166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.502',
	49.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.503',
	71.00,
	-60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.504',
	-38.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.505',
	21.00,
	151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.506',
	-78.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.507',
	61.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.508',
	1.00,
	-71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.509',
	67.00,
	38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.510',
	-87.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.511',
	-55.00,
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.512',
	42.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.513',
	-64.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.514',
	-80.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.515',
	28.00,
	-34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.516',
	81.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.517',
	72.00,
	-133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.518',
	-49.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.519',
	-66.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.520',
	2.00,
	-160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.521',
	-20.00,
	106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.522',
	-2.00,
	137.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.523',
	2.00,
	-10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.524',
	-54.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.525',
	-14.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.526',
	60.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.527',
	-71.00,
	-3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.528',
	-42.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.529',
	-28.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.530',
	-48.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.531',
	39.00,
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
	4.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.533',
	5.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.534',
	-61.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.535',
	50.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.536',
	16.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.537',
	8.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.538',
	-32.00,
	124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.539',
	-86.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.540',
	12.00,
	-73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.541',
	15.00,
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.542',
	-32.00,
	110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.543',
	-37.00,
	115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.544',
	39.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.545',
	-77.00,
	117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.546',
	-69.00,
	-49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.547',
	7.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.548',
	-37.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.549',
	-67.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.550',
	65.00,
	150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.551',
	-61.00,
	-67.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.552',
	-34.00,
	36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.553',
	58.00,
	159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.554',
	-90.00,
	150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.555',
	45.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.556',
	-25.00,
	54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.557',
	-20.00,
	179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.558',
	89.00,
	23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.559',
	8.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.560',
	89.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.561',
	-26.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.562',
	86.00,
	-160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.563',
	-47.00,
	123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.564',
	-26.00,
	153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.565',
	61.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.566',
	51.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.567',
	0.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.568',
	86.00,
	95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.569',
	-54.00,
	166.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.570',
	34.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.571',
	54.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.572',
	-67.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.573',
	36.00,
	3.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.574',
	24.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.575',
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
	'test.576',
	15.00,
	-151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.577',
	-17.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.578',
	-30.00,
	-126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.579',
	54.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.580',
	-42.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.581',
	60.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.582',
	-51.00,
	-54.00,
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
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.584',
	-53.00,
	57.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.585',
	-34.00,
	-101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.586',
	-8.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.587',
	-74.00,
	54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.588',
	-82.00,
	12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.589',
	6.00,
	49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.590',
	-9.00,
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
	52.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.592',
	-21.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.593',
	17.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.594',
	-6.00,
	97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.595',
	-42.00,
	-16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.596',
	-59.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.597',
	0.00,
	124.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.598',
	-9.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.599',
	47.00,
	162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.600',
	-76.00,
	-56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.601',
	25.00,
	-17.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.602',
	-58.00,
	56.00,
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
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.604',
	64.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.605',
	38.00,
	59.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.606',
	18.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.607',
	-5.00,
	-126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.608',
	68.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.609',
	10.00,
	179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.610',
	8.00,
	149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.611',
	27.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.612',
	-8.00,
	-72.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.613',
	-55.00,
	2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.614',
	-5.00,
	-161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.615',
	75.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.616',
	-88.00,
	-141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.617',
	-31.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.618',
	-17.00,
	69.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.619',
	33.00,
	45.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.620',
	2.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.621',
	-56.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.622',
	-10.00,
	10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.623',
	-9.00,
	-49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.624',
	78.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.625',
	-5.00,
	-122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.626',
	-9.00,
	-52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.627',
	5.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.628',
	-63.00,
	58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.629',
	-12.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.630',
	-65.00,
	117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.631',
	45.00,
	81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.632',
	36.00,
	95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.633',
	-8.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.634',
	80.00,
	-104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.635',
	-57.00,
	-47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.636',
	-72.00,
	-122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.637',
	8.00,
	-63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.638',
	88.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.639',
	68.00,
	112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.640',
	51.00,
	-73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.641',
	87.00,
	-154.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.642',
	-13.00,
	-136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.643',
	3.00,
	33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.644',
	59.00,
	22.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.645',
	-17.00,
	-102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.646',
	-40.00,
	151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.647',
	47.00,
	-97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.648',
	-63.00,
	167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.649',
	-60.00,
	-140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.650',
	-78.00,
	-35.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.651',
	83.00,
	170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.652',
	-47.00,
	-48.00,
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
	28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.654',
	-67.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.655',
	-14.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.656',
	-35.00,
	177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.657',
	-7.00,
	-38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.658',
	-80.00,
	107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.659',
	38.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.660',
	-5.00,
	136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.661',
	84.00,
	-90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.662',
	-71.00,
	-125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.663',
	11.00,
	27.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.664',
	68.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.665',
	60.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.666',
	82.00,
	-175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.667',
	76.00,
	-110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.668',
	-71.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.669',
	-30.00,
	-91.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.670',
	51.00,
	-14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.671',
	-19.00,
	-32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.672',
	83.00,
	-29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.673',
	-49.00,
	62.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.674',
	-12.00,
	-11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.675',
	-47.00,
	-140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.676',
	-18.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.677',
	90.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.678',
	71.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.679',
	-27.00,
	-96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.680',
	-84.00,
	-84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.681',
	-75.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.682',
	-41.00,
	93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.683',
	82.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.684',
	28.00,
	-107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.685',
	86.00,
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.686',
	82.00,
	82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.687',
	-30.00,
	160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.688',
	-32.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.689',
	-13.00,
	19.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.690',
	-62.00,
	-93.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.691',
	-75.00,
	24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.692',
	-69.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.693',
	-69.00,
	-43.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.694',
	-35.00,
	-177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.695',
	-35.00,
	156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.696',
	25.00,
	-159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.697',
	75.00,
	-47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.698',
	60.00,
	-47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.699',
	80.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.700',
	-72.00,
	-122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.701',
	-25.00,
	15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.702',
	-33.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.703',
	7.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.704',
	-87.00,
	32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.705',
	-39.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.706',
	76.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.707',
	39.00,
	115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.708',
	30.00,
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
	-30.00,
	42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.710',
	32.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.711',
	-45.00,
	44.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.712',
	-11.00,
	-132.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.713',
	-7.00,
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.714',
	90.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.715',
	36.00,
	68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.716',
	-55.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.717',
	-62.00,
	-60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.718',
	71.00,
	-147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.719',
	4.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.720',
	54.00,
	-171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.721',
	-60.00,
	-30.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.722',
	55.00,
	-46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.723',
	-59.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.724',
	11.00,
	177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.725',
	-7.00,
	96.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.726',
	-55.00,
	-54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.727',
	85.00,
	104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.728',
	9.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.729',
	60.00,
	-141.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.730',
	30.00,
	-80.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.731',
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
	'test.732',
	-45.00,
	-56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.733',
	-78.00,
	-70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.734',
	58.00,
	159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.735',
	48.00,
	-118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.736',
	-40.00,
	104.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.737',
	35.00,
	159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.738',
	-90.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.739',
	78.00,
	6.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.740',
	-84.00,
	-49.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.741',
	28.00,
	54.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.742',
	38.00,
	-176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.743',
	35.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.744',
	52.00,
	-76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.745',
	-34.00,
	-159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.746',
	81.00,
	-33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.747',
	-27.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.748',
	38.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.749',
	33.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.750',
	-29.00,
	-145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.751',
	9.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.752',
	79.00,
	36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.753',
	-59.00,
	-33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.754',
	-63.00,
	-171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.755',
	-51.00,
	111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.756',
	73.00,
	-88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.757',
	9.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.758',
	12.00,
	-81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.759',
	28.00,
	-13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.760',
	-71.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.761',
	4.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.762',
	11.00,
	160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.763',
	-5.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.764',
	88.00,
	101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.765',
	63.00,
	139.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.766',
	-77.00,
	-127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.767',
	65.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.768',
	33.00,
	149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.769',
	-36.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.770',
	35.00,
	-138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.771',
	-83.00,
	5.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.772',
	26.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.773',
	-67.00,
	-68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.774',
	-83.00,
	-36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.775',
	-56.00,
	-175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.776',
	67.00,
	68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.777',
	9.00,
	-177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.778',
	-84.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.779',
	76.00,
	73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.780',
	53.00,
	-150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.781',
	-20.00,
	20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.782',
	47.00,
	-10.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.783',
	-82.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.784',
	-25.00,
	-47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.785',
	-54.00,
	-135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.786',
	-33.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.787',
	-26.00,
	-89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.788',
	26.00,
	165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.789',
	45.00,
	100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.790',
	85.00,
	-131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.791',
	2.00,
	31.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.792',
	46.00,
	-23.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.793',
	44.00,
	78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.794',
	89.00,
	101.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.795',
	-18.00,
	-176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.796',
	2.00,
	21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.797',
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
	'test.798',
	74.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.799',
	-50.00,
	65.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.800',
	-5.00,
	58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.801',
	-14.00,
	85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.802',
	26.00,
	16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.803',
	30.00,
	159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.804',
	-54.00,
	-159.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.805',
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
	'test.806',
	48.00,
	179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.807',
	3.00,
	42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.808',
	38.00,
	34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.809',
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
	'test.810',
	46.00,
	160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.811',
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
	'test.812',
	-33.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.813',
	23.00,
	16.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.814',
	9.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.815',
	69.00,
	61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.816',
	79.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.817',
	40.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.818',
	50.00,
	-126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.819',
	-23.00,
	-153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.820',
	-55.00,
	106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.821',
	-77.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.822',
	-67.00,
	-102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.823',
	-48.00,
	-151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.824',
	-57.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.825',
	-85.00,
	103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.826',
	-2.00,
	-168.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.827',
	87.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.828',
	21.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.829',
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
	'test.830',
	-45.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.831',
	88.00,
	145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.832',
	11.00,
	76.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.833',
	62.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.834',
	88.00,
	-14.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.835',
	43.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.836',
	82.00,
	-56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.837',
	48.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.838',
	37.00,
	39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.839',
	-25.00,
	-171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.840',
	-89.00,
	122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.841',
	37.00,
	-42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.842',
	-32.00,
	-107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.843',
	59.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.844',
	-81.00,
	-165.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.845',
	-38.00,
	32.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.846',
	-82.00,
	99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.847',
	53.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.848',
	-78.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.849',
	-64.00,
	86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.850',
	-76.00,
	-21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.851',
	-79.00,
	-160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.852',
	-49.00,
	71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.853',
	-70.00,
	150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.854',
	51.00,
	-58.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.855',
	-62.00,
	116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.856',
	16.00,
	0.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.857',
	-33.00,
	-1.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.858',
	-10.00,
	176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.859',
	35.00,
	-123.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.860',
	36.00,
	-39.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.861',
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
	'test.862',
	3.00,
	-125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.863',
	-70.00,
	-36.00,
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
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.865',
	-46.00,
	-122.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.866',
	-34.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.867',
	56.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.868',
	65.00,
	133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.869',
	67.00,
	-52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.870',
	14.00,
	158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.871',
	19.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.872',
	84.00,
	63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.873',
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
	'test.874',
	75.00,
	-99.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.875',
	32.00,
	178.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.876',
	34.00,
	-24.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.877',
	-17.00,
	-125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.878',
	64.00,
	136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.879',
	60.00,
	160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.880',
	57.00,
	-41.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.881',
	43.00,
	83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.882',
	50.00,
	-95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.883',
	-3.00,
	-33.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.884',
	11.00,
	-1.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.885',
	22.00,
	-117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.886',
	52.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.887',
	-60.00,
	-158.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.888',
	61.00,
	97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.889',
	81.00,
	90.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.890',
	88.00,
	-105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.891',
	22.00,
	78.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.892',
	-56.00,
	15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.893',
	49.00,
	-64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.894',
	-36.00,
	140.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.895',
	48.00,
	89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.896',
	88.00,
	105.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.897',
	72.00,
	-173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.898',
	-82.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.899',
	87.00,
	-152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.900',
	72.00,
	-161.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.901',
	64.00,
	-64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.902',
	-89.00,
	74.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.903',
	56.00,
	-77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.904',
	-38.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.905',
	-3.00,
	-43.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.906',
	54.00,
	-169.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.907',
	55.00,
	120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.908',
	-20.00,
	-66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.909',
	87.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.910',
	57.00,
	157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.911',
	-65.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.912',
	25.00,
	25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.913',
	71.00,
	-98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.914',
	56.00,
	-155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.915',
	84.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.916',
	86.00,
	129.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.917',
	7.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.918',
	-19.00,
	135.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.919',
	32.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.920',
	25.00,
	-56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.921',
	49.00,
	7.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.922',
	-37.00,
	-98.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.923',
	47.00,
	126.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.924',
	48.00,
	120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.925',
	24.00,
	149.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.926',
	54.00,
	167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.927',
	-30.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.928',
	79.00,
	-136.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.929',
	10.00,
	134.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.930',
	-28.00,
	-92.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.931',
	20.00,
	77.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.932',
	0.00,
	-94.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.933',
	-36.00,
	-117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.934',
	14.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.935',
	-2.00,
	173.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.936',
	-38.00,
	-153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.937',
	-46.00,
	119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.938',
	31.00,
	-120.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.939',
	-62.00,
	-108.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.940',
	-89.00,
	145.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.941',
	75.00,
	153.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.942',
	-67.00,
	-148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.943',
	12.00,
	-86.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.944',
	6.00,
	138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.945',
	-42.00,
	46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.946',
	-68.00,
	152.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.947',
	6.00,
	-133.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.948',
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
	'test.949',
	7.00,
	-91.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.950',
	-35.00,
	63.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.951',
	-88.00,
	121.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.952',
	-35.00,
	-162.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.953',
	-89.00,
	-68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.954',
	-25.00,
	-171.00,
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
	150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.956',
	39.00,
	-156.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.957',
	60.00,
	-170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.958',
	-47.00,
	28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.959',
	29.00,
	37.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.960',
	57.00,
	-66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.961',
	14.00,
	-81.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.962',
	33.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.963',
	-3.00,
	110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.964',
	-1.00,
	-175.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.965',
	60.00,
	87.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.966',
	-81.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.967',
	66.00,
	-68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.968',
	-1.00,
	-160.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.969',
	-43.00,
	-82.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.970',
	3.00,
	-180.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.971',
	-85.00,
	151.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.972',
	65.00,
	-20.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.973',
	12.00,
	-29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.974',
	69.00,
	51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.975',
	-77.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.976',
	42.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.977',
	3.00,
	-117.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.978',
	51.00,
	119.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.979',
	-13.00,
	171.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.980',
	80.00,
	13.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.981',
	7.00,
	102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.982',
	41.00,
	-114.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.983',
	-25.00,
	60.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.984',
	-68.00,
	-85.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.985',
	55.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.986',
	-15.00,
	177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.987',
	-65.00,
	106.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.988',
	36.00,
	118.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.989',
	-72.00,
	-127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.990',
	29.00,
	113.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.991',
	60.00,
	-155.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.992',
	45.00,
	89.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.993',
	50.00,
	0.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.994',
	-56.00,
	-142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.995',
	60.00,
	70.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.996',
	-2.00,
	-172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.997',
	-26.00,
	142.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.998',
	90.00,
	38.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.999',
	-18.00,
	-21.00,
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
	'DkmtxHmjox9lLbEHExYjvJHdA0Q6VUV0MaHOtwudPWk=',
	'QMgmGF+Cii7RxGB99wTLmrMvYGNgcKphHJvyqyTCHzE=',
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
