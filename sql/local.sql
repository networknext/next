
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
	'Local',
	'local'
);

INSERT INTO route_shaders(short_name, force_next) VALUES('local', true);

INSERT INTO buyers
(
	short_name,
	public_key_base64, 
	customer_id,
	route_shader_id

) 
VALUES(
	'local',
	'leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==',
	(select id from customers where customer_code = 'local'),
	(select id from route_shaders where short_name = 'local')
);

INSERT INTO sellers(short_name) VALUES('local');

-- local datacenters

INSERT INTO datacenters(
	display_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'local',
	40.7128,
	-74.0060,
	(select id from sellers where short_name = 'local')
);

-- local relays

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.0',
	'127.0.0.1',
	2000,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.1',
	'127.0.0.1',
	2001,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.2',
	'127.0.0.1',
	2002,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.3',
	'127.0.0.1',
	2003,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.4',
	'127.0.0.1',
	2004,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.5',
	'127.0.0.1',
	2005,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.6',
	'127.0.0.1',
	2006,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.7',
	'127.0.0.1',
	2007,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.8',
	'127.0.0.1',
	2008,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

INSERT INTO relays(
	display_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter)
VALUES(
	'local.9',
	'127.0.0.1',
	2009,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select id from datacenters where display_name = 'local')
);

-- enable datacenters for buyers

INSERT INTO datacenter_maps VALUES(
	(select id from buyers where short_name = 'local'),
	(select id from datacenters where display_name = 'local'),
	true
);
