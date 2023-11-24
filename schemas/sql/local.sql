
SET local.buyer_public_key_base64 = 'fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA==';
SET local.relay_public_key_base64 = 'ayyX2+oaE4FJjoEHGnAWTQ6EeO829If64UEcshgm6xA=';
SET local.relay_private_key_base64 = 'I7cfCSX8Kq62YeFaSd4CTpNwCVr+VlQxcb9+wUukXpk=';

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
	'local.0',
	'127.0.0.1',
	2000,
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
	'local.1',
	'127.0.0.1',
	2001,
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
	'127.0.0.1',
	2002,
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
	'127.0.0.1',
	2003,
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
	'127.0.0.1',
	2004,
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
	'127.0.0.1',
	2005,
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
	'local.6',
	'127.0.0.1',
	2006,
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
	'local.7',
	'127.0.0.1',
	2007,
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
	'local.8',
	'127.0.0.1',
	2008,
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
	'local.9',
	'127.0.0.1',
	2009,
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
