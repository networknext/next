
-- equinix datacenters

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'equinix.atlanta',
	33.7488,
	-84.3877,
	(select seller_id from sellers where seller_name = 'equinix')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'equinix.dallas',
	32.7767,
	-96.7970,
	(select seller_id from sellers where seller_name = 'equinix')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'equinix.fremont',
	37.3387,
	-121.8853,
	(select seller_id from sellers where seller_name = 'equinix')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'equinix.newark',
	40.7357,
	-74.1724,
	(select seller_id from sellers where seller_name = 'equinix')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'equinix.toronto',
	43.6532,
	79.3832,
	(select seller_id from sellers where seller_name = 'equinix')
);
