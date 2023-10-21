
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
	'test.000',
	-43.00,
	107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.001',
	-52.00,
	-51.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.002',
	-46.00,
	-110.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.003',
	25.00,
	50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.004',
	78.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.005',
	-6.00,
	11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.006',
	-54.00,
	-143.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.007',
	83.00,
	164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.008',
	-70.00,
	26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.009',
	-30.00,
	127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.010',
	44.00,
	115.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.011',
	56.00,
	111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.012',
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
	'test.013',
	61.00,
	69.00,
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
	-141.00,
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
	-64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.016',
	-31.00,
	-50.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.017',
	-19.00,
	52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.018',
	-69.00,
	-12.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.019',
	27.00,
	-130.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.020',
	-29.00,
	-11.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.021',
	71.00,
	71.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.022',
	6.00,
	-8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.023',
	22.00,
	176.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.024',
	17.00,
	-138.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.025',
	-25.00,
	2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.026',
	19.00,
	-2.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.027',
	85.00,
	179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.028',
	68.00,
	-109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.029',
	-84.00,
	-83.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.030',
	19.00,
	177.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.031',
	39.00,
	-42.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.032',
	-29.00,
	102.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.033',
	-13.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.034',
	52.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.035',
	56.00,
	148.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.036',
	-14.00,
	64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.037',
	-28.00,
	29.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.038',
	-27.00,
	0.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.039',
	-67.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.040',
	76.00,
	21.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.041',
	-68.00,
	-157.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.042',
	-10.00,
	75.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.043',
	21.00,
	-173.00,
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
	53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.045',
	84.00,
	-4.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.046',
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
	'test.047',
	-81.00,
	73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.048',
	-51.00,
	47.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.049',
	-9.00,
	8.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.050',
	-68.00,
	-48.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.051',
	28.00,
	-179.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.052',
	81.00,
	167.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.053',
	65.00,
	-55.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.054',
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
	'test.055',
	-52.00,
	-116.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.056',
	65.00,
	-111.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.057',
	9.00,
	125.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.058',
	-84.00,
	147.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.059',
	24.00,
	-112.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.060',
	-43.00,
	-84.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.061',
	64.00,
	66.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.062',
	63.00,
	46.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.063',
	15.00,
	107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.064',
	43.00,
	170.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.065',
	30.00,
	131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.066',
	76.00,
	-34.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.067',
	-50.00,
	107.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.068',
	71.00,
	-68.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.069',
	75.00,
	-95.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.070',
	71.00,
	-64.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.071',
	-72.00,
	-172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.072',
	-64.00,
	127.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.073',
	-81.00,
	-172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.074',
	-6.00,
	-100.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.075',
	90.00,
	36.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.076',
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
	'test.077',
	-26.00,
	-164.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.078',
	-66.00,
	-103.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.079',
	-58.00,
	-25.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.080',
	84.00,
	56.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.081',
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
	'test.082',
	-26.00,
	88.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.083',
	41.00,
	128.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.084',
	63.00,
	-28.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.085',
	-74.00,
	150.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.086',
	-26.00,
	-61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.087',
	-24.00,
	109.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.088',
	-89.00,
	172.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.089',
	35.00,
	15.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.090',
	-10.00,
	146.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.091',
	-41.00,
	-73.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.092',
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
	'test.093',
	36.00,
	61.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.094',
	-37.00,
	97.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.095',
	-56.00,
	131.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.096',
	-27.00,
	52.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.097',
	-72.00,
	53.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.098',
	-49.00,
	-26.00,
	(select seller_id from sellers where seller_code = 'test')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.099',
	54.00,
	33.00,
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
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
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.099')
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
