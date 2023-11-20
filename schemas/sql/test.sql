
INSERT INTO route_shaders(route_shader_name,force_next,route_select_threshold,route_switch_threshold) VALUES('raspberry', true, 300, 300);

INSERT INTO buyers
(
	buyer_name,
	buyer_code,
	public_key_base64, 
	route_shader_id,
	live,
	debug,
) 
VALUES(
	'Raspberry',
	'raspberry',
	'UoFYERKJnCt18mU53IsWzlEXD2pYD9yd+TiZiq9+cMF9cHG4kMwRtw==',
	(select route_shader_id from route_shaders where route_shader_name = 'raspberry')
	true,
	true,
);

INSERT INTO sellers(seller_name, seller_code) VALUES('Amazon', 'amazon');
INSERT INTO sellers(seller_name, seller_code) VALUES('Akamai', 'akamai');
INSERT INTO sellers(seller_name, seller_code) VALUES('Alibaba', 'alibaba');
INSERT INTO sellers(seller_name, seller_code) VALUES('Azure', 'azure');
INSERT INTO sellers(seller_name, seller_code) VALUES('DigitalOcean', 'digitalocean');
INSERT INTO sellers(seller_name, seller_code) VALUES('Equinix', 'equinix');
INSERT INTO sellers(seller_name, seller_code) VALUES('G-Core', 'gcore');
INSERT INTO sellers(seller_name, seller_code) VALUES('Google', 'google');
INSERT INTO sellers(seller_name, seller_code) VALUES('HiVelocity', 'hivelocity');
INSERT INTO sellers(seller_name, seller_code) VALUES('Huawei', 'huawei');
INSERT INTO sellers(seller_name, seller_code) VALUES('IBM', 'ibm');
INSERT INTO sellers(seller_name, seller_code) VALUES('Latitude.sh', 'latitude');
INSERT INTO sellers(seller_name, seller_code) VALUES('Oracle', 'oracle');
INSERT INTO sellers(seller_name, seller_code) VALUES('OVH', 'ovh');
INSERT INTO sellers(seller_name, seller_code) VALUES('phoenixNAP', 'phoenixnap');
INSERT INTO sellers(seller_name, seller_code) VALUES('Tencent', 'tencent');
INSERT INTO sellers(seller_name, seller_code) VALUES('VULTR', 'vultr');
INSERT INTO sellers(seller_name, seller_code) VALUES('Yandex', 'yandex');

-- enable datacenters for buyers

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'raspberry'),
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.1'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'raspberry'),
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.2'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'raspberry'),
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.3'),
	true
);

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'raspberry'),
	(select datacenter_id from datacenters where datacenter_name = 'google.iowa.6'),
	true
);
