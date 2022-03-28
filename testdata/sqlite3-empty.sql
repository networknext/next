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

create table database_bin_meta (
  bin_file_creation_time date not null,
  bin_file_author varchar not null
);

create table customers (
  id integer primary key autoincrement,
  automatic_signin_domain varchar null,
  customer_name varchar not null,
  customer_code varchar unique not null unique
);

create table buyers (
  id integer primary key autoincrement,
  sdk_generated_id integer not null,
  is_live_customer boolean not null default false,
  debug boolean not null default false,
  analytics boolean not null default false,
  billing boolean not null default false,
  trial boolean not null default true,
  exotic_location_fee bigint not null default 300,
  standard_location_fee bigint not null default 300,
  public_key bytea not null,
  short_name varchar unique,
  customer_id integer not null,
  looker_seats integer not null default 0,
  constraint fk_customer_id foreign key (customer_id) references customers(id)
);

create table sellers (
  id integer primary key autoincrement,
  public_egress_price bigint not null,
  public_ingress_price bigint,
  short_name varchar not null unique,
  customer_id integer,
  secret boolean not null,
  constraint fk_customer_id foreign key (customer_id) references customers(id)
);

create table route_shaders (
  id integer primary key autoincrement,
  ab_test boolean not null,
  acceptable_latency integer not null,
  acceptable_packet_loss numeric not null,
  analysis_only boolean not null,
  bw_envelope_down_kbps integer not null,
  bw_envelope_up_kbps integer not null,
  disable_network_next boolean not null,
  latency_threshold integer not null,
  multipath boolean not null,
  pro_mode boolean not null,
  reduce_latency boolean not null,
  reduce_packet_loss boolean not null,
  reduce_jitter boolean not null,
  selection_percent integer not null,
  packet_loss_sustained integer not null,
  buyer_id integer not null unique,
  constraint fk_buyer_id foreign key (buyer_id) references buyers(id)
);

create table rs_internal_configs (
  id integer primary key autoincrement,
  max_latency_tradeoff integer not null,
  max_rtt integer not null,
  multipath_overload_threshold integer not null,
  route_switch_threshold integer not null,
  route_select_threshold integer not null,
  rtt_veto_default integer not null,
  rtt_veto_multipath integer not null,
  rtt_veto_packetloss integer not null,
  try_before_you_buy boolean not null,
  force_next boolean not null,
  large_customer boolean not null,
  is_uncommitted boolean not null,
  high_frequency_pings boolean not null,
  route_diversity integer not null,
  multipath_threshold integer not null,
  enable_vanity_metrics boolean not null,
  reduce_pl_min_slice_number integer not null,
  buyer_id integer not null unique,
  constraint fk_buyer_id foreign key (buyer_id) references buyers(id)
);

create table banned_users (
  id integer primary key autoincrement,
  user_id integer not null,
  buyer_id integer not null,
  constraint fk_buyer_id foreign key (buyer_id) references buyers(id)
);

create table datacenters (
  id integer primary key autoincrement,
  hex_id varchar(16),
  display_name varchar not null unique,
  latitude numeric not null,
  longitude numeric not null,
  seller_id integer not null,
  constraint fk_seller_id foreign key (seller_id) references sellers(id)
);

create table relays (
  id integer primary key autoincrement,
  hex_id varchar(16) not null,
  contract_term integer not null,
  display_name varchar not null unique,
  end_date date,
  included_bandwidth_gb integer not null,
  internal_ip inet,
  internal_ip_port integer,
  management_ip varchar not null,
  max_sessions integer not null,
  egress_price_override bigint not null,
  mrc bigint not null,
  overage bigint not null,
  port_speed integer not null,
  max_bandwidth_mbps integer not null,
  public_ip inet,
  public_ip_port integer,
  public_key bytea not null,
  ssh_port integer not null,
  ssh_user varchar not null,
  start_date date,
  bw_billing_rule integer not null,
  datacenter integer not null,
  machine_type integer not null,
  relay_state integer not null,
  billing_supplier integer,
  relay_version varchar not null,
  dest_first boolean not null,
  internal_address_client_routable boolean not null,
  notes varchar,
  constraint fk_bw_billing_rule foreign key (bw_billing_rule) references bw_billing_rules(id),
  constraint fk_datacenter foreign key (datacenter) references datacenters(id),
  constraint fk_machine_type foreign key (machine_type) references machine_types(id),
  constraint fk_relay_state foreign key (relay_state) references relay_states(id),
  constraint fk_billing_supplier foreign key (billing_supplier) references sellers(id)
);

-- datacenter_maps is a junction table between dcs and buyers
create table datacenter_maps (
  alias varchar,
  buyer_id integer not null,
  datacenter_id integer not null,
  primary key (buyer_id, datacenter_id),
  constraint fk_buyer foreign key (buyer_id) references buyers(id),
  constraint fk_datacenter foreign key (datacenter_id) references datacenters(id)
);

create table metadata (
  sync_sequence_number bigint not null
);

create table analytics_dashboard_categories (
  id integer primary key autoincrement,
  order_priority integer not null default 0,
  tab_label varchar not null unique,
  premium boolean not null,
  admin_only boolean not null,
  parent_category_id integer null,
  constraint fk_parent_category_id foreign key (parent_category_id) references analytics_dashboard_categories(id)
);

create table analytics_dashboards (
  id integer primary key autoincrement,
  order_priority integer not null default 0,
  dashboard_name varchar not null,
  looker_dashboard_id integer not null,
  premium boolean not null,
  admin_only boolean not null,
  customer_id integer not null,
  category_id integer not null,
  constraint fk_customer_id foreign key (customer_id) references customers(id),
  constraint fk_category_id foreign key (category_id) references analytics_dashboard_categories(id)
);

-- File generation: 2021/06/10 16:24:45

-- machine_types
insert into machine_types values (0, 'none');
insert into machine_types values (1, 'bare-metal');
insert into machine_types values (2, 'vm');

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
 insert into metadata (sync_sequence_number) values (-1);
