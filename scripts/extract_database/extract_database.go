package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/networknext/backend/modules/common"
	db "github.com/networknext/backend/modules/database"
)

func ParseAddress(input string) net.UDPAddr {
	address := net.UDPAddr{}
	ip_string, port_string, err := net.SplitHostPort(input)
	if err != nil {
		address.IP = net.ParseIP(input)
		address.Port = 0
		return address
	}
	address.IP = net.ParseIP(ip_string)
	address.Port, _ = strconv.Atoi(port_string)
	return address
}

func main() {

	pgsql, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=gaffer dbname=network_next sslmode=disable")
	if err != nil {
		fmt.Printf("error: could not connect to postgres: %v\n", err)
		os.Exit(1)
	}

	err = pgsql.Ping()
	if err != nil {
		fmt.Printf("error: could not ping postgres: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("successfully connected to postgres\n")

	// relays

	type RelayRow struct {
		id                 uint64
		name               string
		datacenter         uint64
		public_ip          string
		public_port        int
		internal_ip        string
		internal_port      int
		ssh_ip             string
		ssh_port           int
		ssh_user           string
		public_key_base64  string
		private_key_base64 string
		version            sql.NullString
		mrc                int
		port_speed         int
		max_sessions       int
	}

	relayRows := make([]RelayRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, display_name, datacenter, public_ip, public_port, internal_ip, internal_port, ssh_ip, ssh_port, ssh_user, public_key_base64, private_key_base64, version, mrc, port_speed, max_sessions FROM relays")
		if err != nil {
			fmt.Printf("error: could not extract relays: %v\n", err)
			os.Exit(1)
		}

		defer rows.Close()

		for rows.Next() {
			row := RelayRow{}
			if err := rows.Scan(&row.id, &row.name, &row.datacenter, &row.public_ip, &row.public_port, &row.internal_ip, &row.internal_port, &row.ssh_ip, &row.ssh_port, &row.ssh_user, &row.public_key_base64, &row.private_key_base64, &row.version, &row.mrc, &row.port_speed, &row.max_sessions); err != nil {
				fmt.Printf("error: failed to scan relay row: %v\n", err)
				os.Exit(1)
			}
			relayRows = append(relayRows, row)
		}
	}

	// datacenters

	type DatacenterRow struct {
		id        uint64
		name      string
		latitude  float32
		longitude float32
		seller_id uint64
	}

	datacenterRows := make([]DatacenterRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, display_name, latitude, longitude, seller_id FROM datacenters")
		if err != nil {
			fmt.Printf("error: could not extract datacenters: %v\n", err)
			os.Exit(1)
		}

		defer rows.Close()

		for rows.Next() {
			row := DatacenterRow{}
			if err := rows.Scan(&row.id, &row.name, &row.latitude, &row.longitude, &row.seller_id); err != nil {
				fmt.Printf("error: failed to scan datacenter row: %v\n", err)
				os.Exit(1)
			}
			datacenterRows = append(datacenterRows, row)
		}
	}

	// buyers

	type BuyerRow struct {
		id                uint64
		name              string
		public_key_base64 string
		customer_id       uint64
		route_shader_id   uint64
	}

	buyerRows := make([]BuyerRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, short_name, public_key_base64, customer_id, route_shader_id FROM buyers")
		if err != nil {
			fmt.Printf("error: could not extract buyers: %v\n", err)
			os.Exit(1)
		}

		defer rows.Close()

		for rows.Next() {
			row := BuyerRow{}
			if err := rows.Scan(&row.id, &row.name, &row.public_key_base64, &row.customer_id, &row.route_shader_id); err != nil {
				fmt.Printf("error: failed to scan buyer row: %v\n", err)
				os.Exit(1)
			}
			buyerRows = append(buyerRows, row)
		}
	}

	// sellers

	type SellerRow struct {
		id          uint64
		name        string
		customer_id sql.NullInt64
	}

	sellerRows := make([]SellerRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, short_name, customer_id FROM sellers")
		if err != nil {
			fmt.Printf("error: could not extract sellers: %v\n", err)
			os.Exit(1)
		}

		defer rows.Close()

		for rows.Next() {
			row := SellerRow{}
			if err := rows.Scan(&row.id, &row.name, &row.customer_id); err != nil {
				fmt.Printf("error: failed to scan seller row: %v\n", err)
				os.Exit(1)
			}
			sellerRows = append(sellerRows, row)
		}
	}

	// customers

	type CustomerRow struct {
		id            uint64
		customer_name string
		customer_code string
		live          bool
		debug         bool
	}

	customerRows := make([]CustomerRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, customer_name, customer_code, live, debug FROM customers")
		if err != nil {
			fmt.Printf("error: could not extract customers: %v\n", err)
			os.Exit(1)
		}

		defer rows.Close()

		for rows.Next() {
			row := CustomerRow{}
			if err := rows.Scan(&row.id, &row.customer_name, &row.customer_code, &row.live, &row.debug); err != nil {
				fmt.Printf("error: failed to scan customer row: %v\n", err)
				os.Exit(1)
			}
			customerRows = append(customerRows, row)
		}
	}

	// route shaders

	type RouteShaderRow struct {
		id                           uint64
		ab_test                      bool
		acceptable_latency           int
		acceptable_packet_loss       float32
		packet_loss_sustained        float32
		analysis_only                bool
		bandwidth_envelope_down_kbps int
		bandwidth_envelope_up_kbps   int
		disable_network_next         bool
		latency_threshold            int
		multipath                    bool
		reduce_latency               bool
		reduce_packet_loss           bool
		selection_percent            int
		max_latency_tradeoff         int
		max_next_rtt                 int
		route_switch_threshold       int
		route_select_threshold       int
		rtt_veto_default             int
		rtt_veto_multipath           int
		rtt_veto_packetloss          int
		force_next                   bool
		route_diversity              int
	}

	routeShaderRows := make([]RouteShaderRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, ab_test, acceptable_latency, acceptable_packet_loss, packet_loss_sustained, analysis_only, bandwidth_envelope_down_kbps, bandwidth_envelope_up_kbps, disable_network_next, latency_threshold, multipath, reduce_latency, reduce_packet_loss, selection_percent, max_latency_tradeoff, max_next_rtt, route_switch_threshold, route_select_threshold, rtt_veto_default, rtt_veto_multipath, rtt_veto_packetloss, force_next, route_diversity FROM route_shaders")
		if err != nil {
			fmt.Printf("error: could not extract route shaders: %v\n", err)
			os.Exit(1)
		}

		defer rows.Close()

		for rows.Next() {
			row := RouteShaderRow{}
			if err := rows.Scan(&row.id, &row.ab_test, &row.acceptable_latency, &row.acceptable_packet_loss, &row.packet_loss_sustained, &row.analysis_only, &row.bandwidth_envelope_down_kbps, &row.bandwidth_envelope_up_kbps, &row.disable_network_next, &row.latency_threshold, &row.multipath, &row.reduce_latency, &row.reduce_packet_loss, &row.selection_percent, &row.max_latency_tradeoff, &row.max_next_rtt, &row.route_switch_threshold, &row.route_select_threshold, &row.rtt_veto_default, &row.rtt_veto_multipath, &row.rtt_veto_packetloss, &row.force_next, &row.route_diversity); err != nil {
				fmt.Printf("error: failed to scan route shader row: %v\n", err)
				os.Exit(1)
			}
			routeShaderRows = append(routeShaderRows, row)
		}
	}

	// datacenter maps

	type DatacenterMapRow struct {
		buyer_id            uint64
		datacenter_id       uint64
		enable_acceleration bool
	}

	datacenterMapRows := make([]DatacenterMapRow, 0)
	{
		rows, err := pgsql.Query("SELECT buyer_id, datacenter_id, enable_acceleration FROM datacenter_maps")
		if err != nil {
			fmt.Printf("error: could not extract datacenter maps: %v\n", err)
			os.Exit(1)
		}

		defer rows.Close()

		for rows.Next() {
			row := DatacenterMapRow{}
			if err := rows.Scan(&row.buyer_id, &row.datacenter_id, &row.enable_acceleration); err != nil {
				fmt.Printf("error: failed to scan datacenter map row: %v\n", err)
				os.Exit(1)
			}
			datacenterMapRows = append(datacenterMapRows, row)
		}
	}

	// print out rows

	fmt.Printf("\nrelays:\n")
	for _, row := range relayRows {
		fmt.Printf("%d: %s, %d, %s, %d, %s, %d, %s, %d, %s, %s, %s, %s, %d, %d, %d\n", row.id, row.name, row.datacenter, row.public_ip, row.public_port, row.internal_ip, row.internal_port, row.ssh_ip, row.ssh_port, row.ssh_user, row.public_key_base64, row.private_key_base64, row.version.String, row.mrc, row.port_speed, row.max_sessions)
	}

	fmt.Printf("\ndatacenters:\n")
	for _, row := range datacenterRows {
		fmt.Printf("%d: %s, %.1f, %.1f, %d\n", row.id, row.name, row.latitude, row.longitude, row.seller_id)
	}

	fmt.Printf("\nbuyers:\n")
	for _, row := range buyerRows {
		fmt.Printf("%d: %s, %s, %d\n", row.id, row.name, row.public_key_base64, row.customer_id)
	}

	fmt.Printf("\nsellers:\n")
	for _, row := range sellerRows {
		fmt.Printf("%d: %s, %d\n", row.id, row.name, row.customer_id.Int64)
	}

	fmt.Printf("\ncustomers:\n")
	for _, row := range customerRows {
		fmt.Printf("%d: %s, %s, %v, %v\n", row.id, row.customer_name, row.customer_code, row.live, row.debug)
	}

	fmt.Printf("\nroute shaders:\n")
	for _, row := range routeShaderRows {
		fmt.Printf("%d: %v, %d, %.1f, %v, %d, %d, %v, %d, %v, %v, %v, %d, %d, %d, %d, %d, %d, %d, %d, %v\n", row.id, row.ab_test, row.acceptable_latency, row.acceptable_packet_loss, row.analysis_only, row.bandwidth_envelope_down_kbps, row.bandwidth_envelope_up_kbps, row.disable_network_next, row.latency_threshold, row.multipath, row.reduce_latency, row.reduce_packet_loss, row.selection_percent, row.max_latency_tradeoff, row.max_next_rtt, row.route_switch_threshold, row.route_select_threshold, row.rtt_veto_default, row.rtt_veto_multipath, row.rtt_veto_packetloss, row.force_next)
	}

	fmt.Printf("\ndatacenter maps:\n")
	for _, row := range datacenterMapRows {
		fmt.Printf("(%d,%d): %v\n", row.buyer_id, row.datacenter_id, row.enable_acceleration)
	}

	// index datacenters by postgres id

	datacenterIndex := make(map[uint64]DatacenterRow)
	for _, row := range datacenterRows {
		datacenterIndex[row.id] = row
	}

	// index customers by postgres id

	customerIndex := make(map[uint64]CustomerRow)
	for _, row := range customerRows {
		customerIndex[row.id] = row
	}

	// index buyers by postgres id

	buyerIndex := make(map[uint64]BuyerRow)
	for _, row := range buyerRows {
		buyerIndex[row.id] = row
	}

	// index sellers by postgres id

	sellerIndex := make(map[uint64]SellerRow)
	for _, row := range sellerRows {
		sellerIndex[row.id] = row
	}

	// index route shaders by postgres id

	routeShaderIndex := make(map[uint64]RouteShaderRow)
	for _, row := range routeShaderRows {
		routeShaderIndex[row.id] = row
	}

	// build database

	fmt.Printf("\nbuilding network next database...\n\n")

	database := db.CreateDatabase()

	database.CreationTime = time.Now().Format("Monday 02 January 2006 15:04:05 MST")
	database.Creator = "extract_database"

	database.Relays = make([]db.Relay, len(relayRows))

	for i, row := range sellerRows {

		seller := db.Seller{}

		seller.Id = row.id
		seller.Name = row.name

		database.SellerMap[seller.Id] = seller

		fmt.Printf("seller %d: %s [%d]\n", i, seller.Name, seller.Id)
	}

	for i, row := range buyerRows {

		buyer := db.Buyer{}

		buyer.Name = row.name

		data, err := base64.StdEncoding.DecodeString(row.public_key_base64)
		if err != nil {
			fmt.Printf("error: could not decode public key base64 for buyer %s: %v\n", buyer.Name, err)
			os.Exit(1)
		}

		if len(data) != 40 {
			fmt.Printf("error: buyer public key data must be 40 bytes\n")
			os.Exit(1)
		}

		buyer.Id = binary.LittleEndian.Uint64(data[:8])
		buyer.PublicKey = data[8:40]

		customer_row, customer_exists := customerIndex[row.customer_id]
		if !customer_exists {
			fmt.Printf("error: buyer %s does not have a customer?!\n", buyer.Name)
			os.Exit(1)
		}

		buyer.Live = customer_row.live
		buyer.Debug = customer_row.debug

		route_shader_row, route_shader_exists := routeShaderIndex[row.route_shader_id]
		if !route_shader_exists {
			fmt.Printf("error: buyer %s does not have a route shader?!\n", buyer.Name)
			os.Exit(1)
		}

		buyer.RouteShader.DisableNetworkNext = route_shader_row.disable_network_next
		buyer.RouteShader.AnalysisOnly = route_shader_row.analysis_only
		buyer.RouteShader.SelectionPercent = route_shader_row.selection_percent
		buyer.RouteShader.ABTest = route_shader_row.ab_test
		buyer.RouteShader.ReduceLatency = route_shader_row.reduce_latency
		buyer.RouteShader.ReducePacketLoss = route_shader_row.reduce_packet_loss
		buyer.RouteShader.Multipath = route_shader_row.multipath
		buyer.RouteShader.AcceptableLatency = int32(route_shader_row.acceptable_latency)
		buyer.RouteShader.LatencyThreshold = int32(route_shader_row.latency_threshold)
		buyer.RouteShader.AcceptablePacketLoss = route_shader_row.acceptable_packet_loss
		buyer.RouteShader.PacketLossSustained = route_shader_row.packet_loss_sustained
		buyer.RouteShader.BandwidthEnvelopeUpKbps = int32(route_shader_row.bandwidth_envelope_up_kbps)
		buyer.RouteShader.BandwidthEnvelopeDownKbps = int32(route_shader_row.bandwidth_envelope_down_kbps)
		buyer.RouteShader.RouteSelectThreshold = int32(route_shader_row.route_select_threshold)
		buyer.RouteShader.RouteSwitchThreshold = int32(route_shader_row.route_switch_threshold)
		buyer.RouteShader.MaxLatencyTradeOff = int32(route_shader_row.max_latency_tradeoff)
		buyer.RouteShader.RTTVeto_Default = int32(route_shader_row.rtt_veto_default)
		buyer.RouteShader.RTTVeto_Multipath = int32(route_shader_row.rtt_veto_multipath)
		buyer.RouteShader.RTTVeto_PacketLoss = int32(route_shader_row.rtt_veto_packetloss)
		buyer.RouteShader.MaxNextRTT = int32(route_shader_row.max_next_rtt)
		buyer.RouteShader.ForceNext = route_shader_row.force_next
		buyer.RouteShader.RouteDiversity = int32(route_shader_row.route_diversity)

		database.BuyerMap[buyer.Id] = buyer

		fmt.Printf("buyer %d: %s [%x] (live=%v, debug=%v)\n", i, buyer.Name, buyer.Id, buyer.Live, buyer.Debug)
	}

	for i, row := range datacenterRows {

		datacenter := db.Datacenter{}

		datacenter.Id = common.DatacenterId(row.name)
		datacenter.Name = row.name
		datacenter.Latitude = row.latitude
		datacenter.Longitude = row.longitude

		seller_row, seller_exists := sellerIndex[row.seller_id]
		if !seller_exists {
			fmt.Printf("error: datacenter %s doesn't have a seller?!\n", datacenter.Name)
			os.Exit(1)
		}

		if !strings.Contains(datacenter.Name, seller_row.name) {
			fmt.Printf("datacenter '%s' does not contain the seller name '%s' as a substring. are you sure this datacenter has the right seller?\n", datacenter.Name, seller_row.name)
			os.Exit(1)
		}

		database.DatacenterMap[datacenter.Id] = datacenter

		fmt.Printf("datacenter %d: %s [%x] (%.1f,%.1f)\n", i, datacenter.Name, datacenter.Id, datacenter.Latitude, datacenter.Longitude)
	}

	for i, row := range relayRows {

		relay := &database.Relays[i]

		relay.Name = row.name

		relay.PublicAddress = ParseAddress(row.public_ip)
		relay.PublicAddress.Port = row.public_port

		relay.Id = common.HashString(relay.PublicAddress.String())

		relay.InternalAddress = ParseAddress(row.internal_ip)
		relay.InternalAddress.Port = row.internal_port

		if relay.InternalAddress.String() != "0.0.0.0:0" {
			relay.HasInternalAddress = true
		}

		relay.SSHAddress = ParseAddress(row.ssh_ip)
		relay.SSHUser = row.ssh_user

		if relay.SSHAddress.String() == "0.0.0.0:0" {
			relay.SSHAddress = relay.PublicAddress
		}

		relay.PublicKey, err = base64.StdEncoding.DecodeString(row.public_key_base64)
		if err != nil {
			fmt.Printf("error: could not decode public key base64 for relay %s: %v\n", relay.Name, err)
			os.Exit(1)
		}
		if len(relay.PublicKey) != 32 {
			fmt.Printf("error: relay public key must be 32 bytes\n")
			os.Exit(1)
		}

		relay.PrivateKey, err = base64.StdEncoding.DecodeString(row.private_key_base64)
		if err != nil {
			fmt.Printf("error: could not decode private key base64 for relay %s: %v\n", relay.Name, err)
		}
		if len(relay.PrivateKey) != 32 {
			fmt.Printf("error: relay private key must be 32 bytes\n")
			os.Exit(1)
		}

		relay.MaxSessions = row.max_sessions
		relay.PortSpeed = row.port_speed
		relay.Version = row.version.String

		datacenter_row, datacenter_exists := datacenterIndex[row.datacenter]
		if !datacenter_exists {
			fmt.Printf("error: relay %s doesn't have a datacenter?!\n", relay.Name)
			os.Exit(1)
		}

		relay.DatacenterId = common.DatacenterId(datacenter_row.name)

		if !strings.Contains(relay.Name, datacenter_row.name) {
			fmt.Printf("error: relay '%s' does not contain the datacenter name '%s' as a substring. are you sure this relay has the right datacenter?\n", relay.Name, datacenter_row.name)
			os.Exit(1)
		}

		relay.Datacenter = database.DatacenterMap[relay.DatacenterId]
		if relay.Datacenter.Id != relay.DatacenterId {
			fmt.Printf("error: relay '%s' has a bad datacenter?!\n", relay.Name)
			os.Exit(1)
		}

		seller_row, seller_exists := sellerIndex[datacenter_row.seller_id]
		if !seller_exists {
			fmt.Printf("error: relay %s doesn't have a seller?!\n", relay.Name)
			os.Exit(1)
		}

		relay.Seller = database.SellerMap[seller_row.id]

		fmt.Printf("relay %d: %s -> %s [%x]\n", i, relay.Name, datacenter_row.name, relay.DatacenterId)

		database.RelayMap[relay.Id] = *relay
	}

	for i, row := range datacenterMapRows {

		buyer_row, buyer_exists := buyerIndex[row.buyer_id]
		if !buyer_exists {
			fmt.Printf("error: datacenter map doesn't have a buyer?!\n")
			os.Exit(1)
		}

		datacenter_row, datacenter_exists := datacenterIndex[row.datacenter_id]
		if !datacenter_exists {
			fmt.Printf("error: datacenter map doesn't have a datacenter?!\n")
			os.Exit(1)
		}

		buyerName := buyer_row.name
		buyerId := uint64(0)
		datacenterName := datacenter_row.name
		datacenterId := common.DatacenterId(datacenterName)

		for _,v := range database.BuyerMap {
			if v.Name == buyerName {
				buyerId = v.Id
			}
		}

		if buyerId == 0 {
			fmt.Printf("error: could not find runtime buyer id for buyer %s?!\n", buyerName)
			os.Exit(1)
		}

		fmt.Printf("datacenter map %d: %s [%x] -> %s [%x] enabled\n", i, buyerName, buyerId, datacenterName, datacenterId)

		datacenterMap := db.DatacenterMap{}
		datacenterMap.EnableAcceleration = row.enable_acceleration
		if database.DatacenterMaps[buyerId] == nil {
			database.DatacenterMaps[buyerId] = make(map[uint64]db.DatacenterMap)
		}
		database.DatacenterMaps[buyerId][datacenterId] = datacenterMap
	}

	// print database

	database.Save("database.bin")

	// load it back in to be sure

	loaded, err := db.LoadDatabase("database.bin")
	if err != nil {
		fmt.Printf("error: could not load database.bin: %v\n", err)
		os.Exit(1)
	}

	_ = loaded
}
