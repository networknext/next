
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

INSERT INTO sellers(seller_name) VALUES('amazon');
INSERT INTO sellers(seller_name) VALUES('akamai');
INSERT INTO sellers(seller_name) VALUES('alibaba');
INSERT INTO sellers(seller_name) VALUES('azure');
INSERT INTO sellers(seller_name) VALUES('digitalocean');
INSERT INTO sellers(seller_name) VALUES('equinix');
INSERT INTO sellers(seller_name) VALUES('gcore');
INSERT INTO sellers(seller_name) VALUES('google');
INSERT INTO sellers(seller_name) VALUES('hivelocity');
INSERT INTO sellers(seller_name) VALUES('huawei');
INSERT INTO sellers(seller_name) VALUES('ibm');
INSERT INTO sellers(seller_name) VALUES('latitude');
INSERT INTO sellers(seller_name) VALUES('oracle');
INSERT INTO sellers(seller_name) VALUES('ovh');
INSERT INTO sellers(seller_name) VALUES('phoenixnap');
INSERT INTO sellers(seller_name) VALUES('tencent');
INSERT INTO sellers(seller_name) VALUES('vultr');
INSERT INTO sellers(seller_name) VALUES('yandex');

\i 'sellers/amazon.sql'
\i 'sellers/akamai.sql'
\i 'sellers/alibaba.sql'
\i 'sellers/azure.sql'
\i 'sellers/digitalocean.sql'
\i 'sellers/equinix.sql'
\i 'sellers/gcore.sql'
\i 'sellers/google.sql'
\i 'sellers/hivelocity.sql'
\i 'sellers/huawei.sql'
\i 'sellers/ibm.sql'
\i 'sellers/latitude.sql'
\i 'sellers/oracle.sql'
\i 'sellers/ovh.sql'
\i 'sellers/phoenixnap.sql'
\i 'sellers/tencent.sql'
\i 'sellers/vultr.sql'
\i 'sellers/yandex.sql'

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
