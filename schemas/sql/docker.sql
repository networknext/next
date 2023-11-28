
SET local.buyer_public_key_base64 = '9qzGNONKAHTBaPsm+b9pPUgEvekv3iKZBdXJt7eSBePFkeWtoxpGig==';
SET local.relay_public_key_base64 = '02YtwLT5RTPlEjxEe0oo/0EP3OFLOJdWLA5jxz3J5VY=';
SET local.relay_private_key_base64 = 'JB2hC7sEaj2ujpthoOWyEqKAqsBzQgrutUBopPShiuM=';

INSERT INTO route_shaders(route_shader_name, route_select_threshold, route_switch_threshold) VALUES('docker', 300, 300);

INSERT INTO buyers
(
	live,
	debug,
	buyer_name,
	buyer_code,
	public_key_base64, 
	route_shader_id
) 
VALUES(
	true,
	false,
	'Docker',
	'docker',
	current_setting('local.buyer_public_key_base64'),
	(select route_shader_id from route_shaders where route_shader_name = 'docker')
);

INSERT INTO sellers(seller_name, seller_code) VALUES('Docker', 'docker');

-- docker datacenters

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'docker',
	40.7128,
	-74.0060,
	(select seller_id from sellers where seller_code = 'docker')
);

-- docker relays

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.0',
	'10.20.1.0',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.1',
	'10.20.1.1',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.2',
	'10.20.1.2',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.3',
	'10.20.1.3',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.4',
	'10.20.1.4',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.5',
	'10.20.1.5',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.6',
	'10.20.1.6',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.7',
	'10.20.1.7',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.8',
	'10.20.1.8',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'docker.9',
	'10.20.1.9',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'docker')
);

-- enable datacenters for buyers

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'docker'),
	(select datacenter_id from datacenters where datacenter_name = 'docker'),
	true
);
