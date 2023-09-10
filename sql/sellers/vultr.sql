
-- vultr datacenters

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.amsterdam',
   'ams',
   2.367600,
   4.904100,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.atlanta',
   'atl',
   33.748798,
   -84.387703,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.bangalore',
   'blr',
   12.971600,
   77.594597,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.mumbai',
   'bom',
   19.076000,
   72.877701,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.paris',
   'cdg',
   48.856602,
   2.352200,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.delhi',
   'del',
   28.704100,
   77.102501,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.dallas',
   'dfw',
   32.776699,
   -96.796997,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.newyork',
   'ewr',
   40.712799,
   -74.005997,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.frankfurt',
   'fra',
   50.110901,
   8.682100,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.honolulu',
   'hnl',
   21.309900,
   -157.858093,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.seoul',
   'icn',
   37.566502,
   126.977997,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.osaka',
   'itm',
   34.693699,
   135.502304,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.johannesburg',
   'jnb',
   -26.204100,
   28.047300,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.losangeles',
   'lax',
   34.052200,
   -118.243698,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.london',
   'lhr',
   51.507198,
   -0.127600,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.madrid',
   'mad',
   40.416801,
   -3.703800,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.melbourne',
   'mel',
   -37.813599,
   144.963104,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.mexico',
   'mex',
   19.432600,
   -99.133202,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.miami',
   'mia',
   25.761700,
   -80.191803,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.tokyo',
   'nrt',
   35.676201,
   139.650299,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.chicago',
   'ord',
   41.878101,
   -87.629799,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.saopaulo',
   'sao',
   -23.555799,
   -46.639599,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.santiago',
   'scl',
   -33.448898,
   -70.669296,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.seattle',
   'sea',
   47.606201,
   -122.332100,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.singapore',
   'sgp',
   1.352100,
   103.819801,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.siliconvalley',
   'sjc',
   37.338699,
   -121.885300,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.stockholm',
   'sto',
   59.329300,
   18.068600,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.sydney',
   'syd',
   -33.868801,
   151.209305,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.warsaw',
   'waw',
   52.229698,
   21.012199,
   (select seller_id from sellers where seller_name = 'vultr')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'vultr.toronto',
   'yto',
   43.653198,
   -79.383202,
   (select seller_id from sellers where seller_name = 'vultr')
);
