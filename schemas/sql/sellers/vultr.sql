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
