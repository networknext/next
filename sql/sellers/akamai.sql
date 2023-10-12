
-- akamai datacenters

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.mumbai',
   'ap-west',
   19.076000,
   72.877701,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.toronto',
   'ca-central',
   43.653198,
   -79.383202,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.sydney',
   'ap-southeast',
   -33.868801,
   151.209305,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.dallas',
   'us-central',
   32.776699,
   -96.796997,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.fremont',
   'us-west',
   37.548500,
   -121.988602,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.atlanta',
   'us-southeast',
   33.748798,
   -84.387703,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.newyork',
   'us-east',
   40.712799,
   -74.005997,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.london',
   'eu-west',
   51.507198,
   -0.127600,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.singapore',
   'ap-south',
   1.352100,
   103.819801,
   (select seller_id from sellers where seller_name = 'akamai')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'akamai.frankfurt',
   'eu-central',
   50.110901,
   8.682100,
   (select seller_id from sellers where seller_name = 'akamai')
);
