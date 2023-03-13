
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
