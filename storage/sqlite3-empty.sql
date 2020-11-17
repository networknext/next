-- enforce primary keys
PRAGMA foreign_keys = on;

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
  debug boolean not null default false,
  automatic_signin_domain varchar null,
  customer_name varchar not null,
  customer_code varchar unique not null
);

create table buyers (
  id integer primary key autoincrement,
  is_live_customer boolean not null default false,
  debug boolean not null default false,
  public_key bytea not null,
  short_name varchar unique,
  customer_id integer not null,
  constraint fk_customer_id foreign key (customer_id) references customers(id)
);

create table sellers (
  id integer primary key autoincrement,
  public_egress_price bigint not null,
  public_ingress_price bigint,
  short_name varchar unique,
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
  large_customer boolean not null,
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

-- File generation: 2020/11/12 12:55:14

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
