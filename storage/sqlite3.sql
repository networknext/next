create table relay_states (
  id integer not null,
  name varchar not null,
  primary key (id)
);

create table machine_types (
  id integer not null,
  name varchar not null,
  primary key (id)
);

create table bw_billing_rules (
  id integer not null,
  name varchar not null,
  primary key (id)
);

create table customers (
  id integer primary key autoincrement,
  active boolean not null,
  automatic_signin_domain varchar null,
  customer_name varchar not null,
  customer_code varchar not null
);

create table buyers (
  id integer primary key autoincrement,
  is_live_customer boolean default false,
  sdk3_public_key_data bytea not null,
  sdk3_public_key_id bigint not null,
  display_name varchar not null unique,
  customer_id integer,
  constraint fk_customer_id foreign key (customer_id) references customers(id)
);

create table sellers (
  id integer primary key autoincrement,
  public_egress_price bigint not null,
  public_ingress_price bigint,
  display_name varchar not null unique,
  customer_id integer,
  constraint fk_customer_id foreign key (customer_id) references customers(id)
);

create table route_shaders (
  id integer primary key autoincrement,
  ab_test boolean not null,
  acceptable_latency integer not null,
  acceptable_packet_loss numeric not null,
  bw_envelope_down_kbps integer not null,
  bw_envelope_up_kbps integer not null,
  disable_network_next boolean not null,
  latency_threshold integer not null,
  multipath boolean not null,
  pro_mode boolean not null,
  reduce_latency boolean not null,
  reduce_packet_loss boolean not null,
  selection_percent integer not null,
  buyer_id integer not null,
  constraint fk_buyer_id foreign key (buyer_id) references buyers(id)
);

create table rs_internal_configs (
  id integer primary key autoincrement,
  max_latency_tradeoff integer not null,
  multipath_overload_threshold integer not null,
  route_switch_threshold integer not null,
  rtt_veto_default integer not null,
  rtt_veto_multipath integer not null,
  rtt_veto_packetloss integer not null,
  try_before_you_buy boolean not null,
  force_next boolean not null,
  buyer_id integer not null,
  constraint fk_buyer_id foreign key (buyer_id) references buyers(id)
);

create table banned_users (
  id integer primary key autoincrement,
  user_id bytea not null,
  buyer_id integer not null,
  constraint fk_buyer_id foreign key (buyer_id) references buyers(id)
);

create table datacenters (
  id integer primary key autoincrement,
  display_name varchar not null unique,
  enabled boolean not null,
  latitude numeric not null,
  longitude numeric not null,
  supplier_name varchar,
  street_address varchar not null,
  seller_id integer not null,
  constraint fk_seller_id foreign key (seller_id) references sellers(id)
);

create table relays (
  id integer primary key autoincrement,
  contract_term integer not null,
  display_name varchar not null,
  end_date date not null,
  included_bandwidth_gb integer not null,
  management_ip inet not null,
  max_sessions integer not null,
  mrc bigint not null,
  overage bigint not null,
  port_speed integer not null,
  public_ip inet not null,
  public_ip_port integer not null,
  public_key bytea not null,
  ssh_port integer not null,
  ssh_user varchar not null,
  start_date date not null,
  update_key bytea not null,
  bw_billing_rule integer not null,
  datacenter integer not null,
  machine_type integer not null,
  relay_state integer not null,
  constraint fk_bw_billing_rule foreign key (bw_billing_rule) references bw_billing_rules(id),
  constraint fk_datacenter foreign key (datacenter) references datacenters(id),
  constraint fk_machine_type foreign key (machine_type) references machine_types(id),
  constraint fk_relay_state foreign key (relay_state) references relay_states(id)
);

-- datacenter_maps is a junction table between dcs and buyers
create table datacenter_maps (
  alias varchar not null,
  buyer_id integer not null,
  datacenter_id integer not null,
  primary key (buyer_id, datacenter_id),
  constraint fk_buyer foreign key (buyer_id) references buyers(id),
  constraint fk_datacenter foreign key (datacenter_id) references datacenters(id)
);

create table metadata (
  sync_sequence_number bigint not null
);
-- File generation: 2020/10/21 13:54:32

-- machine_types
insert into machine_types values (0, 'none');
insert into machine_types values (1, 'vm');
insert into machine_types values (2, 'bare-metal');

-- bw_billing_rules
insert into bw_billing_rules values (0, 'none');
insert into bw_billing_rules values (1, 'flat');
insert into bw_billing_rules values (2, 'burst');
insert into bw_billing_rules values (3, 'pool');

-- bw_billing_rules
insert into relay_states values (0, 'enabled');
insert into relay_states values (1, 'maintenance');
insert into relay_states values (2, 'disabled');
insert into relay_states values (3, 'quarantine');
insert into relay_states values (4, 'decommissioned');
insert into relay_states values (5, 'offline');

-- metadata
 insert into metadata (sync_sequence_number) values (0);

-- customers
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'maxihost', 'maxihost');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'linode', 'linode');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (0, '', 'Andrews Test Company 2', 'andrew-test-company-again');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (0, '', 'Andrews Test Company', 'andrew-test-company');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (0, '', 'Test Company', 'test');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, 'valvesoftware.com', 'Valve', 'valve');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'vultr', 'vultr');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (0, '', 'Network Next', 'next');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'amazon', 'amazon');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'google', 'google');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, 'ghost_army.com.net.gov', 'Ghost Army', 'ghost-army');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, 'wolfjawstudios.com', 'Wolfjaw Studios', 'wjs');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (0, '', 'Test', 'testttttt');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'Raspberry', 'pi');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'networknext', 'nn2');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, 'pubg.com', 'PUBG Corporation', 'pubg');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'riot', 'riot');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'azure', 'azure');
insert into customers (active, automatic_signin_domain, customer_name, customer_code) values (1, '', 'Digital Ocean', 'digital-ocean');

-- sellers
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('riot', 1000000000, NULL, (select id from customers where customer_name = 'riot'));
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('zenlayer', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('ovh', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('phoenixnap', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('amazon', 500000000, NULL, (select id from customers where customer_name = 'amazon'));
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('ibm', 9000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('100tb', 500000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('totalserversolutions', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('kamatera', 5000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('singlehop', 1000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('hostroyale', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('oracle', 8500000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('cdsglobal', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('aws', 12000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('multiplay', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('intergrid', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('linode', 500000000, NULL, (select id from customers where customer_name = 'linode'));
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('serversdotcom', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('hosthink', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('google', 500000000, NULL, (select id from customers where customer_name = 'google'));
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('gcore', 0, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('i3d', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('digitalocean', 500000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('leaseweb', 5000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('inap', 3000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('10gbps', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('maxihost', 1000000000, NULL, (select id from customers where customer_name = 'maxihost'));
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('hostirian', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('packet', 0, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('dedicatedsolutions', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('valve', 1000000000, NULL, (select id from customers where customer_name = 'Valve'));
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('vultr', 500000000, NULL, (select id from customers where customer_name = 'vultr'));
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('limelight', 4000000000, NULL, NULL);
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('azure', 0, NULL, (select id from customers where customer_name = 'azure'));
insert into sellers (display_name, public_egress_price, public_ingress_price, customer_id) values ('stackpath', 4000000000, NULL, NULL);

-- buyers
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (1, 'dfJlOdyLFs5RFw9qWA/cnfk4mYqvfnDBfXBxuJDMEbc=', 3142537350691193170, 'Raspberry', (select id from customers where customer_name = 'Raspberry'));
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (1, 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==', 0, 'Ghost Army', (select id from customers where customer_name = 'Ghost Army'));
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (0, 'OsMdghCNP49qRzLxd9qlpak+vHVKDyyc5RMq8i15o7I=', 7468032365971412001, '', NULL);
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (1, 'HufiX1o3MVv2s9P09vB8bJ5rHUTdukgz1gdoUs3Q+dI=', -3288759583582907005, 'test', (select id from customers where customer_name = 'Test'));
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (0, 'isbsYFuv0Zo5WfbJzr59l4E9hMKt6CWVw6RPOnANTQg=', 5663308403387422823, '', NULL);
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (1, 'uLk+H9QWvDxBLyEAP1JBmS5U/rHXi0RyrFammau9t2c=', -4774169926669966443, 'test', (select id from customers where customer_name = 'Test'));
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (0, 'ZFD0Mweea7EOghFvwXNOphwx2jW4qzc33e49d/5Jm/I=', 752500966263632276, '', NULL);
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (0, 'pX1UcYlTtmZjq8IqMc+I4U+NvLJfOcBqu8esM/C8ItI=', 2516944711831843865, 'Valve', (select id from customers where customer_name = 'Valve'));
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (0, 'WQ4gx2Z9M+oO7eia/bqO/lDAO8R4cr8pE4W3riuKjgrhN2hIkjE0pQ==', 4132671100774797655, 'Wolfjaw Studios', (select id from customers where customer_name = 'Wolfjaw Studios'));
insert into buyers (is_live_customer, sdk3_public_key_data, sdk3_public_key_id, display_name, customer_id) values (0, '', 0, 'Test Buyer', NULL);

-- datacenters
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 53.349800, -6.260300, '', 'Dublin, Ireland', 'azure.dublin.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.507400, -0.127800, 'euw2-az2', 'London, UK', 'amazon.london.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.693700, 135.502300, '', 'Osaka, Japan', 'google.osaka.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.271000, -79.941400, '', 'Roanoke, Virginia, United States', 'azure.roanoke.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.110900, 8.682100, 'euc1-az1', 'Frankfurt, Germany', 'amazon.frankfurt.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.050000, -118.250000, '', 'Los Angeles, CA, USA', 'valve.losangeles', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.376900, 8.541700, '', 'Zürich, Switzerland', 'google.zurich.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 13.080000, 80.270000, '', 'Chennai, India', 'valve.chennai', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.507400, -0.127800, '', 'London, UK', 'azure.london.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 39.043800, -77.487400, '', 'Ashburn, Virginia, USA', 'google.ashburn.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.110900, 8.682100, 'euc1-az3', 'Frankfurt, Germany', 'amazon.frankfurt.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 36.169900, -115.139800, '', 'Las Vegas, California', 'google.lasvegas.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.340400, 103.709000, '', 'Jurong West, Singapore', 'google.singapore.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.507400, -0.127800, '', 'London, UK', 'google.london.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.431600, -78.656900, '', '', 'riot.virginia', ( select id from sellers where display_name = 'riot'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 40.230000, -3.340000, '', 'Madrid, Spain', 'valve.madrid', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.606200, -122.332100, '', 'Seattle, WA, USA', 'riot.seattle', ( select id from sellers where display_name = 'riot'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 22.319300, 114.169400, '', 'Hong Kong', 'google.hongkong.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.566500, 126.978000, 'apne2-az3', 'Seoul, South Korea', 'amazon.seoul.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.110900, 8.682100, 'euc1-az2', 'Frankfurt, Germany', 'amazon.frankfurt.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 60.569300, 27.187800, '', 'Hamina, Finland', 'google.hamina.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.501700, -73.567300, 'cac1-az1', 'Montreal, QC, Canada', 'amazon.montreal.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -33.868800, 151.209300, '', 'Sydney, NSW, Australia', 'azure.sydney', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 59.913900, 10.752200, '', 'Oslo, Norway', 'azure.oslo', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.507400, -0.127800, '', 'London, UK', 'google.london.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.480000, -97.530000, '', 'Oklahoma City, Oklahoma, USA', 'valve.oklahoma', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 59.329300, 18.068600, 'eun1-az1', 'Stockholm, Sweden', 'amazon.stockholm.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.501700, -73.567300, '', 'Montréal, Canada', 'google.montreal.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -33.868800, 151.209300, '', 'Sydney, Australia', 'google.sydney.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 52.366700, 4.894500, '', 'Amsterdam, Netherlands', 'azure.amsterdam.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.110900, 8.682100, '', 'Frankfurt, Germany', 'google.frankfurt.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.586835, -93.624959, '', 'Des Moines, Iowa, USA', 'azure.desmoines', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 60.569300, 27.187800, '', 'Hamina, Finland', 'google.hamina.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 48.856600, 2.352200, '', 'Paris, France', 'azure.paris.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -26.178975, 28.079942, '', 'Johannesburg, South Africa', 'valve.johannesburg', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.594600, -121.178700, '', 'The Dalles, Oregon, USA', 'google.oregon.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.261944, -95.860833, '', '', 'google.iowa', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.507400, -0.127800, '', 'London, UK', 'azure.london.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.338208, -121.886329, '', 'San Jose', 'azure.sanjose', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 22.319300, 114.169400, '', 'Hong Kong', 'azure.hongkong', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 48.856600, 2.352200, 'euw3-az3', 'Paris, France', 'amazon.paris.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 36.726500, -78.128900, '', 'South Hill, Virginia, United States', 'azure.southhill.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.878100, -87.629800, '', 'Chicago, Illinois, United States of America', 'vultr.chicago', ( select id from sellers where display_name = 'vultr'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.586800, -93.625000, '', 'Des Moines, Iowa, United States', 'azure.iowa.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 48.856600, 2.352200, 'euw3-az1', 'Paris, France', 'amazon.paris.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.261900, -95.860800, '', 'Council Bluffs, Iowa, USA', 'google.iowa.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 19.076000, 72.877700, '', 'Mumbai, India', 'google.mumbai.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.680000, 139.680000, '', 'Tokyo, Japan', 'valve.tokyo.1', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 53.349800, -6.260300, '', 'Dublin, Ireland', 'azure.dublin.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -12.060000, -77.050000, '', 'Lima, Peru', 'valve.lima', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 52.366700, 4.894500, '', 'Amsterdam, Netherlands', 'azure.amsterdam.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 52.370000, 4.900000, '', 'Amsterdam, Netherlands', 'valve.amsterdam', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 39.010000, -77.430000, '', 'Sterling, Virginia, USA', 'valve.sterling', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 43.804100, -120.554200, 'usw2-az3', 'Oregon, USA', 'amazon.oregon.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 40.760800, -111.891000, '', 'Salt Lake City, Utah', 'google.saltlakecity.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 33.749000, -84.388000, '', 'Atlanta, Georgia, United States of America', 'linode.atlanta', ( select id from sellers where display_name = 'linode'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.261900, -95.860800, '', 'Council Bluffs, Iowa, USA', 'google.iowa.6', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.352100, 103.819800, '', 'Singapore', 'azure.singapore.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.840000, -87.690000, '', 'Chicago, Illinois, USA', 'valve.chicago', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 48.860000, 2.350000, '', 'Paris, France', 'valve.paris', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 33.760000, -84.390000, '', 'Atlanta, Georgia, USA', 'valve.atlanta', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.676200, 139.650300, '', 'Tokyo, Japan', 'azure.tokyo.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -23.550500, -46.633300, '', 'São Paulo, Brazil', 'azure.saopaulo', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 24.051800, 120.516100, '', 'Changhua County, Taiwan', 'google.taiwan.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 22.319300, 114.169400, 'ape1-az3', 'Hong Kong', 'amazon.hongkong.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.120000, 8.680000, '', 'Frankfurt, Germany', 'valve.frankfurt', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.693700, 135.502300, '', 'Osaka, Japan', 'google.osaka.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 52.132600, 5.291300, '', 'Netherlands', 'google.netherlands.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 53.349800, -6.260300, 'euw1-az2', 'Dublin, Ireland', 'amazon.dublin.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 43.804100, -120.554200, 'usw2-az4', 'Oregon, USA', 'amazon.oregon.4', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -23.550500, -46.633300, '', 'São Paulo, Brazil', 'google.saopaulo.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.774900, -122.419400, '', 'San Francisco, CA, United States', 'azure.sanfrancisco', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.271000, -79.941400, '', 'Roanoke, Virginia, United States', 'azure.roanoke.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 19.076000, 72.877700, 'aps1-az1', 'Mumbai, India', 'amazon.mumbai.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 60.569300, 27.187800, '', 'Hamina, Finland', 'google.hamina.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 53.349800, -6.260300, 'euw1-az3', 'Dublin, Ireland', 'amazon.dublin.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.676200, 139.650300, '', 'Tokyo, Japan', 'google.tokyo.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -33.868800, 151.209300, '', 'Sydney, Australia', 'google.sydney.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 53.349800, -6.260300, 'euw1-az1', 'Dublin, Ireland', 'amazon.dublin.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.376900, 8.541700, '', 'Zürich, Switzerland', 'google.zurich.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.052200, -118.243700, '', 'Los Angeles, California', 'google.losangeles.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 39.043800, -77.487400, '', 'Ashburn, Virginia, USA', 'google.ashburn.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 22.310000, 113.920000, '', 'Hong Kong, China', 'valve.hongkong', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 26.066700, 50.557700, 'mes1-az3', 'Bahrain', 'amazon.bahrain.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.431600, -78.656900, 'use1-az2', 'Virginia, USA', 'amazon.virginia.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.234300, -119.852600, '', 'Quincy, Washington, United States', 'azure.quincy.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -33.860000, 151.210000, '', 'Sydney, Australia', 'valve.sydney', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 40.712775, -74.005973, '', '', 'stackpath.newyork', ( select id from sellers where display_name = 'stackpath'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -26.204100, 28.047300, '', 'Johannesburg, South Africa', 'azure.johannesburg', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -23.530000, -46.640000, '', 'Sao Paulo, Brazil', 'valve.saopaulo', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -23.550500, -46.633300, '', 'São Paulo, Brazil', 'google.saopaulo.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.052234, -118.243685, '', '', 'riot.losangeles', ( select id from sellers where display_name = 'riot'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.431600, -78.656900, 'use1-az6', 'Virginia, USA', 'amazon.virginia.6', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 33.072357, -80.038859, '', '', 'google.southcarolina', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 19.076000, 72.877700, 'aps1-az3', 'Mumbai, India', 'amazon.mumbai.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.052200, -118.243700, '', 'Los Angeles, California', 'google.losangeles.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.594600, -121.178700, '', 'The Dalles, Oregon, USA', 'google.oregon.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 13.082700, 80.270700, '', 'Chennai, India', 'azure.chennai', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.501689, -73.567256, '', '', 'google.montreal', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -23.550500, -46.633300, 'sae1-az1', 'São Paulo, Brazil', 'amazon.saopaulo.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.878114, -87.629798, '', '', 'riot.chicago', ( select id from sellers where display_name = 'riot'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.352100, 103.819800, '', 'Singapore', 'azure.singapore.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.676200, 139.650300, 'apne1-az4', 'Tokyo, Japan', 'amazon.tokyo.4', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 26.066700, 50.557700, 'mes1-az1', 'Bahrain', 'amazon.bahrain.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (0, 40.417300, -82.907100, '', 'Ohio, USA', 'packet.ohio', ( select id from sellers where display_name = 'packet'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.610000, -122.330000, '', 'Seattle, Washington, USA', 'valve.seattle', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 19.076000, 72.877700, '', 'Mumbai, India', 'google.mumbai.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 43.653200, -79.383200, '', 'Toronto, Ontario, Canada', 'digitalocean.toronto', ( select id from sellers where display_name = 'digitalocean'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.449100, 3.818400, '', 'St. Ghislain, Belgium', 'google.stghislain.4', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.338200, -121.886300, 'usw1-az1', 'San Jose, California, USA', 'amazon.sanjose.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -23.550500, -46.633300, 'sae1-az2', 'São Paulo, Brazil', 'amazon.saopaulo.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.878114, -87.629798, '', '', 'maxihost.chicago', ( select id from sellers where display_name = 'maxihost'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 14.350000, 120.590000, '', 'Manila, Philippines', 'valve.manila', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.052234, -118.243685, '', '', 'google.losangeles', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 48.846883, 2.337249, '', 'Luxembourg, Paris, France', 'valve.luxembourg', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.606209, -122.332071, '', '', 'stackpath.seattle', ( select id from sellers where display_name = 'stackpath'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 33.196000, -80.013100, '', 'Moncks Corner, South Carolina, USA', 'google.southcarolina.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.110900, 8.682100, '', 'Frankfurt, Germany', 'google.frankfurt.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.271000, -79.941400, '', 'Roanoke, Virginia, United States', 'azure.roanoke.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.449100, 3.818400, '', 'St. Ghislain, Belgium', 'google.stghislain.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.501700, -73.567300, '', 'Montréal, Canada', 'google.montreal.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 46.813900, -71.208000, '', 'Quebec, QC, Canada', 'azure.quebec', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.338200, -121.886300, '', 'San Jose, California, United States of America', 'amazon.sanjose', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.261900, -95.860800, '', 'Council Bluffs, Iowa, USA', 'google.iowa.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.352100, 103.819800, 'apse1-az1', 'Singapore', 'amazon.singapore.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.676200, 139.650300, '', 'Tokyo, Japan', 'google.tokyo.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 36.726500, -78.128900, '', 'South Hill, Virginia, United States', 'azure.southhill.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -33.230000, -70.470000, '', 'Paris, Santiago, Chile', 'valve.santiago', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.606209, -122.332071, '', 'Seattle, USA', 'azure.seattle', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 32.776664, -96.796988, '', '', 'vultr.dallas', ( select id from sellers where display_name = 'vultr'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.052234, -118.243685, '', '', 'stackpath.losangeles', ( select id from sellers where display_name = 'stackpath'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.052200, -118.243700, '', 'Los Angeles, California', 'google.losangeles.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 40.410435, -82.960910, '', '', 'amazon.ohio', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 33.196000, -80.013100, '', 'Moncks Corner, South Carolina, USA', 'google.southcarolina.4', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.352100, 103.819800, 'apse1-az2', 'Singapore', 'amazon.singapore.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.352100, 103.819800, 'apse1-az3', 'Singapore', 'amazon.singapore.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 43.804100, -120.554200, 'usw2-az1', 'Oregon, USA', 'amazon.oregon.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 59.400000, 17.900000, '', 'Stockholm, Sweden', 'valve.stockholm.1', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 53.349800, -6.260300, '', 'Dublin, Ireland', 'azure.dublin.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.507400, -0.127800, '', 'London, UK', 'google.london.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -23.550500, -46.633300, 'sae1-az3', 'São Paulo, Brazil', 'amazon.saopaulo.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 32.776664, -96.796988, '', '', 'stackpath.dallas', ( select id from sellers where display_name = 'stackpath'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.234300, -119.852600, '', 'Quincy, Washington, United States', 'azure.quincy.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 19.076000, 72.877700, '', 'Mumbai, India', 'google.mumbai.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.261900, -95.860800, '', 'Council Bluffs, Iowa, USA', 'google.iowa.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.505100, -122.675000, '', '', 'riot.portland', ( select id from sellers where display_name = 'riot'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.594600, -121.178700, '', 'The Dalles, Oregon, USA', 'google.oregon.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 52.366700, 4.894500, '', 'Amsterdam, Netherlands', 'azure.amsterdam.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.586800, -93.625000, '', 'Des Moines, Iowa, United States', 'azure.iowa.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.606209, -122.332071, '', '', 'limelight.seattle', ( select id from sellers where display_name = 'limelight'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.676200, 139.650300, '', 'Tokyo, Japan', 'azure.tokyo.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.338200, -121.886300, 'usw1-az3', 'San Jose, California, USA', 'amazon.sanjose.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.431600, -78.656900, '', 'Virginia, United States of America', 'amazon.virginia', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -35.280900, 149.130000, '', 'Canberra, ACT, Australia', 'azure.canberra', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 52.220000, 21.000000, '', 'Warsaw, Poland', 'valve.warsaw', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.676200, 139.650300, '', 'Tokyo, Japan', 'google.tokyo.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -33.868800, 151.209300, 'apse2-az1', 'Sydney, NSW, Australia', 'amazon.sydney.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -33.868800, 151.209300, 'apse2-az2', 'Sydney, NSW, Australia', 'amazon.sydney.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.676200, 139.650300, 'apne1-az1', 'Tokyo, Japan', 'amazon.tokyo.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 40.417300, -82.907100, '', 'Ohio, United States of America', 'amazon.ohio.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.637788, -121.202680, '', '', 'google.oregon', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.680000, 139.680000, '', 'Tokyo, Japan', 'valve.tokyo.2', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 24.453900, 54.377300, '', 'Abu Dhabi, United Arab Emirates', 'azure.abudhabi', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 29.424122, -98.493628, '', 'San Antonio, USA', 'azure.sanantonio', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.340400, 103.709000, '', 'Jurong West, Singapore', 'google.singapore.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.340400, 103.709000, '', 'Jurong West, Singapore', 'google.singapore.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -33.868800, 151.209300, 'apse2-az3', 'Sydney, NSW, Australia', 'amazon.sydney.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.566500, 126.978000, 'apne2-az1', 'Seoul, South Korea', 'amazon.seoul.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.566500, 126.978000, 'apne2-az2', 'Seoul, South Korea', 'amazon.seoul.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 18.580000, 72.490000, '', 'Mumbai, India', 'valve.mumbai', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.693700, 135.502300, '', 'Osaka, Japan', 'azure.osaka', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 48.120000, 16.220000, '', 'Vienna, Austria', 'valve.vienna', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.449100, 3.818400, '', 'St. Ghislain, Belgium', 'google.stghislain.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.234300, -119.852600, '', 'Quincy, Washington, United States', 'azure.quincy.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 48.856600, 2.352200, 'euw3-az2', 'Paris, France', 'amazon.paris.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 40.760800, -111.891000, '', 'Salt Lake City, Utah, United States', 'azure.saltlakecity', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.693700, 135.502300, '', 'Osaka, Japan', 'google.osaka.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -33.868800, 151.209300, '', 'Sydney, Australia', 'google.sydney.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 36.726500, -78.128900, '', 'South Hill, Virginia, United States', 'azure.southhill.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.501700, -73.567300, '', 'Montréal, Canada', 'google.montreal.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 38.943750, -77.524431, '', '', 'google.virginia.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 24.051800, 120.516100, '', 'Changhua County, Taiwan', 'google.taiwan.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 18.520400, 73.856700, '', 'Pune, India', 'azure.pune', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 0.000000, 0.000000, '', '', 'local', ( select id from sellers where display_name = 'local'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.481600, -3.179100, '', 'Cardiff, UK', 'azure.cardiff', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 0.000000, 0.000000, '', '', 'local', ( select id from sellers where display_name = 'local'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 40.417300, -82.907100, 'use2-az1', 'Ohio, USA', 'amazon.ohio.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.507400, -0.127800, 'euw2-az3', 'London, UK', 'amazon.london.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.431600, -78.656900, 'use1-az5', 'Virginia, USA', 'amazon.virginia.5', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 50.110900, 8.682100, '', 'Frankfurt, Germany', 'google.frankfurt.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.774929, -122.419415, '', '', 'digitalocean.sanfrancisco', ( select id from sellers where display_name = 'digitalocean'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 43.804100, -120.554200, 'usw2-az2', 'Oregon, USA', 'amazon.oregon.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.586800, -93.625000, '', 'Des Moines, Iowa, United States', 'azure.iowa.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.376900, 8.541700, '', 'Zürich, Switzerland', 'google.zurich.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 33.196000, -80.013100, '', 'Moncks Corner, South Carolina, USA', 'google.southcarolina.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 59.329300, 18.068600, 'eun1-az3', 'Stockholm, Sweden', 'amazon.stockholm.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 25.250000, 55.300000, '', 'Dubai, United Arab Emirates', 'valve.dubai', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.179600, 129.075600, '', 'Busan, South Korea', 'azure.busan', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.507400, -0.127800, 'euw2-az1', 'London, UK', 'amazon.london.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 34.052234, -118.243685, '', '', 'packet.losangeles', ( select id from sellers where display_name = 'packet'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.507400, -0.127800, '', 'London, UK', 'azure.london.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 48.856600, 2.352200, '', 'Paris, France', 'azure.paris.2', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -23.550500, -46.633300, '', 'São Paulo, Brazil', 'google.saopaulo.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 22.319300, 114.169400, 'ape1-az2', 'Hong Kong', 'amazon.hongkong.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 51.510000, -0.130000, '', 'London, UK', 'valve.london', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 19.076000, 72.877700, '', 'Mumbai, India', 'azure.mumbai', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 19.076000, 72.877700, 'aps1-az2', 'Mumbai, India', 'amazon.mumbai.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 22.319300, 114.169400, '', 'Hong Kong', 'google.hongkong.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 25.204800, 55.270800, '', 'Dubai, United Arab Emirates', 'azure.dubai', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.676200, 139.650300, 'apne1-az2', 'Tokyo, Japan', 'amazon.tokyo.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.548500, -121.988600, '', 'Fremont, California, United States of America', 'linode.fremont', ( select id from sellers where display_name = 'linode'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.280000, 103.830000, '', 'Singapore', 'valve.singapore', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 40.712800, -74.006000, '', 'New York, United States of America', 'digitalocean.newyork', ( select id from sellers where display_name = 'digitalocean'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 48.856600, 2.352200, '', 'Paris, France', 'azure.paris.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 41.878114, -87.629798, '', 'Chicago, USA', 'azure.chicago', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 39.043800, -77.487400, '', 'Ashburn, Virginia, USA', 'google.ashburn.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 39.043757, -77.487442, '', '', 'i3d.ashburn', ( select id from sellers where display_name = 'i3d'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 22.319300, 114.169400, '', 'Hong Kong', 'google.hongkong.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.431600, -78.656900, 'use1-az4', 'Virginia, USA', 'amazon.virginia.4', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 22.319300, 114.169400, 'ape1-az1', 'Hong Kong', 'amazon.hongkong.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.120000, -119.280000, '', 'Moses Lake, Washington, USA', 'valve.moseslake', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 26.066700, 50.557700, 'mes1-az2', 'Bahrain', 'amazon.bahrain.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 35.676200, 139.650300, '', 'Tokyo, Japan', 'azure.tokyo.3', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 52.132600, 5.291300, '', 'Netherlands', 'google.netherlands.2', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.431600, -78.656900, 'use1-az1', 'Virginia, USA', 'amazon.virginia.1', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 40.417300, -82.907100, 'use2-az3', 'Ohio, USA', 'amazon.ohio.3', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 45.501700, -73.567300, 'cac1-az2', 'Montreal, QC, Canada', 'amazon.montreal.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 52.132600, 5.291300, '', 'Netherlands', 'google.netherlands.3', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 32.776700, -96.797000, '', 'Dallas, Texas, United States of America', 'linode.dallas', ( select id from sellers where display_name = 'linode'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 47.606200, -122.332100, '', 'Seattle, Washington, United States of America', 'vultr.seattle', ( select id from sellers where display_name = 'vultr'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 38.713430, -78.159081, '', 'Washington, D.C., USA', 'azure.washingtondc', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 59.329300, 18.068600, 'eun1-az2', 'Stockholm, Sweden', 'amazon.stockholm.2', ( select id from sellers where display_name = 'amazon'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, -37.813600, 144.963100, '', 'Melbourne, VIC, Australia', 'azure.melbourne', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 43.653200, -79.383200, '', 'Toronto, ON, Canada', 'azure.toronto', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.566500, 126.978000, '', 'Seoul, South Korea', 'azure.seoul', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 1.352100, 103.819800, '', 'Singapore', 'azure.singapore.1', ( select id from sellers where display_name = 'azure'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 32.776664, -96.796988, '', '', 'i3d.dallas', ( select id from sellers where display_name = 'i3d'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 24.051800, 120.516100, '', 'Changhua County, Taiwan', 'google.taiwan.1', ( select id from sellers where display_name = 'google'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 59.340000, 17.870000, '', 'Stockholm, Sweden', 'valve.stockholm.2', ( select id from sellers where display_name = 'valve'));
insert into datacenters (enabled, latitude, longitude, supplier_name, street_address, display_name, seller_id ) values (1, 37.431600, -78.656900, 'use1-az3', 'Virginia, USA', 'amazon.virginia.3', ( select id from sellers where display_name = 'amazon'));

-- rs_internal_configs
insert into rs_internal_configs (max_latency_tradeoff, multipath_overload_threshold, route_switch_threshold, rtt_veto_default, rtt_veto_multipath, rtt_veto_packetloss, try_before_you_buy, force_next, buyer_id) values (10, 500, 5, -5, -20, -20, 0, 1, ( select id from buyers where display_name = 'Raspberry'));
insert into rs_internal_configs (max_latency_tradeoff, multipath_overload_threshold, route_switch_threshold, rtt_veto_default, rtt_veto_multipath, rtt_veto_packetloss, try_before_you_buy, force_next, buyer_id) values (0, 0, 0, 0, 0, 0, 0, 0, ( select id from buyers where display_name = 'Andrews Test Company 2'));
insert into rs_internal_configs (max_latency_tradeoff, multipath_overload_threshold, route_switch_threshold, rtt_veto_default, rtt_veto_multipath, rtt_veto_packetloss, try_before_you_buy, force_next, buyer_id) values (0, 0, 0, 0, 0, 0, 0, 0, ( select id from buyers where display_name = 'Andrews Test Company'));
insert into rs_internal_configs (max_latency_tradeoff, multipath_overload_threshold, route_switch_threshold, rtt_veto_default, rtt_veto_multipath, rtt_veto_packetloss, try_before_you_buy, force_next, buyer_id) values (0, 0, 0, 0, 0, 0, 0, 0, ( select id from buyers where display_name = 'Test'));

-- route_shaders
insert into route_shaders (ab_test, acceptable_latency, acceptable_packet_loss, bw_envelope_down_kbps, bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, pro_mode, reduce_latency, reduce_packet_loss, selection_percent, buyer_id) values (0, 25, 1.000000, 1024, 1024, 0, 5, 1, 0, 1, 1, 100, ( select id from buyers where display_name = 'Raspberry'));
insert into route_shaders (ab_test, acceptable_latency, acceptable_packet_loss, bw_envelope_down_kbps, bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, pro_mode, reduce_latency, reduce_packet_loss, selection_percent, buyer_id) values (0, 0, 0.000000, 0, 0, 0, 0, 0, 0, 0, 0, 0, ( select id from buyers where display_name = 'Andrews Test Company 2'));
insert into route_shaders (ab_test, acceptable_latency, acceptable_packet_loss, bw_envelope_down_kbps, bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, pro_mode, reduce_latency, reduce_packet_loss, selection_percent, buyer_id) values (0, 0, 0.000000, 0, 0, 0, 0, 0, 0, 0, 0, 0, ( select id from buyers where display_name = 'Andrews Test Company'));
insert into route_shaders (ab_test, acceptable_latency, acceptable_packet_loss, bw_envelope_down_kbps, bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, pro_mode, reduce_latency, reduce_packet_loss, selection_percent, buyer_id) values (0, 0, 0.000000, 0, 0, 0, 0, 0, 0, 0, 0, 0, ( select id from buyers where display_name = 'Test'));

-- relays
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.11', 0, 0, 0, 1000, '127.0.0.11', '40000', 'iA42xSYC66Cr5LzZD1FsiMVjI6HhN9NhnNOrTD6ILnc=', 22, 'root', '0001-01-01T00:00:00Z', 'iA42xSYC66Cr5LzZD1FsiMVjI6HhN9NhnNOrTD6ILnc=', 0, ( select id from datacenters where display_name = 'valve.losangeles'), ( select id from machine_types where name = 'none'), 1, 'valve.losangeles');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '104.160.130.235', 0, 0, 0, 10000, '104.160.130.163', '40000', 'a95ZFBTUjKYWks/kAxvOZePaCpyWwyCgSBe857rmV1U=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.chicago'), ( select id from machine_types where name = 'none'), 4, 'riot.chicago.b');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '35.236.75.236', 0, 12682000000000, 0, 10000, '35.236.75.236', '40000', 'Klm7zDmoJe+kGaFQ0gShOtdBJzEV1PxldTygUwx/3xo=', 22, 'root', '0001-01-01T00:00:00Z', 'aEqYYm7LngpBh0OaVb2pcN+JOxu6qBQnpJVibGa1epE=', 0, ( select id from datacenters where display_name = 'google.losangeles.3'), ( select id from machine_types where name = 'none'), 4, 'google.losangeles.3');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.6', 0, 0, 0, 1000, '127.0.0.6', '40000', 'Zzbuq5XNPjyAHMvy5+BarVgExQBmH+Xh/RF9SDsZHkI=', 22, 'root', '0001-01-01T00:00:00Z', 'Zzbuq5XNPjyAHMvy5+BarVgExQBmH+Xh/RF9SDsZHkI=', 0, ( select id from datacenters where display_name = 'valve.frankfurt'), ( select id from machine_types where name = 'none'), 1, 'valve.frankfurt');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '40.118.134.228', 0, 0, 0, 1000, '40.118.134.228', '40000', 'HA4m1yLKRLUixi8pN5vB5h5Xa2IVjW/TGBs5e3Liz0g=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'azure.sanfrancisco'), ( select id from machine_types where name = 'none'), 4, 'azure.sanfrancisco.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.25', 0, 0, 0, 1000, '127.0.0.25', '40000', 'przX5XewgLenr51tcv3wDwu6bKTMZ7hp/RbQK0Pg5BQ=', 22, 'root', '0001-01-01T00:00:00Z', 'przX5XewgLenr51tcv3wDwu6bKTMZ7hp/RbQK0Pg5BQ=', 0, ( select id from datacenters where display_name = 'valve.stockholm.1'), ( select id from machine_types where name = 'none'), 1, 'valve.stockholm.2');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '172.105.158.138', 0, 0, 0, 0, '172.105.158.138', '40000', 'dZQU1e393lk2362uu66LwgQPguXyXajh9Do+bCRhBQs=', 22, 'root', '0001-01-01T00:00:00Z', '7rc0W/q7wVl0Jp4722U+rtIIqMJEcBmc4MnPtKOapU4=', 0, ( select id from datacenters where display_name = 'linode.atlanta'), ( select id from machine_types where name = 'none'), 4, 'linode.atlanta');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.15', 0, 0, 0, 1000, '127.0.0.15', '40000', 'ai2ovAPbc7mAx80g/2p4G2UpsDnY0gGt8kmZ80N49wY=', 22, 'root', '0001-01-01T00:00:00Z', 'ai2ovAPbc7mAx80g/2p4G2UpsDnY0gGt8kmZ80N49wY=', 0, ( select id from datacenters where display_name = 'valve.chennai'), ( select id from machine_types where name = 'none'), 1, 'valve.chennai');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '35.221.40.80', 0, 0, 0, 0, '35.221.40.80', '40000', 'R+T04mnDnXkGXRKg7XsmKwBxpWDEdNrwrPEb3wFxD2I=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'google.virginia.1'), ( select id from machine_types where name = 'none'), 4, 'google.virginia');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.29', 0, 0, 0, 1000, '127.0.0.29', '40000', 'AtmWzkT3XaGNswcRZE5oY3mPxAhGG/wmboL/gh3rnhs=', 22, 'root', '0001-01-01T00:00:00Z', 'AtmWzkT3XaGNswcRZE5oY3mPxAhGG/wmboL/gh3rnhs=', 0, ( select id from datacenters where display_name = 'valve.vienna'), ( select id from machine_types where name = 'none'), 1, 'valve.vienna');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.19', 0, 0, 0, 1000, '127.0.0.19', '40000', '5sbGYs5aKi8ePgX5sls85+/LzJ9TQ7JVNNyCpDYCFz8=', 22, 'root', '0001-01-01T00:00:00Z', '5sbGYs5aKi8ePgX5sls85+/LzJ9TQ7JVNNyCpDYCFz8=', 0, ( select id from datacenters where display_name = 'valve.chicago'), ( select id from machine_types where name = 'none'), 1, 'valve.chicago');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '104.160.130.241', 0, 0, 0, 10000, '104.160.130.169', '40000', '9kK5aVJYhsJJDaPcWf4o0MW7jyDBR3UUPJMgpX3MFzk=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.seattle'), ( select id from machine_types where name = 'none'), 4, 'riot.seattle.a');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '157.55.169.155', 0, 0, 0, 1000, '157.55.169.155', '40000', 'XUtu0fmp2mg0xC4JkVqyCFjRkj5Ny6jbF4ftewYLn24=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'azure.chicago'), ( select id from machine_types where name = 'none'), 4, 'azure.chicago.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '104.160.130.245', 0, 0, 0, 10000, '104.160.130.173', '40000', 'PKn/mwtatB9vlAp15+SETSeb0bBLUAGTHI6IJxm1dzY=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.portland'), ( select id from machine_types where name = 'none'), 4, 'riot.portland.a');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (1, '0001-01-01T00:00:00Z', 20000, '186.233.186.30', 0, 12300000000000, 5000000000, 1000, '186.233.186.30', '40000', 'CukXXNi8O1o3SazKDFk8/7nsn4foBVH1Ikgcs/XZ4F0=', 22, 'ubuntu', '0001-01-01T00:00:00Z', '75xgazQ4IvNkSOMLhXZbQzyNmxdkz4ZuE3aj9tZNw6U=', 2, ( select id from datacenters where display_name = 'maxihost.chicago'), ( select id from machine_types where name = 'none'), 0, 'maxihost.chicago');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.2', 0, 0, 0, 1000, '127.0.0.2', '40000', 'kZkaj+I1xdYLOgLTBq+rDosbn/xtUGei8IzZdJ07C3E=', 22, 'root', '0001-01-01T00:00:00Z', 'kZkaj+I1xdYLOgLTBq+rDosbn/xtUGei8IzZdJ07C3E=', 0, ( select id from datacenters where display_name = 'valve.atlanta'), ( select id from machine_types where name = 'none'), 1, 'valve.atlanta');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '104.160.130.237', 0, 0, 0, 10000, '104.160.130.165', '40000', '/ZNEMDGTFuG5ha20va4SjsGyJjqyeKhhCINlnulQcGM=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.losangeles'), ( select id from machine_types where name = 'none'), 4, 'riot.losangeles.a');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.5', 0, 0, 0, 1000, '127.0.0.5', '40000', '3o5tDY5TDk5uRaHDEmW089UNNzgxNDYd85Y9Zr1zHA4=', 22, 'root', '0001-01-01T00:00:00Z', '3o5tDY5TDk5uRaHDEmW089UNNzgxNDYd85Y9Zr1zHA4=', 0, ( select id from datacenters where display_name = 'valve.moseslake'), ( select id from machine_types where name = 'none'), 1, 'valve.moseslake');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.1', 0, 0, 0, 1000, '127.0.0.1', '40000', 'vwbOo9/VWaXQZcjkTYCR0BfuEVfSc+omw30nVOjiLnY=', 22, 'root', '0001-01-01T00:00:00Z', 'vwbOo9/VWaXQZcjkTYCR0BfuEVfSc+omw30nVOjiLnY=', 0, ( select id from datacenters where display_name = 'valve.johannesburg'), ( select id from machine_types where name = 'none'), 1, 'valve.johannesburg');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '34.105.97.141', 0, 0, 0, 0, '34.105.97.141', '40000', '5UDG5Kxo8ecYfPbpINQLA06JmkaTrdkFd/2vMUyXXkk=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'google.oregon.1'), ( select id from machine_types where name = 'none'), 4, 'google.oregon.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '3.134.174.102', 0, 0, 0, 0, '3.134.174.102', '40000', 'syaPbbU4QBvZtxHYQ2X1GNjv+HbJeTtKQvmHDHNCmE4=', 22, 'ubuntu', '0001-01-01T00:00:00Z', 'l3NdoUAMwh/E2VI/K6QjnAMP6TKSh909ocb02NzreGw=', 0, ( select id from datacenters where display_name = 'amazon.ohio.2'), ( select id from machine_types where name = 'none'), 4, 'amazon.ohio.2');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '157.245.214.34', 0, 0, 0, 0, '157.245.214.34', '40000', 'LlWl1oO34JJrMa6CS7gXrT9dngMIm5kUtXxDyYZ5UgM=', 22, 'root', '0001-01-01T00:00:00Z', 'mS51Aslm27cLMWQLEccgrmNhelfI+X3lD0KI7bfOWSY=', 0, ( select id from datacenters where display_name = 'digitalocean.newyork'), ( select id from machine_types where name = 'none'), 4, 'digitalocean.newyork');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.20', 0, 0, 0, 1000, '127.0.0.20', '40000', 'bzn84br+qg/+DPCp5hC5RMocaHch130kLVFNZIvabw8=', 22, 'root', '0001-01-01T00:00:00Z', 'bzn84br+qg/+DPCp5hC5RMocaHch130kLVFNZIvabw8=', 0, ( select id from datacenters where display_name = 'valve.paris'), ( select id from machine_types where name = 'none'), 1, 'valve.paris');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.1', 0, 0, 0, 1000, '127.0.0.1', '40000', 'UP5oen2Aj6MdlajCqjp6a+2sx+WwhH47niJ5n5aG1U0=', 22, 'root', '0001-01-01T00:00:00Z', 'UP5oen2Aj6MdlajCqjp6a+2sx+WwhH47niJ5n5aG1U0=', 0, ( select id from datacenters where display_name = 'valve.amsterdam'), ( select id from machine_types where name = 'none'), 1, 'valve.amsterdam');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '45.79.228.25', 0, 0, 0, 0, '45.79.228.25', '40000', 'SkZOD9qpMfrVt5mLVfuxyRBNGfH0y8noHGEjtZrU5CQ=', 22, 'root', '0001-01-01T00:00:00Z', '7xC1foiH4jKYMbiXEwElWgrNVw9gTQ6BnBLvSZB7/7E=', 0, ( select id from datacenters where display_name = 'linode.fremont'), ( select id from machine_types where name = 'none'), 4, 'linode.fremont');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.3', 0, 0, 0, 1000, '127.0.0.3', '40000', 'LVpYBnxkMwzygIKOlwVE+Zy4ZYBlbDiOcoOi6N7bMCQ=', 22, 'root', '0001-01-01T00:00:00Z', 'LVpYBnxkMwzygIKOlwVE+Zy4ZYBlbDiOcoOi6N7bMCQ=', 0, ( select id from datacenters where display_name = 'valve.mumbai'), ( select id from machine_types where name = 'none'), 1, 'valve.mumbai');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.9', 0, 0, 0, 1000, '127.0.0.9', '40000', 'SUovLyz8A7GJDMhxhUS6FOtJ6oYVFl96pYLWjTUYYjE=', 22, 'root', '0001-01-01T00:00:00Z', 'SUovLyz8A7GJDMhxhUS6FOtJ6oYVFl96pYLWjTUYYjE=', 0, ( select id from datacenters where display_name = 'valve.sterling'), ( select id from machine_types where name = 'none'), 1, 'valve.sterling');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.1', 0, 0, 0, 1000, '127.0.0.1', '40000', 'cZycjY54wJeocx2xor4E7HWUA4wmhmpdpr+IeEZ9nQs=', 22, 'root', '0001-01-01T00:00:00Z', 'cZycjY54wJeocx2xor4E7HWUA4wmhmpdpr+IeEZ9nQs=', 0, ( select id from datacenters where display_name = 'valve.dubai'), ( select id from machine_types where name = 'none'), 1, 'valve.dubai');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '54.177.19.113', 0, 0, 0, 10000, '54.177.19.113', '40000', 'h+RubSZ2XEOJMLECOfkqZSsJoKtclNzBNRt9nA/G0DY=', 22, 'ubuntu', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'amazon.sanjose.1'), ( select id from machine_types where name = 'none'), 0, 'amazon.sanjose.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '104.160.130.251', 0, 0, 0, 10000, '104.160.130.179', '40000', 'T/FEkywtoTewBb1kUMkNq8loy717bcM26GteYrtqYSs=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.virginia'), ( select id from machine_types where name = 'none'), 4, 'riot.virginia.b');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.30', 0, 0, 0, 1000, '127.0.0.30', '40000', 'NWgfUMMtRQjJllVE2RSjG8p8aa5m4uNqOOjHtENYDUw=', 22, 'root', '0001-01-01T00:00:00Z', 'NWgfUMMtRQjJllVE2RSjG8p8aa5m4uNqOOjHtENYDUw=', 0, ( select id from datacenters where display_name = 'valve.warsaw'), ( select id from machine_types where name = 'none'), 1, 'valve.warsaw');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '149.28.213.194', 0, 0, 0, 0, '149.28.213.194', '40000', '6d0L0qxtDy613nl3xg2QIbV+f6dRwR7wVFmBXixebjE=', 22, 'root', '0001-01-01T00:00:00Z', 'mfY2hPqrx0321rekwFsGzj0PYcXQ0/XpP7T8ekBTGEU=', 0, ( select id from datacenters where display_name = ''), ( select id from machine_types where name = 'none'), 4, 'vultr.sanjose');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.27', 0, 0, 0, 1000, '127.0.0.27', '40000', 'Zdwt7WewKxEfoFc7kjBU7EGTucggQlbpI3aHPnu8D2E=', 22, 'root', '0001-01-01T00:00:00Z', 'Zdwt7WewKxEfoFc7kjBU7EGTucggQlbpI3aHPnu8D2E=', 0, ( select id from datacenters where display_name = 'valve.tokyo.1'), ( select id from machine_types where name = 'none'), 1, 'valve.tokyo.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '178.128.227.107', 0, 0, 0, 0, '178.128.227.107', '40000', 'Y2fuAJoHfLH/eW5L3mJ4QpeyTkxGjZytK57w5kqPqBg=', 22, 'root', '0001-01-01T00:00:00Z', 'du6h/3eogCTpxrRWHa0b89lxtBDoVb1ATeDCDpUsWvw=', 0, ( select id from datacenters where display_name = 'digitalocean.toronto'), ( select id from machine_types where name = 'none'), 4, 'digitalocean.toronto');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '149.28.53.151', 0, 0, 0, 0, '149.28.53.151', '40000', '/op1p7cr5hrs6erSIuu4vYZLsUenywa8pumi6TqITlI=', 22, 'root', '0001-01-01T00:00:00Z', 'qWCjXxXQK3DyKWNDq0VCuO9Or2/go6U76Cbs1rHxlGM=', 0, ( select id from datacenters where display_name = ''), ( select id from machine_types where name = 'none'), 4, 'vultr.newyork');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (1, '0001-01-01T00:00:00Z', 15000, '66.42.78.118', 0, 12000000000000, 1000000000, 10000, '66.42.78.118', '40000', 'z4FJdMFTovU+QuAcSpDIDdQq2y0BiAZwans1kQFGkT4=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 2, ( select id from datacenters where display_name = 'vultr.seattle'), ( select id from machine_types where name = 'none'), 0, 'vultr.seattle');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '40.65.126.93', 0, 0, 0, 1000, '40.65.126.93', '40000', 'onrD95jQkuHbt7BX45J2EItlsONYkpZdCX360TTXp3A=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'azure.seattle'), ( select id from machine_types where name = 'none'), 4, 'azure.seattle.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '40.86.82.134', 0, 0, 0, 1000, '40.86.82.134', '40000', 'KkaZX+ULf9PoL0TMJwm06ezWu5ez+WTuZgttcDG6gho=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'azure.desmoines'), ( select id from machine_types where name = 'none'), 4, 'azure.desmoines.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.26', 0, 0, 0, 1000, '127.0.0.26', '40000', 'ndCq3uFaaNbJK2bytsokPEpWKzm5lfBoTdbNsW8aoTU=', 22, 'root', '0001-01-01T00:00:00Z', 'ndCq3uFaaNbJK2bytsokPEpWKzm5lfBoTdbNsW8aoTU=', 0, ( select id from datacenters where display_name = 'valve.sydney'), ( select id from machine_types where name = 'none'), 1, 'valve.sydney');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '20.185.47.134', 0, 0, 0, 1000, '20.185.47.134', '40000', 'jLEBjqPvf07kXh+ztTDd+7tW0XKPif96oGckXnZLHFA=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'azure.roanoke.1'), ( select id from machine_types where name = 'none'), 4, 'azure.roanoke.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.8', 0, 0, 0, 1000, '127.0.0.8', '40000', '85/0EBXoG/GysBoT4G7lIrTWSTzGhIDfIAl1dcRibVo=', 22, 'root', '0001-01-01T00:00:00Z', '85/0EBXoG/GysBoT4G7lIrTWSTzGhIDfIAl1dcRibVo=', 0, ( select id from datacenters where display_name = 'valve.hongkong'), ( select id from machine_types where name = 'none'), 1, 'valve.hongkong');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '52.37.242.229', 0, 0, 0, 10000, '52.37.242.229', '40000', 'C7StixitZFbMp6nJ1VSqo+nVMa2bgTQymoKvuXm12iY=', 22, 'ubuntu', '0001-01-01T00:00:00Z', 'l3NdoUAMwh/E2VI/K6QjnAMP6TKSh909ocb02NzreGw=', 0, ( select id from datacenters where display_name = 'amazon.oregon.1'), ( select id from machine_types where name = 'none'), 4, 'amazon.oregon.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.21', 0, 0, 0, 1000, '127.0.0.21', '40000', 'bGqlYKbFzYNqKP/oaAhNPearQyGvCxwq8YimRTGk8Fg=', 22, 'root', '0001-01-01T00:00:00Z', 'bGqlYKbFzYNqKP/oaAhNPearQyGvCxwq8YimRTGk8Fg=', 0, ( select id from datacenters where display_name = 'valve.santiago'), ( select id from machine_types where name = 'none'), 1, 'valve.santiago');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.28', 0, 0, 0, 1000, '127.0.0.28', '40000', '2R4rAs6lyt2jTNzx9jMk06rRc9J4ty4/qnCCTCoFJS8=', 22, 'root', '0001-01-01T00:00:00Z', '2R4rAs6lyt2jTNzx9jMk06rRc9J4ty4/qnCCTCoFJS8=', 0, ( select id from datacenters where display_name = 'valve.tokyo.2'), ( select id from machine_types where name = 'none'), 1, 'valve.tokyo.2');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '104.160.130.239', 0, 0, 0, 1000, '104.160.130.167', '40000', 'dDiUUI+LanCrKG5ieQd6SPc8Z2m2hCWsWpROHNTPLV0=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.losangeles'), ( select id from machine_types where name = 'none'), 4, 'riot.losangeles.b');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.12', 0, 0, 0, 1000, '127.0.0.12', '40000', '/BNnLSH7XuJpDatEVvyJf/z7vXrA4y9uK9g//0xuWG4=', 22, 'root', '0001-01-01T00:00:00Z', '/BNnLSH7XuJpDatEVvyJf/z7vXrA4y9uK9g//0xuWG4=', 0, ( select id from datacenters where display_name = 'valve.london'), ( select id from machine_types where name = 'none'), 1, 'valve.london');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '104.160.130.247', 0, 0, 0, 10000, '104.160.130.175', '40000', 'jj2fn0cWu4Fm6qVZ8/fHB+kk2r5pXDscPs1yLJiDzCY=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.portland'), ( select id from machine_types where name = 'none'), 4, 'riot.portland.b');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '165.227.10.28', 0, 0, 0, 0, '165.227.10.28', '40000', 'YlQMg7yQm+WpSxz8ji1kIB2yF40VIisZEcCh53IojHg=', 22, 'root', '0001-01-01T00:00:00Z', '0DerixPLMy6wQqKg+48Y4pUQJSu95XrpF1TQsZejv4I=', 0, ( select id from datacenters where display_name = 'digitalocean.sanfrancisco'), ( select id from machine_types where name = 'none'), 4, 'digitalocean.sanfrancisco');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '35.226.96.92', 0, 12682000000000, 0, 10000, '35.226.96.92', '40000', 'VKwkL4v3Tk58sCvyBhJNgmRIrhY6pdBnnJI5p09iVHE=', 22, 'root', '0001-01-01T00:00:00Z', 'aEqYYm7LngpBh0OaVb2pcN+JOxu6qBQnpJVibGa1epE=', 0, ( select id from datacenters where display_name = 'google.iowa'), ( select id from machine_types where name = 'vm'), 0, 'google.iowa.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '34.106.29.193', 0, 0, 0, 0, '34.106.29.193', '40000', 'cKAADyU2A0UYK5GSqUGJ9hb+FBZevz8yy6/o2g6KIlk=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'google.saltlakecity.1'), ( select id from machine_types where name = 'none'), 4, 'google.saltlakecity.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '45.76.175.86', 0, 0, 0, 0, '45.76.175.86', '40000', 'PLFtcxlNGRVAS8HMLzSysIwpLeLQIQtJBsphQBv0PTw=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = ''), ( select id from machine_types where name = 'none'), 4, 'vultr.losangeles');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '34.125.125.84', 0, 0, 0, 0, '34.125.125.84', '40000', 'wCrSLat6zjNZvL5AT5t9l2pMatO3vI6KTawFCiNDHDA=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'google.lasvegas.1'), ( select id from machine_types where name = 'none'), 4, 'google.lasvegas.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (1, '0001-01-01T00:00:00Z', 5000, '45.76.58.249', 0, 12000000000000, 1000000000, 10000, '45.76.58.249', '40000', 'gaXbJv7IbA9nbQ/lZ7sv+0rXjmwNTTOTSCmhmcCqZW0=', 22, 'root', '0001-01-01T00:00:00Z', 'QXgOTb4GEkC2/FPNczEpRtgL5yqDNsBbP0QH6RDBheU=', 2, ( select id from datacenters where display_name = 'vultr.dallas'), ( select id from machine_types where name = 'none'), 0, 'vultr.dallas');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.23', 0, 0, 0, 1000, '127.0.0.23', '40000', 'bcOsgEaU2zooyDx9ZrPOpziJu7eQ6A75IlgkCrbMhFw=', 22, 'root', '0001-01-01T00:00:00Z', 'bcOsgEaU2zooyDx9ZrPOpziJu7eQ6A75IlgkCrbMhFw=', 0, ( select id from datacenters where display_name = 'valve.singapore'), ( select id from machine_types where name = 'none'), 1, 'valve.singapore');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (1, '0001-01-01T00:00:00Z', 5000, '45.76.24.216', 0, 12000000000000, 1000000000, 10000, '45.76.24.216', '40000', 'ctm2D0DMaQLQ4GwxCuRtfLXa8atViVt7BQMnvAfCdzU=', 22, 'root', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 2, ( select id from datacenters where display_name = 'vultr.chicago'), ( select id from machine_types where name = 'none'), 0, 'vultr.chicago');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '104.160.130.233', 0, 0, 0, 10000, '104.160.130.161', '40000', '7+5mXxTFOkJq4tLQj/mmGhZ3CC3zqEiJvfw2wH3YIik=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.chicago'), ( select id from machine_types where name = 'none'), 4, 'riot.chicago.a');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '34.94.28.121', 0, 12682000000000, 0, 0, '34.94.28.121', '40000', 'JujvvQZlfgf/Y6wPh4OqtoxmJD696lJiz7N5v64b5QY=', 22, 'root', '0001-01-01T00:00:00Z', 'XZU2E9VyajuOPyvcp0oDtIH12nwpsAFrM4inKn+FxEs=', 0, ( select id from datacenters where display_name = 'google.losangeles.1'), ( select id from machine_types where name = 'none'), 0, 'google.losangeles.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.18', 0, 0, 0, 1000, '127.0.0.18', '40000', 'O58g/CAFaf2lLBXoq4Gn5xjPQyeatq1dEr7UJyrkUUs=', 22, 'root', '0001-01-01T00:00:00Z', 'O58g/CAFaf2lLBXoq4Gn5xjPQyeatq1dEr7UJyrkUUs=', 0, ( select id from datacenters where display_name = 'valve.oklahoma'), ( select id from machine_types where name = 'none'), 1, 'valve.oklahoma');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.24', 0, 0, 0, 1000, '127.0.0.24', '40000', 'OYKxEasQZc1djK7Hjcdpq/kJYdv2KXpAThPshq24hGc=', 22, 'root', '0001-01-01T00:00:00Z', 'OYKxEasQZc1djK7Hjcdpq/kJYdv2KXpAThPshq24hGc=', 0, ( select id from datacenters where display_name = 'valve.stockholm.1'), ( select id from machine_types where name = 'none'), 1, 'valve.stockholm.1');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '144.202.31.156', 0, 0, 0, 0, '144.202.31.156', '40000', 'i32nTzlCMvlPY6ziEdb1vDawBCxy+R85Nkk+qe2N41c=', 22, 'root', '0001-01-01T00:00:00Z', 'C4fGxbQYHwpHEjvZRBxcUIikvhDQGL3O9L/8m9SLrRY=', 0, ( select id from datacenters where display_name = ''), ( select id from machine_types where name = 'none'), 4, 'vultr.atlanta');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '34.199.100.144', 0, 0, 0, 0, '34.199.100.144', '40000', 'q08jOQNJe81RMandWpkZY87kaal3Lfi6HZniU1WWZT4=', 22, 'ubuntu', '0001-01-01T00:00:00Z', 'WeK6GXz+A3zleZy8LnUz7Zvh2b6vr/iLqRXKLDTHcyo=', 0, ( select id from datacenters where display_name = 'amazon.virginia.5'), ( select id from machine_types where name = 'none'), 4, 'amazon.virginia.5');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.1', 0, 0, 0, 1000, '127.0.0.1', '40000', 'Lx+jYhG7zc9Wv6HQoCBpHUUpI0nGTgru9tvwWViPOTw=', 22, 'root', '0001-01-01T00:00:00Z', 'Lx+jYhG7zc9Wv6HQoCBpHUUpI0nGTgru9tvwWViPOTw=', 0, ( select id from datacenters where display_name = 'valve.luxembourg'), ( select id from machine_types where name = 'none'), 1, 'valve.luxembourg');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.17', 0, 0, 0, 1000, '127.0.0.17', '40000', 'w66rvMlyXc1Wt3LVgtb1MXMxX9799EzeDa9k0Ddlp2o=', 22, 'root', '0001-01-01T00:00:00Z', 'w66rvMlyXc1Wt3LVgtb1MXMxX9799EzeDa9k0Ddlp2o=', 0, ( select id from datacenters where display_name = 'valve.manila'), ( select id from machine_types where name = 'none'), 1, 'valve.manilla');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.16', 0, 0, 0, 1000, '127.0.0.16', '40000', 'g8PrpH1EZzF1VLDlaAS6cvnzwR/8HLUCQ/HStLH8jAg=', 22, 'root', '0001-01-01T00:00:00Z', 'g8PrpH1EZzF1VLDlaAS6cvnzwR/8HLUCQ/HStLH8jAg=', 0, ( select id from datacenters where display_name = 'valve.madrid'), ( select id from machine_types where name = 'none'), 1, 'valve.madrid');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.7', 0, 0, 0, 1000, '127.0.0.7', '40000', 'hH+PPgTnAymR5nBPpI/0AWb4f9rHcFeiXJ8OJ2ltXDg=', 22, 'root', '0001-01-01T00:00:00Z', 'hH+PPgTnAymR5nBPpI/0AWb4f9rHcFeiXJ8OJ2ltXDg=', 0, ( select id from datacenters where display_name = 'valve.saopaulo'), ( select id from machine_types where name = 'none'), 1, 'valve.saopaulo');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '104.160.130.249', 0, 0, 0, 10000, '104.160.130.177', '40000', 'eaNsl6Rwvo+UNbsCVYnuLVlkT9r9FfPzr9pq3VwqkzU=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.virginia'), ( select id from machine_types where name = 'none'), 4, 'riot.virginia.a');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 0, '34.73.237.48', 0, 10559000000000, 0, 0, '34.73.237.48', '40000', 'CcdkT/dCuL7CdONr7wxrHVXLYBR105bqBdZKPK31QiU=', 22, 'root', '0001-01-01T00:00:00Z', 'xP1hdTgUGDtqxqZvWrVNiWax2zzmdp2jKiOtkuHT0T8=', 0, ( select id from datacenters where display_name = 'google.southcarolina.2'), ( select id from machine_types where name = 'none'), 4, 'google.southcarolina.2');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.13', 0, 0, 0, 1000, '127.0.0.13', '40000', 'lFLa3gzFu32wUTxI8HBFgif7wjC0+9T3XQGlUHSeDAU=', 22, 'root', '0001-01-01T00:00:00Z', 'lFLa3gzFu32wUTxI8HBFgif7wjC0+9T3XQGlUHSeDAU=', 0, ( select id from datacenters where display_name = 'valve.lima'), ( select id from machine_types where name = 'none'), 1, 'valve.lima');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '127.0.0.1:1234', 0, 0, 0, 1000, '205.196.6.66', '32700', 'A+uQzzhLfC2bVvhKRHNJKYRu6G4vQ4Ax9nmOqpXgXUE=', 22, 'root', '0001-01-01T00:00:00Z', 'A+uQzzhLfC2bVvhKRHNJKYRu6G4vQ4Ax9nmOqpXgXUE=', 0, ( select id from datacenters where display_name = 'valve.seattle'), ( select id from machine_types where name = 'none'), 1, 'valve.seattle');
insert into relays (contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state, display_name) values (0, '0001-01-01T00:00:00Z', 1, '104.160.130.243', 0, 0, 0, 10000, '104.160.130.171', '40000', 'MsII4QkivI6wak+d0NBMFR+wlCP9aUjue5olgf/6nUI=', 22, 'nnadmin', '0001-01-01T00:00:00Z', 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=', 0, ( select id from datacenters where display_name = 'riot.seattle'), ( select id from machine_types where name = 'none'), 4, 'riot.seattle.b');

-- datacenter maps
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('syd', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.sydney'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('man', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.manila'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('waw', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.warsaw'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('multiplay.chicago', ( select id from buyers where display_name = 'Raspberry'), ( select id from datacenters where display_name = 'vultr.chicago'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('tyo1', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.tokyo.2'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('atl', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.atlanta'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('scl', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.santiago'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('maa', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.chennai'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('lux', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.luxembourg'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('iad', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.sterling'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('lax', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.losangeles'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('', ( select id from buyers where display_name = 'test'), ( select id from datacenters where display_name = 'local'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('hkg', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.hongkong'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('lim', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.lima'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('mwh', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.moseslake'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('gru', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.saopaulo'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('', ( select id from buyers where display_name = '0f318c4f6dfda7d9'), ( select id from datacenters where display_name = '47cca9eed5ac8619'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('dxb', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.dubai'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('mad', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.madrid'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('fra', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.frankfurt'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('sto2', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.stockholm.2'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('jnb', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.johannesburg'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('okc', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.oklahoma'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('par', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.paris'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('ams', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.amsterdam'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('lhr', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.london'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('sto', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.stockholm.1'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('ord', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.chicago'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('some.local.alias', ( select id from buyers where display_name = 'test'), ( select id from datacenters where display_name = 'local'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('sgp', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.singapore'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('bom', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.mumbai'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('some.other.chicago', ( select id from buyers where display_name = 'Raspberry'), ( select id from datacenters where display_name = 'vultr.chicago'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('tyo', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.tokyo.1'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('vie', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.vienna'));
insert into datacenter_maps (alias, buyer_id, datacenter_id) values ('sea', ( select id from buyers where display_name = 'Valve'), ( select id from datacenters where display_name = 'valve.seattle'));

-- SQLite only has limited ALTER TABLE support, so we must make new tables to drop the display_name column
-- buyers
BEGIN TRANSACTION;
create temporary table buyers_backup (
	id integer primary key autoincrement,
	is_live_customer boolean default false,
	sdk3_public_key_data bytea not null,
	sdk3_public_key_id bigint not null,
	customer_id integer,
	constraint fk_customer_id foreign key (customer_id) references customers(id)
);
INSERT INTO buyers_backup SELECT id, is_live_customer, sdk3_public_key_data, sdk3_public_key_id, customer_id FROM buyers;
DROP TABLE buyers;
create table buyers (
	id integer primary key autoincrement,
	is_live_customer boolean default false,
	sdk3_public_key_data bytea not null,
	sdk3_public_key_id bigint not null,
	customer_id integer,
	constraint fk_customer_id foreign key (customer_id) references customers(id)
);
INSERT INTO buyers SELECT id, is_live_customer, sdk3_public_key_data, sdk3_public_key_id, customer_id FROM buyers_backup;
DROP TABLE buyers_backup;
COMMIT;

-- sellers
BEGIN TRANSACTION;
create temporary table sellers_backup (
	id integer primary key autoincrement,
	public_egress_price bigint not null,
	public_ingress_price bigint,
	customer_id integer,
	constraint fk_customer_id foreign key (customer_id) references customers(id)
);
INSERT INTO sellers_backup SELECT id, public_egress_price, public_ingress_price, customer_id FROM sellers;
DROP TABLE sellers;
create table sellers (
	id integer primary key autoincrement,
	public_egress_price bigint not null,
	public_ingress_price bigint,
	customer_id integer,
	constraint fk_customer_id foreign key (customer_id) references customers(id)
);
INSERT INTO sellers SELECT id, public_egress_price, public_ingress_price, customer_id FROM sellers_backup;
DROP TABLE sellers_backup;
COMMIT;

-- datacenters
BEGIN TRANSACTION;
create temporary table datacenters_backup (
	id integer primary key autoincrement,
	enabled boolean not null,
	latitude numeric not null,
	longitude numeric not null,
	supplier_name varchar,
	street_address varchar not null,
	seller_id integer not null,
	constraint fk_seller_id foreign key (seller_id) references sellers(id)
);
INSERT INTO datacenters_backup SELECT id, enabled, latitude, longitude, supplier_name, street_address, seller_id FROM datacenters;
DROP TABLE datacenters;
create table datacenters (
	id integer primary key autoincrement,
	enabled boolean not null,
	latitude numeric not null,
	longitude numeric not null,
	supplier_name varchar,
	street_address varchar not null,
	seller_id integer not null,
	constraint fk_seller_id foreign key (seller_id) references sellers(id)
);
INSERT INTO datacenters SELECT id, enabled, latitude, longitude, supplier_name, street_address, seller_id FROM datacenters_backup;
DROP TABLE datacenters_backup;
COMMIT;

-- relays
BEGIN TRANSACTION;
create temporary table relays_backup (
	id integer primary key autoincrement,
	contract_term integer not null,
	end_date date not null,
	included_bandwidth_gb integer not null,
	management_ip inet not null,
	max_sessions integer not null,
	mrc bigint not null,
	overage bigint not null,
	port_speed integer not null,
	public_ip inet not null,
	public_ip_port integer not null,
	public_key bytea not null,
	ssh_port integer not null,
	ssh_user varchar not null,
	start_date date not null,
	update_key bytea not null,
	bw_billing_rule integer not null,
	datacenter integer not null,
	machine_type integer not null,
	relay_state integer not null,
	constraint fk_bw_billing_rule foreign key (bw_billing_rule) references bw_billing_rules(id),
	constraint fk_datacenter foreign key (datacenter) references datacenters(id),
	constraint fk_machine_type foreign key (machine_type) references machine_types(id),
	constraint fk_relay_state foreign key (relay_state) references relay_states(id)
);
INSERT INTO relays_backup SELECT id, contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state FROM relays;
DROP TABLE relays;
create table relays (
	id integer primary key autoincrement,
	contract_term integer not null,
	end_date date not null,
	included_bandwidth_gb integer not null,
	management_ip inet not null,
	max_sessions integer not null,
	mrc bigint not null,
	overage bigint not null,
	port_speed integer not null,
	public_ip inet not null,
	public_ip_port integer not null,
	public_key bytea not null,
	ssh_port integer not null,
	ssh_user varchar not null,
	start_date date not null,
	update_key bytea not null,
	bw_billing_rule integer not null,
	datacenter integer not null,
	machine_type integer not null,
	relay_state integer not null,
	constraint fk_bw_billing_rule foreign key (bw_billing_rule) references bw_billing_rules(id),
	constraint fk_datacenter foreign key (datacenter) references datacenters(id),
	constraint fk_machine_type foreign key (machine_type) references machine_types(id),
	constraint fk_relay_state foreign key (relay_state) references relay_states(id)
);INSERT INTO relays SELECT id, contract_term, end_date, included_bandwidth_gb, management_ip, max_sessions, mrc, overage, port_speed, public_ip, public_ip_port, public_key, ssh_port, ssh_user, start_date, update_key, bw_billing_rule, datacenter, machine_type, relay_state FROM relays_backup;
DROP TABLE relays_backup;
COMMIT;
