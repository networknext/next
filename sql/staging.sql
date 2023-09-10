
INSERT INTO route_shaders(route_shader_name, force_next) VALUES('test', true);

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
	'Test',
	'test',
	'leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==',
	(select route_shader_id from route_shaders where route_shader_name = 'test')
);

INSERT INTO sellers(seller_name, seller_code) VALUES('Test', 'test');

-- test datacenters

INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test',
	40.7128,
	-74.0060,
	(select seller_id from sellers where seller_code = 'test')
);

-- test relays

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.0',
	'127.0.0.1',
	2000,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.1',
	'127.0.0.1',
	2001,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.2',
	'127.0.0.1',
	2002,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.3',
	'127.0.0.1',
	2003,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.4',
	'127.0.0.1',
	2004,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.5',
	'127.0.0.1',
	2005,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.6',
	'127.0.0.1',
	2006,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.7',
	'127.0.0.1',
	2007,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.8',
	'127.0.0.1',
	2008,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.9',
	'127.0.0.1',
	2009,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test')
);

-- enable datacenters for buyers

INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test'),
	true
);
