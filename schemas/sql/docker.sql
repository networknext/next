
SET local.buyer_public_key_base64 = 'qfWdjLu3HjFti1XkneM4z+0+ZUDa03eb7iq/wA0G+Uu6heI5RhRDuQ==';
SET local.relay_public_key_base64 = 'Ugg4eokkkU6SaDyAVK9yF2rE5TBMVNEJruMUp+nrLAc=';
SET local.relay_private_key_base64 = 'Gp87UJIdEYWFPNPy+b2pRu4Xj05eyFZOiPVIjRDwAtU=';

INSERT INTO route_shaders(route_shader_name, force_next) VALUES('local', true);

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
	true,
	'Local',
	'local',
	current_setting('local.buyer_public_key_base64'),
	(select route_shader_id from route_shaders where route_shader_name = 'local')
);

INSERT INTO sellers(seller_name, seller_code) VALUES('Local', 'local');

-- local datacenters

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'local',
	40.7128,
	-74.0060,
	(select seller_id from sellers where seller_code = 'local')
);

-- local relays

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'local.1',
	'10.5.0.5',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'local')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'local.2',
	'10.5.0.6',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'local')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'local.3',
	'10.5.0.7',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'local')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'local.4',
	'10.5.0.8',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'local')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'local.5',
	'10.5.0.9',
	40000,
	current_setting('local.relay_public_key_base64'),
	current_setting('local.relay_private_key_base64'),
	(select datacenter_id from datacenters where datacenter_name = 'local')
);

-- enable datacenters for buyers

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'local'),
	(select datacenter_id from datacenters where datacenter_name = 'local'),
	true
);
