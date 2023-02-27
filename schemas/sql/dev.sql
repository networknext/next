
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

-- amazon datacenters

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.ohio.2',
	40.4173, 
	-82.9071,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.oregon.1',
	45.8399,
	-119.7006,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.sanjose.1',
	37.3387,
	-121.8853,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.saopaulo.1',
	-23.5558, 
	-46.6396,
	(select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'amazon.virginia.1',
	39.0438,
	-77.4874,
	(select seller_id from sellers where seller_name = 'amazon')
);

-- google datacenters

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.iowa.1',
	41.2619,
	-95.8608,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.lasvegas.1',
	36.1716,
	-115.1391,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.oregon.2',
	45.8399,
	-119.7006,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saltlakecity.1',
	40.7608,
	-111.8910,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.santiago.1',
	-33.4489,
	-70.6693,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.saopaulo.1',
	-23.5558, 
	-46.6396,
	(select seller_id from sellers where seller_name = 'google')
);

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'google.virginia.3',
	39.0438,
	-77.4874,
	(select seller_id from sellers where seller_name = 'google')
);

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

-- amazon relays

INSERT INTO relays(
	relay_name,
	public_ip,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'amazon.ohio.2',
	'3.138.73.252',
	'ubuntu',
	'ZlpZo9pD3sFPXynxtc5IV+02TrUmHYuxJc1uffyQkmI=',
	'vLKdOuM8tpbcF6ZtkeThlkYNVT7SWPd9c2eAdvFQQq0=',
	(select datacenter_id from datacenters where datacenter_name = 'amazon.ohio.2')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'amazon.oregon.1',
	'44.242.70.57',
	'ubuntu',
	'MIxnbFMdR04xFwwipYtokcymfh0+xCGCbaryAb5U9zw=',
	'la9ykstfh8f93K7+sKbIi7GQhAW9GIaGkBRs61d47tM=',
	(select datacenter_id from datacenters where datacenter_name = 'amazon.oregon.1')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'amazon.sanjose.1',
	'52.52.246.62',
	'ubuntu',
	'HgZtHcZWzyihZUunYoU6Jmh2wnoEQEM3skDOo5Q4Nyo=',
	'dxG105dIZhb8ajIMyRZKSIPBaBXQG/fsveOZR4eAivg=',
	(select datacenter_id from datacenters where datacenter_name = 'amazon.sanjose.1')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'amazon.saopaulo.1',
	'54.94.14.133',
	'ubuntu',
	'cwuB9C364MMEVeFOvo0VFH+HueZzesjyJ9FIP5gpBhw=',
	'nvtdVeYSmEFVFo1sUTg3y4C1pvhM5w3JKl5PIc/UWC0=',
	(select datacenter_id from datacenters where datacenter_name = 'amazon.saopaulo.1')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	ssh_user,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'amazon.virginia.1',
	'34.232.104.206',
	'ubuntu',
	'WvoyGHCOBSyNPamA1qQ4WlSTpnBhIWepht0utgUSPQ8=',
	'R9IfNVadwq8PPNu21VvwSROoccGwr9z7xT8VlMPTeas=',
	(select datacenter_id from datacenters where datacenter_name = 'amazon.virginia.1')
);

-- google relays

INSERT INTO relays(
	relay_name,
	public_ip,
	internal_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'google.iowa.1',
	'35.226.96.92',
	'10.128.0.8',
	'ZlpZo9pD3sFPXynxtc5IV+02TrUmHYuxJc1uffyQkmI=',
	'vLKdOuM8tpbcF6ZtkeThlkYNVT7SWPd9c2eAdvFQQq0=',
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.1')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	internal_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'google.lasvegas.1',
	'34.125.125.84',
	'10.182.0.7',
	'xKAh+sLW1ghyIkHfOVzkinZZU0mjZF52a+aQ1cv9RRg=',
	'Goql8jwWRUYZpV8XtTPjXC+pDLUzrQ0zpbi8OvElHYw=',
	(select datacenter_id from datacenters where datacenter_name = 'google.lasvegas.1')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	internal_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'google.oregon.2',
	'34.168.209.101',
	'10.138.0.12',
	'fqbsO0Yw/mP82T0JIPblltGtf9xsLbtSqiWaF5/xv38=',
	'Hy4r9eTSq9vEeiYyaOyghdll39FZndwCzihzZQ6RVqA=',
	(select datacenter_id from datacenters where datacenter_name = 'google.oregon.2')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	internal_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'google.saltlakecity.1',
	'34.106.29.193',
	'10.180.0.6',
	'9AftbXkssUIQfK9/zsG/KsPaONd/Uq9FeM/x5iHkGlQ=',
	'sNA5bYajFhUo561VEkpqr6KnTlEK4bDrSUyY5NsDv8E=',
	(select datacenter_id from datacenters where datacenter_name = 'google.saltlakecity.1')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	internal_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'google.santiago.1',
	'34.176.85.20',
	'10.194.0.5',
	'E1ZkLyobOMFZPP7cbqpKcEEb79Z0ZIW/IDaUSWliOl0=',
	'opsJFrG1lO5HUZanT5+RML0qdJaFj5ws8YIOVNtRcG0=',
	(select datacenter_id from datacenters where datacenter_name = 'google.santiago.1')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	internal_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'google.saopaulo.1',
	'34.151.248.241',
	'10.158.0.15',
	'qunlVxGncMg5b650wXgtYBmJAzetry+K9ancBayMWzw=',
	'1vpJ9L6jntr+KvqHSkZvgH9EnkVE/stS+60pfAdXEkg=',
	(select datacenter_id from datacenters where datacenter_name = 'google.saopaulo.1')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	internal_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'google.virginia.3',
	'35.236.236.4',
	'10.150.0.19',
	'5KcEWA5Digp5hBm5TOfXtX3twEk/etE0SB8rwlIrjWQ=',
	'SCrHFjgowY4n4fEkPZnS8wvxseCUiwFGHvaCSkJItqo=',
	(select datacenter_id from datacenters where datacenter_name = 'google.virginia.3')
);

-- linode relays

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'linode.atlanta',
	'45.79.196.195',
	'fUFfw6vkvqv2z+SPFr+I5ZcMpp+p0KxkaLq162MB+jI=',
	'NDyymSNlYO/lCyFDsx7FARildlNStM+tLdzWABZsOdk=',
	(select datacenter_id from datacenters where datacenter_name = 'linode.atlanta')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'linode.dallas',
	'173.255.196.49',
	'fxNrvZ8tLBd8z81NFlB+Oik1epnT6dk1fbXswAjhkDs=',
	'KPK/6dseWtF1bV58FfqR7VgnyxTfMyhaFUbqXDcSlyk=',
	(select datacenter_id from datacenters where datacenter_name = 'linode.dallas')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'linode.fremont',
	'45.33.40.47',
	'cfFFH314m0osgX4mhddhcTOl4uPKNh/0Q3wmV1BNnWo=',
	'TwchbD8m1iZwZJT3D6MzvSkqeJIIO3IBdHZ38ssrMtI=',
	(select datacenter_id from datacenters where datacenter_name = 'linode.fremont')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'linode.newark',
	'23.239.15.72',
	'q0v0bKejhaSZWUTyJOGk05XDX7YTEQ9+9tgKtwoHelA=',
	'L/cWImT8zyPyroDVYrNqMWzsBx73XgBcUFHMaGkg0hg=',
	(select datacenter_id from datacenters where datacenter_name = 'linode.newark')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'linode.toronto',
	'172.105.104.156',
	'j31ln+qWzsS/ZY3rjLBdsE94dKIZmWhhvWGUL5/E6hM=',
	'z8kiGRNyGcMz1BfFC8aBtPI10y2OQkSz2VJrz0QXP2M=',
	(select datacenter_id from datacenters where datacenter_name = 'linode.toronto')
);

-- vultr relays

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'vultr.atlanta',
	'144.202.26.241',
	'IYOu0yZo0dm4KW+QHYjOnyS1NYwBH5is6K+0gZluUhs=',
	'JZKEycAXfKji0VU9YijBYbJyrxjZc36+c5UunijL520=',
	(select datacenter_id from datacenters where datacenter_name = 'vultr.atlanta')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'vultr.chicago',
	'45.76.24.216',
	'la+n/QYoNKwp9TtuPPdA6uvQ2W+/cxSAQqRRJU7e9Q4=',
	'5SxSbKBSEg8tRisoJspIOquh3I0yg7iLx00gWANHxpo=',
	(select datacenter_id from datacenters where datacenter_name = 'vultr.chicago')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'vultr.dallas',
	'144.202.68.72',
	'6MkrKzH1r0wAgr4gKFx46oQCYJiOHiGPBv53eHKGDww=',
	'0IA+AAbGyKLfyZo83i0G4upSItVDtcWzykx9Y7PFXb0=',
	(select datacenter_id from datacenters where datacenter_name = 'vultr.dallas')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'vultr.honolulu',
	'208.83.233.36',
	'2Gb84VP9pVB7RDTFA2uyNP1mo2OwY3jAUkq9tkGIHRU=',
	'hEeViWLYNeSJx0cXwOZ+YWH41DRoTM1rhK0ZVKg/RGA=',
	(select datacenter_id from datacenters where datacenter_name = 'vultr.honolulu')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'vultr.losangeles',
	'149.248.16.231',
	'uJ+ls/NZ0Le9HdlwUR4gcirsjBfSGlGqWBi+eYOxHHc=',
	'rLZsQtQj5l+OYCD0qgFpn6DbcbgjZNEMhpz12YucOKE=',
	(select datacenter_id from datacenters where datacenter_name = 'vultr.losangeles')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'vultr.miami',
	'45.32.160.85',
	'SEusGh4CTLUH4LDtdVhoTlgXDN7xVmMManZUg2PPblM=',
	'H/CoFsxtt1GtdD9Iu5Cfta3JrBkZmpw+yshf3QBpQmw=',
	(select datacenter_id from datacenters where datacenter_name = 'vultr.miami')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'vultr.newyork',
	'45.76.6.145',
	'h6u6/CvUPX+zL7VX0lNADbKJ1/Tf3POt+jWSBmR3ZTk=',
	'gVfQB1MNeFi82asRgvN7cfdk7ZrQDHaUDGVyDjHhEF8=',
	(select datacenter_id from datacenters where datacenter_name = 'vultr.newyork')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'vultr.seattle',
	'144.202.82.197',
	'9FduULKBYgmTF8SboZLrHoCGHjJOiTTWA4SvFfKoamA=',
	'YO1iBsbKjGtIzuQXhf8vdgGyWKp2HHAqg+zcrnXS5zE=',
	(select datacenter_id from datacenters where datacenter_name = 'vultr.seattle')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'vultr.siliconvalley',
	'149.28.197.73',
	'pbyRbBQA94+qyIJhJaQ8ZRiIfDgzrL4ucfb+Z/NUni8=',
	'U6zE8ysmZ57jUWRUs/MtC4BV25GlJK/xr1ABWLnxumg=',
	(select datacenter_id from datacenters where datacenter_name = 'vultr.siliconvalley')
);

-- enable datacenters for buyers

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_name = 'raspberry'),
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.1'),
	true
);
