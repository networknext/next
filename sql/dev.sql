
\c network_next

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

INSERT INTO buyers
(
	public_key_base64, 
	customer_id
) 
VALUES(
	'UoFYERKJnCt18mU53IsWzlEXD2pYD9yd+TiZiq9+cMF9cHG4kMwRtw==',
	(select id from customers where customer_code = 'raspberry')
);

INSERT INTO sellers(short_name) VALUES('google');
INSERT INTO sellers(short_name) VALUES('amazon');
INSERT INTO sellers(short_name) VALUES('vultr');
INSERT INTO sellers(short_name) VALUES('linode');

-- amazon datacenters

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.ohio.2',
	40.4173, 
	-82.9071,
	(select id from sellers where short_name = 'amazon')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.oregon.1',
	45.8399,
	-119.7006,
	(select id from sellers where short_name = 'amazon')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.sanjose.1',
	37.3387,
	-121.8853,
	(select id from sellers where short_name = 'amazon')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.saopaulo.1',
	-23.5558, 
	-46.6396,
	(select id from sellers where short_name = 'amazon')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.virginia.1',
	39.0438,
	-77.4874,
	(select id from sellers where short_name = 'amazon')
);

-- google datacenters

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.dallas.1',
	32.7767,
	-96.7970,
	(select id from sellers where short_name = 'google')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.iowa.1',
	41.2619,
	-95.8608,
	(select id from sellers where short_name = 'google')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.lasvegas.1',
	36.1716,
	-115.1391,
	(select id from sellers where short_name = 'google')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.losangeles.1',
	34.0522,
	118.2437,
	(select id from sellers where short_name = 'google')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.oregon.2',
	45.8399,
	-119.7006,
	(select id from sellers where short_name = 'google')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saltlakecity.1',
	40.7608,
	-111.8910,
	(select id from sellers where short_name = 'google')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.santiago.1',
	-33.4489,
	-70.6693,
	(select id from sellers where short_name = 'google')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saopaulo.1',
	-23.5558, 
	-46.6396,
	(select id from sellers where short_name = 'google')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.virginia.3',
	39.0438,
	-77.4874,
	(select id from sellers where short_name = 'google')
);

-- linode datacenters

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.atlanta',
	33.7488,
	-84.3877,
	(select id from sellers where short_name = 'linode')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.dallas',
	32.7767,
	-96.7970,
	(select id from sellers where short_name = 'linode')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.fremont',
	37.3387,
	-121.8853,
	(select id from sellers where short_name = 'linode')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.newark',
	40.7357,
	-74.1724,
	(select id from sellers where short_name = 'linode')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'linode.toronto',
	43.6532,
	79.3832,
	(select id from sellers where short_name = 'linode')
);

-- vultr datacenters 

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.atlanta',
	33.7488,
	-84.3877,
	(select id from sellers where short_name = 'vultr')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.chicago',
	41.8781,
	-87.6298,
	(select id from sellers where short_name = 'vultr')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.dallas',
	32.7767,
	-96.7970,
	(select id from sellers where short_name = 'vultr')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.honolulu',
	21.3099,
	-157.8581,
	(select id from sellers where short_name = 'vultr')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.losangeles',
	34.0522,
	118.2437,
	(select id from sellers where short_name = 'vultr')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.miami',
	25.7617,
	-80.1918,
	(select id from sellers where short_name = 'vultr')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.newyork',
	40.7128,
	-74.0060,
	(select id from sellers where short_name = 'vultr')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.seattle',
	47.6062,
	-122.3321,
	(select id from sellers where short_name = 'vultr')
);

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'vultr.siliconvalley',
	37.3387,
	-121.8853,
	(select id from sellers where short_name = 'vultr')
);

-- amazon relays

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'amazon.ohio.2',
	'3.138.73.252',
	40000,
	'3.138.73.252',
	22,
	'ubuntu',
	'ZlpZo9pD3sFPXynxtc5IV+02TrUmHYuxJc1uffyQkmI=',
	'vLKdOuM8tpbcF6ZtkeThlkYNVT7SWPd9c2eAdvFQQq0=',
	(select id from datacenters where display_name = 'amazon.ohio.2')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'amazon.oregon.1',
	'44.242.70.57',
	40000,
	'44.242.70.57',
	22,
	'ubuntu',
	'MIxnbFMdR04xFwwipYtokcymfh0+xCGCbaryAb5U9zw=',
	'la9ykstfh8f93K7+sKbIi7GQhAW9GIaGkBRs61d47tM=',
	(select id from datacenters where display_name = 'amazon.oregon.1')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'amazon.sanjose.1',
	'52.52.246.62',
	40000,
	'52.52.246.62',
	22,
	'ubuntu',
	'HgZtHcZWzyihZUunYoU6Jmh2wnoEQEM3skDOo5Q4Nyo=',
	'dxG105dIZhb8ajIMyRZKSIPBaBXQG/fsveOZR4eAivg=',
	(select id from datacenters where display_name = 'amazon.sanjose.1')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'amazon.saopaulo.1',
	'54.94.14.133',
	40000,
	'54.94.14.133',
	22,
	'ubuntu',
	'cwuB9C364MMEVeFOvo0VFH+HueZzesjyJ9FIP5gpBhw=',
	'nvtdVeYSmEFVFo1sUTg3y4C1pvhM5w3JKl5PIc/UWC0=',
	(select id from datacenters where display_name = 'amazon.saopaulo.1')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'amazon.virginia.1',
	'34.232.104.206',
	40000,
	'34.232.104.206',
	22,
	'ubuntu',
	'WvoyGHCOBSyNPamA1qQ4WlSTpnBhIWepht0utgUSPQ8=',
	'R9IfNVadwq8PPNu21VvwSROoccGwr9z7xT8VlMPTeas=',
	(select id from datacenters where display_name = 'amazon.virginia.1')
);

-- google relays

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'google.iowa.1',
	'35.226.96.92',
	40000,
	'35.226.96.92',
	22,
	'root',
	'fjv8SS5z4/YCc6a8JSrv/YdJPTWjAeSSUzpkrPk+4MA=',
	'SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=',
	(select id from datacenters where display_name = 'google.iowa.1')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'google.lasvegas.1',
	'34.125.125.84',
	40000,
	'34.125.125.84',
	22,
	'root',
	'xKAh+sLW1ghyIkHfOVzkinZZU0mjZF52a+aQ1cv9RRg=',
	'Goql8jwWRUYZpV8XtTPjXC+pDLUzrQ0zpbi8OvElHYw=',
	(select id from datacenters where display_name = 'google.lasvegas.1')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'google.losangeles.1',
	'34.94.28.121',
	40000,
	'34.94.28.121',
	22,
	'root',
	'XVnmFpOfx7DmuhZkrkHw+UpuIegRCNBDlHSlfOH/ozo=',
	'KzRn1WQHVMJftQe/UoCUqgCeCea46u4iWBN/1ADlXNg=',
	(select id from datacenters where display_name = 'google.losangeles.1')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'google.oregon.1',
	'34.168.209.101',
	40000,
	'34.168.209.101',
	22,
	'root',
	'fqbsO0Yw/mP82T0JIPblltGtf9xsLbtSqiWaF5/xv38=',
	'Hy4r9eTSq9vEeiYyaOyghdll39FZndwCzihzZQ6RVqA=',
	(select id from datacenters where display_name = 'google.oregon.2')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'google.saltlakecity.1',
	'34.106.29.193',
	40000,
	'34.106.29.193',
	22,
	'root',
	'9AftbXkssUIQfK9/zsG/KsPaONd/Uq9FeM/x5iHkGlQ=',
	'sNA5bYajFhUo561VEkpqr6KnTlEK4bDrSUyY5NsDv8E=',
	(select id from datacenters where display_name = 'google.saltlakecity.1')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'google.santiago.1',
	'34.176.85.20',
	40000,
	'34.176.85.20',
	22,
	'root',
	'E1ZkLyobOMFZPP7cbqpKcEEb79Z0ZIW/IDaUSWliOl0=',
	'opsJFrG1lO5HUZanT5+RML0qdJaFj5ws8YIOVNtRcG0=',
	(select id from datacenters where display_name = 'google.santiago.1')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'google.saopaulo.1',
	'34.151.248.241',
	40000,
	'34.151.248.241',
	22,
	'root',
	'qunlVxGncMg5b650wXgtYBmJAzetry+K9ancBayMWzw=',
	'1vpJ9L6jntr+KvqHSkZvgH9EnkVE/stS+60pfAdXEkg=',
	(select id from datacenters where display_name = 'google.saopaulo.1')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'google.virginia.3',
	'35.236.236.4',
	40000,
	'35.236.236.4',
	22,
	'root',
	'5KcEWA5Digp5hBm5TOfXtX3twEk/etE0SB8rwlIrjWQ=',
	'SCrHFjgowY4n4fEkPZnS8wvxseCUiwFGHvaCSkJItqo=',
	(select id from datacenters where display_name = 'google.virginia.3')
);

/*
INSERT INTO datacenter_maps VALUES('local',1,2);
INSERT INTO datacenter_maps VALUES('ghost-army.map',2,3);
*/
