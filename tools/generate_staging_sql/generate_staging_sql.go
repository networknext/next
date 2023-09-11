package main

import (
	"fmt"
	"os"

	"github.com/networknext/next/modules/common"
)

const NumRelays = 1000

func main() {

	// generate staging.sql

	fmt.Printf("\nGenerating staging.sql\n")

	file, err := os.Create("sql/staging.sql")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	header := `
INSERT INTO route_shaders(route_shader_name) VALUES('test');

INSERT INTO buyers
(
	buyer_name,
	buyer_code,
	live,
	public_key_base64, 
	route_shader_id
) 
VALUES(
	'Test',
	'test',
	true,
	'leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==',
	(select route_shader_id from route_shaders where route_shader_name = 'test')
);

INSERT INTO sellers(seller_name, seller_code) VALUES('Test', 'test');
`

	fmt.Fprintf(file, header)

	datacenter_format := `
INSERT INTO datacenters(
	datacenter_name, 
	latitude, 
	longitude, 
	seller_id)
VALUES(
	'test.%03d',
	%.2f,
	%.2f,
	(select seller_id from sellers where seller_code = 'test')
);
`

	for i := 0; i < NumRelays; i++ {
		fmt.Fprintf(file, datacenter_format, i, randomLatitude(), randomLongitude())
	}

	relay_format := `
INSERT INTO relays(
	relay_name,
	public_ip,
	public_port,
	public_key_base64,
	private_key_base64,
	datacenter_id)
VALUES(
	'test.%03d',
	'127.0.0.1',
	%03d,
	'9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=',
	'lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=',
	(select datacenter_id from datacenters where datacenter_name = 'test.%03d')
);
`

	for i := 0; i < NumRelays; i++ {
		fmt.Fprintf(file, relay_format, i, 10000+i, i)
	}

	settings_format := `
INSERT INTO buyer_datacenter_settings VALUES(
	(select buyer_id from buyers where buyer_code = 'test'),
	(select datacenter_id from datacenters where datacenter_name = 'test.%03d'),
	true
);
`

	for i := 0; i < NumRelays; i += 10 {
		fmt.Fprintf(file, settings_format, i)
	}

	file.Close()

	fmt.Printf("\n")
}

func randomLatitude() float32 {
	return float32(common.RandomInt(-90, +90))
}

func randomLongitude() float32 {
	return float32(common.RandomInt(-180, +180))
}
