ALTER TABLE relays
ADD COLUMN acceptable_packet_loss numeric not null default 0.0,
