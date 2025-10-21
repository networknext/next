package admin

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/networknext/next/modules/core"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/nacl/box"
)

type Controller struct {
	pgsql *sql.DB
}

func CreateController(config string) *Controller {
	pgsql, err := sql.Open("postgres", config)
	if err != nil {
		panic(fmt.Sprintf("could not connect to postgres: %v", err))
	}
	err = pgsql.Ping()
	if err != nil {
		panic(fmt.Sprintf("could not ping postgres: %v", err))
	}
	fmt.Printf("successfully connected to postgres\n")
	return &Controller{pgsql: pgsql}
}

// -----------------------------------------------------------------------

type RouteShaderData struct {
	RouteShaderId                 uint64  `json:"route_shader_id"`
	RouteShaderName               string  `json:"route_shader_name"`
	ABTest                        bool    `json:"ab_test"`
	AcceptableLatency             int     `json:"acceptable_latency"`
	AcceptablePacketLossInstant   float64 `json:"acceptable_packet_loss_instant"`
	AcceptablePacketLossSustained float64 `json:"acceptable_packet_loss_sustained"`
	BandwidthEnvelopeUpKbps       int     `json:"bandwidth_envelope_up_kbps"`
	BandwidthEnvelopeDownKbps     int     `json:"bandwidth_envelope_down_kbps"`
	DisableNetworkNext            bool    `json:"disable_network_next"`
	LatencyReductionThreshold     int     `json:"latency_reduction_threshold"`
	SelectionPercent              float64 `json:"selection_percent"`
	MaxLatencyTradeOff            int     `json:"max_latency_trade_off"`
	MaxNextRTT                    int     `json:"max_next_rtt"`
	RouteSwitchThreshold          int     `json:"route_switch_threshold"`
	RouteSelectThreshold          int     `json:"route_select_threshold"`
	RTTVeto                       int     `json:"rtt_veto"`
	ForceNext                     bool    `json:"force_next"`
	RouteDiversity                int     `json:"route_diversity"`
}

func (controller *Controller) RouteShaderDefaults() *RouteShaderData {
	routeShader := core.NewRouteShader()
	data := RouteShaderData{}
	data.ABTest = routeShader.ABTest
	data.AcceptableLatency = int(routeShader.AcceptableLatency)
	data.AcceptablePacketLossInstant = float64(routeShader.AcceptablePacketLossInstant)
	data.AcceptablePacketLossSustained = float64(routeShader.AcceptablePacketLossSustained)
	data.BandwidthEnvelopeUpKbps = int(routeShader.BandwidthEnvelopeUpKbps)
	data.BandwidthEnvelopeDownKbps = int(routeShader.BandwidthEnvelopeDownKbps)
	data.DisableNetworkNext = routeShader.DisableNetworkNext
	data.LatencyReductionThreshold = int(routeShader.LatencyReductionThreshold)
	data.SelectionPercent = float64(routeShader.SelectionPercent)
	data.MaxLatencyTradeOff = int(routeShader.MaxLatencyTradeOff)
	data.MaxNextRTT = int(routeShader.MaxNextRTT)
	data.RouteSwitchThreshold = int(routeShader.RouteSwitchThreshold)
	data.RouteSelectThreshold = int(routeShader.RouteSelectThreshold)
	data.RTTVeto = int(routeShader.RTTVeto)
	data.ForceNext = routeShader.ForceNext
	data.RouteDiversity = int(routeShader.RouteDiversity)
	return &data
}

func (controller *Controller) CreateRouteShader(routeShaderData *RouteShaderData) (uint64, error) {
	sql := `
INSERT INTO route_shaders 
(
	route_shader_name,
	ab_test,
	acceptable_latency,
	acceptable_packet_loss_instant,
	acceptable_packet_loss_sustained,
	bandwidth_envelope_up_kbps,
	bandwidth_envelope_down_kbps,
	disable_network_next,
	latency_reduction_threshold,
	selection_percent,
	max_latency_trade_off,
	max_next_rtt,
	route_switch_threshold,
	route_select_threshold,
	rtt_veto,
	force_next,
	route_diversity
)
VALUES
(
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13,
	$14,
	$15,
	$16,
	$17
)
RETURNING route_shader_id;`
	result := controller.pgsql.QueryRow(sql,
		routeShaderData.RouteShaderName,
		routeShaderData.ABTest,
		routeShaderData.AcceptableLatency,
		routeShaderData.AcceptablePacketLossInstant,
		routeShaderData.AcceptablePacketLossSustained,
		routeShaderData.BandwidthEnvelopeUpKbps,
		routeShaderData.BandwidthEnvelopeDownKbps,
		routeShaderData.DisableNetworkNext,
		routeShaderData.LatencyReductionThreshold,
		routeShaderData.SelectionPercent,
		routeShaderData.MaxLatencyTradeOff,
		routeShaderData.MaxNextRTT,
		routeShaderData.RouteSwitchThreshold,
		routeShaderData.RouteSelectThreshold,
		routeShaderData.RTTVeto,
		routeShaderData.ForceNext,
		routeShaderData.RouteDiversity,
	)
	routeShaderId := uint64(0)
	if err := result.Scan(&routeShaderId); err != nil {
		return 0, fmt.Errorf("could not insert route shader: %v\n", err)
	}
	return routeShaderId, nil
}

func (controller *Controller) ReadRouteShaders() ([]RouteShaderData, error) {
	routeShaders := make([]RouteShaderData, 0)
	sql := `
SELECT
	route_shader_id,
	route_shader_name,
	ab_test,
	acceptable_latency,
	acceptable_packet_loss_instant,
	acceptable_packet_loss_sustained,
	bandwidth_envelope_up_kbps,
	bandwidth_envelope_down_kbps,
	disable_network_next,
	latency_reduction_threshold,
	selection_percent,
	max_latency_trade_off,
	max_next_rtt,
	route_switch_threshold,
	route_select_threshold,
	rtt_veto,
	force_next,
	route_diversity
FROM
	route_shaders;`
	rows, err := controller.pgsql.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("could not read route shaders: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := RouteShaderData{}
		err := rows.Scan(
			&row.RouteShaderId,
			&row.RouteShaderName,
			&row.ABTest,
			&row.AcceptableLatency,
			&row.AcceptablePacketLossInstant,
			&row.AcceptablePacketLossSustained,
			&row.BandwidthEnvelopeUpKbps,
			&row.BandwidthEnvelopeDownKbps,
			&row.DisableNetworkNext,
			&row.LatencyReductionThreshold,
			&row.SelectionPercent,
			&row.MaxLatencyTradeOff,
			&row.MaxNextRTT,
			&row.RouteSwitchThreshold,
			&row.RouteSelectThreshold,
			&row.RTTVeto,
			&row.ForceNext,
			&row.RouteDiversity,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan route shader row: %v\n", err)
		}
		routeShaders = append(routeShaders, row)
	}
	return routeShaders, nil
}

func (controller *Controller) ReadRouteShader(routeShaderId uint64) (RouteShaderData, error) {
	routeShader := RouteShaderData{}
	sql := `
SELECT
	route_shader_id,
	route_shader_name,
	ab_test,
	acceptable_latency,
	acceptable_packet_loss_instant,
	acceptable_packet_loss_sustained,
	bandwidth_envelope_up_kbps,
	bandwidth_envelope_down_kbps,
	disable_network_next,
	latency_reduction_threshold,
	selection_percent,
	max_latency_trade_off,
	max_next_rtt,
	route_switch_threshold,
	route_select_threshold,
	rtt_veto,
	force_next,
	route_diversity
FROM
	route_shaders
WHERE
	route_shader_id = $1;`
	rows, err := controller.pgsql.Query(sql, routeShaderId)
	if err != nil {
		return routeShader, fmt.Errorf("could not read route shader: %v\n", err)
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&routeShader.RouteShaderId,
			&routeShader.RouteShaderName,
			&routeShader.ABTest,
			&routeShader.AcceptableLatency,
			&routeShader.AcceptablePacketLossInstant,
			&routeShader.AcceptablePacketLossSustained,
			&routeShader.BandwidthEnvelopeUpKbps,
			&routeShader.BandwidthEnvelopeDownKbps,
			&routeShader.DisableNetworkNext,
			&routeShader.LatencyReductionThreshold,
			&routeShader.SelectionPercent,
			&routeShader.MaxLatencyTradeOff,
			&routeShader.MaxNextRTT,
			&routeShader.RouteSwitchThreshold,
			&routeShader.RouteSelectThreshold,
			&routeShader.RTTVeto,
			&routeShader.ForceNext,
			&routeShader.RouteDiversity,
		)
		if err != nil {
			return routeShader, fmt.Errorf("could not scan route shader row: %v\n", err)
		}
		routeShader.RouteShaderId = routeShaderId
		return routeShader, nil
	}
	return routeShader, fmt.Errorf("route shader %x not found", routeShaderId)
}

func (controller *Controller) UpdateRouteShader(routeShaderData *RouteShaderData) error {
	// IMPORTANT: Cannot change route shader id once created
	sql := `
UPDATE route_shaders 
SET 
	route_shader_name = $1, 
	ab_test = $2,
	acceptable_latency = $3,
	acceptable_packet_loss_instant = $4,
	acceptable_packet_loss_sustained = $5,
	bandwidth_envelope_up_kbps = $6,
	bandwidth_envelope_down_kbps = $7,
	disable_network_next = $8,
	latency_reduction_threshold = $9,
	selection_percent = $10,
	max_latency_trade_off = $11,
	max_next_rtt = $12,
	route_switch_threshold = $13,
	route_select_threshold = $14,
	rtt_veto = $15,
	force_next = $16,
	route_diversity = $17
WHERE
	route_shader_id = $18;`
	_, err := controller.pgsql.Exec(sql,
		routeShaderData.RouteShaderName,
		routeShaderData.ABTest,
		routeShaderData.AcceptableLatency,
		routeShaderData.AcceptablePacketLossInstant,
		routeShaderData.AcceptablePacketLossSustained,
		routeShaderData.BandwidthEnvelopeUpKbps,
		routeShaderData.BandwidthEnvelopeDownKbps,
		routeShaderData.DisableNetworkNext,
		routeShaderData.LatencyReductionThreshold,
		routeShaderData.SelectionPercent,
		routeShaderData.MaxLatencyTradeOff,
		routeShaderData.MaxNextRTT,
		routeShaderData.RouteSwitchThreshold,
		routeShaderData.RouteSelectThreshold,
		routeShaderData.RTTVeto,
		routeShaderData.ForceNext,
		routeShaderData.RouteDiversity,
		routeShaderData.RouteShaderId,
	)
	return err
}

func (controller *Controller) DeleteRouteShader(routeShaderId uint64) error {
	sql := "DELETE FROM route_shaders WHERE route_shader_id = $1;"
	_, err := controller.pgsql.Exec(sql, routeShaderId)
	return err
}

// -----------------------------------------------------------------------

type BuyerData struct {
	BuyerId         uint64 `json:"buyer_id"`
	BuyerName       string `json:"buyer_name"`
	BuyerCode       string `json:"buyer_code"`
	PublicKeyBase64 string `json:"public_key_base64"`
	RouteShaderId   uint64 `json:"route_shader_id"`
	Live            bool   `json:"live"`
	Debug           bool   `json:"debug"`
}

func (controller *Controller) CreateBuyer(buyerData *BuyerData) (uint64, error) {
	// IMPORTANT: Don't allow update of a buyer without a valid buyer public key!!!
	{
		data, err := base64.StdEncoding.DecodeString(buyerData.PublicKeyBase64)
		if err != nil || len(data) != 40 {
			return 0, fmt.Errorf("could not create buyer: invalid public key\n")
		}
	}
	sql := "INSERT INTO buyers (buyer_name, buyer_code, public_key_base64, route_shader_id, live, debug) VALUES ($1, $2, $3, $4, $5, $6) RETURNING buyer_id;"
	result := controller.pgsql.QueryRow(sql, buyerData.BuyerName, buyerData.BuyerCode, buyerData.PublicKeyBase64, buyerData.RouteShaderId, buyerData.Live, buyerData.Debug)
	buyerId := uint64(0)
	if err := result.Scan(&buyerId); err != nil {
		return 0, fmt.Errorf("could not insert buyer: %v\n", err)
	}
	return buyerId, nil
}

func (controller *Controller) ReadBuyers() ([]BuyerData, error) {
	buyers := make([]BuyerData, 0)
	rows, err := controller.pgsql.Query("SELECT buyer_id, buyer_name, buyer_code, public_key_base64, route_shader_id, live, debug FROM buyers;")
	if err != nil {
		return nil, fmt.Errorf("could not read buyers: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := BuyerData{}
		if err := rows.Scan(&row.BuyerId, &row.BuyerName, &row.BuyerCode, &row.PublicKeyBase64, &row.RouteShaderId, &row.Live, &row.Debug); err != nil {
			return nil, fmt.Errorf("could not scan buyer row: %v\n", err)
		}
		buyers = append(buyers, row)
	}
	return buyers, nil
}

func (controller *Controller) ReadBuyer(buyerId uint64) (BuyerData, error) {
	buyer := BuyerData{}
	rows, err := controller.pgsql.Query("SELECT buyer_name, buyer_code, public_key_base64, route_shader_id, live, debug FROM buyers WHERE buyer_id = $1;", buyerId)
	if err != nil {
		return buyer, fmt.Errorf("could not read buyer: %v\n", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&buyer.BuyerName, &buyer.BuyerCode, &buyer.PublicKeyBase64, &buyer.RouteShaderId, &buyer.Live, &buyer.Debug); err != nil {
			return buyer, fmt.Errorf("could not scan buyer row: %v\n", err)
		}
		buyer.BuyerId = buyerId
		return buyer, nil
	}
	return buyer, fmt.Errorf("buyer %x not found", buyerId)
}

func (controller *Controller) UpdateBuyer(buyerData *BuyerData) error {
	// IMPORTANT: Don't allow update of a buyer without a valid buyer public key!!!
	{
		data, err := base64.StdEncoding.DecodeString(buyerData.PublicKeyBase64)
		if err != nil || len(data) != 40 {
			return fmt.Errorf("could not update buyer: invalid public key\n")
		}
	}
	// IMPORTANT: Cannot change buyer id once created
	sql := "UPDATE buyers SET buyer_name = $1, buyer_code = $2, public_key_base64 = $3, route_shader_id = $4, live = $5, debug = $6 WHERE buyer_id = $7;"
	_, err := controller.pgsql.Exec(sql, buyerData.BuyerName, buyerData.BuyerCode, buyerData.PublicKeyBase64, buyerData.RouteShaderId, buyerData.Live, buyerData.Debug, buyerData.BuyerId)
	return err
}

func (controller *Controller) DeleteBuyer(buyerId uint64) error {
	sql := "DELETE FROM buyers WHERE buyer_id = $1;"
	_, err := controller.pgsql.Exec(sql, buyerId)
	return err
}

// -----------------------------------------------------------------------

type SellerData struct {
	SellerId   uint64 `json:"seller_id"`
	SellerName string `json:"seller_name"`
	SellerCode string `json:"seller_code"`
}

func (controller *Controller) CreateSeller(sellerData *SellerData) (uint64, error) {
	var result *sql.Row
	sql := "INSERT INTO sellers (seller_name, seller_code) VALUES ($1, $2) RETURNING seller_id;"
	result = controller.pgsql.QueryRow(sql, sellerData.SellerName, sellerData.SellerCode)
	sellerId := uint64(0)
	if err := result.Scan(&sellerId); err != nil {
		return 0, fmt.Errorf("could not insert seller: %v\n", err)
	}
	return sellerId, nil
}

func (controller *Controller) ReadSellers() ([]SellerData, error) {
	sellers := make([]SellerData, 0)
	rows, err := controller.pgsql.Query("SELECT seller_id, seller_name, seller_code FROM sellers;")
	if err != nil {
		return nil, fmt.Errorf("could not read sellers: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := SellerData{}
		if err := rows.Scan(&row.SellerId, &row.SellerName, &row.SellerCode); err != nil {
			return nil, fmt.Errorf("could not scan seller row: %v\n", err)
		}
		sellers = append(sellers, row)
	}
	return sellers, nil
}

func (controller *Controller) ReadSeller(sellerId uint64) (SellerData, error) {
	seller := SellerData{}
	rows, err := controller.pgsql.Query("SELECT seller_name, seller_code FROM sellers WHERE seller_id = $1;", sellerId)
	if err != nil {
		return seller, fmt.Errorf("could not read seller: %v\n", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&seller.SellerName, &seller.SellerCode); err != nil {
			return seller, fmt.Errorf("could not scan seller row: %v\n", err)
		}
		seller.SellerId = sellerId
		return seller, nil
	}
	return seller, fmt.Errorf("seller %x not found", sellerId)
}

func (controller *Controller) UpdateSeller(sellerData *SellerData) error {
	// IMPORTANT: Cannot change seller id once created
	var err error
	sql := "UPDATE sellers SET seller_name = $1, seller_code = $2 WHERE seller_id = $3;"
	_, err = controller.pgsql.Exec(sql, sellerData.SellerName, sellerData.SellerCode, sellerData.SellerId)
	return err
}

func (controller *Controller) DeleteSeller(sellerId uint64) error {
	sql := "DELETE FROM sellers WHERE seller_id = $1;"
	_, err := controller.pgsql.Exec(sql, sellerId)
	return err
}

// -----------------------------------------------------------------------

type DatacenterData struct {
	DatacenterId   uint64  `json:"datacenter_id"`
	DatacenterName string  `json:"datacenter_name"`
	NativeName     string  `json:"native_name"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	SellerId       uint64  `json:"seller_id"`
	Notes          string  `json:"notes"`
}

func (controller *Controller) CreateDatacenter(datacenterData *DatacenterData) (uint64, error) {
	sql := "INSERT INTO datacenters (datacenter_name, native_name, latitude, longitude, seller_id, notes) VALUES ($1, $2, $3, $4, $5, $6) RETURNING datacenter_id;"
	result := controller.pgsql.QueryRow(sql, datacenterData.DatacenterName, datacenterData.NativeName, datacenterData.Latitude, datacenterData.Longitude, datacenterData.SellerId, datacenterData.Notes)
	datacenterId := uint64(0)
	if err := result.Scan(&datacenterId); err != nil {
		return 0, fmt.Errorf("could not insert datacenter: %v\n", err)
	}
	return datacenterId, nil
}

func (controller *Controller) ReadDatacenters() ([]DatacenterData, error) {
	datacenters := make([]DatacenterData, 0)
	rows, err := controller.pgsql.Query("SELECT datacenter_id, datacenter_name, native_name, latitude, longitude, seller_id, notes FROM datacenters;")
	if err != nil {
		return nil, fmt.Errorf("could not read datacenters: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := DatacenterData{}
		if err := rows.Scan(&row.DatacenterId, &row.DatacenterName, &row.NativeName, &row.Latitude, &row.Longitude, &row.SellerId, &row.Notes); err != nil {
			return nil, fmt.Errorf("could not scan datacenter row: %v\n", err)
		}
		datacenters = append(datacenters, row)
	}
	return datacenters, nil
}

func (controller *Controller) ReadDatacenter(datacenterId uint64) (DatacenterData, error) {
	datacenter := DatacenterData{}
	rows, err := controller.pgsql.Query("SELECT datacenter_name, native_name, latitude, longitude, seller_id, notes FROM datacenters WHERE datacenter_id = $1;", datacenterId)
	if err != nil {
		return datacenter, fmt.Errorf("could not read datacenter: %v\n", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&datacenter.DatacenterName, &datacenter.NativeName, &datacenter.Latitude, &datacenter.Longitude, &datacenter.SellerId, &datacenter.Notes); err != nil {
			return datacenter, fmt.Errorf("could not scan datacenter row: %v\n", err)
		}
		datacenter.DatacenterId = datacenterId
		return datacenter, nil
	}
	return datacenter, fmt.Errorf("datacenter %x not found", datacenterId)
}

func (controller *Controller) UpdateDatacenter(datacenterData *DatacenterData) error {
	// IMPORTANT: Cannot change datacenter id once created
	var err error
	if datacenterData.SellerId != 0 {
		sql := "UPDATE datacenters SET datacenter_name = $1, native_name = $2, latitude = $3, longitude = $4, seller_id = $5, notes = $6 WHERE datacenter_id = $7;"
		_, err = controller.pgsql.Exec(sql, datacenterData.DatacenterName, datacenterData.NativeName, datacenterData.Latitude, datacenterData.Longitude, datacenterData.SellerId, datacenterData.Notes, datacenterData.DatacenterId)
	} else {
		sql := "UPDATE datacenters SET datacenter_name = $1, native_name = $2, latitude = $3, longitude = $4, notes = $5 WHERE datacenter_id = $6;"
		_, err = controller.pgsql.Exec(sql, datacenterData.DatacenterName, datacenterData.NativeName, datacenterData.Latitude, datacenterData.Longitude, datacenterData.Notes, datacenterData.DatacenterId)
	}
	return err
}

func (controller *Controller) DeleteDatacenter(datacenterId uint64) error {
	sql := "DELETE FROM datacenters WHERE datacenter_id = $1;"
	_, err := controller.pgsql.Exec(sql, datacenterId)
	return err
}

// -----------------------------------------------------------------------

type RelayData struct {
	RelayId          uint64 `json:"relay_id"`
	RelayName        string `json:"relay_name"`
	DatacenterId     uint64 `json:"datacenter_id"`
	PublicIP         string `json:"public_ip"`
	PublicPort       int    `json:"public_port"`
	InternalIP       string `json:"internal_ip"`
	InternalPort     int    `json:"internal_port"`
	InternalGroup    string `json:"internal_group"`
	SSH_IP           string `json:"ssh_ip"`
	SSH_Port         int    `json:"ssh_port"`
	SSH_User         string `json:"ssh_user"`
	PublicKeyBase64  string `json:"public_key_base64"`
	PrivateKeyBase64 string `json:"private_key_base64"`
	Version          string `json:"version"`
	MRC              int    `json:"mrc"`
	PortSpeed        int    `json:"port_speed"`
	MaxSessions      int    `json:"max_sessions"`
	BandwidthPrice   int    `json:"bandwidth_price"`
	Notes            string `json:"notes"`
}

func (controller *Controller) CreateRelay(relayData *RelayData) (uint64, error) {
	// IMPORTANT: Don't allow update of a relay without a valid public key!!!
	{
		data, err := base64.StdEncoding.DecodeString(relayData.PublicKeyBase64)
		if err != nil || len(data) != 32 {
			return 0, fmt.Errorf("could not create relay: invalid public key\n")
		}
	}
	// IMPORTANT: Don't allow update of a relay without a valid private key!!!
	{
		data, err := base64.StdEncoding.DecodeString(relayData.PrivateKeyBase64)
		if err != nil || len(data) != 32 {
			return 0, fmt.Errorf("could not create relay: invalid private key\n")
		}
	}
	query := `
INSERT INTO relays 
(
	relay_name,
	datacenter_id,
	public_ip,
	public_port,
	internal_ip,
	internal_port,
	internal_group,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	version,
	mrc,
	port_speed,
	max_sessions,
	bandwidth_price,
	notes
)
VALUES
(
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13,
	$14,
	$15,
	$16,
	$17,
	$18
)
RETURNING relay_id;`
	result := controller.pgsql.QueryRow(query,
		relayData.RelayName,
		relayData.DatacenterId,
		relayData.PublicIP,
		relayData.PublicPort,
		relayData.InternalIP,
		relayData.InternalPort,
		relayData.InternalGroup,
		relayData.SSH_IP,
		relayData.SSH_Port,
		relayData.SSH_User,
		relayData.PublicKeyBase64,
		relayData.PrivateKeyBase64,
		relayData.Version,
		relayData.MRC,
		relayData.PortSpeed,
		relayData.MaxSessions,
		relayData.BandwidthPrice,
		relayData.Notes,
	)
	relayId := uint64(0)
	if err := result.Scan(&relayId); err != nil {
		return 0, fmt.Errorf("could not insert relay: %v\n", err)
	}
	return relayId, nil
}

func (controller *Controller) ReadRelays() ([]RelayData, error) {
	relays := make([]RelayData, 0)
	query := `
SELECT
	relay_id,
	relay_name,
	datacenter_id,
	public_ip,
	public_port,
	internal_ip,
	internal_port,
	internal_group,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	version,
	mrc,
	port_speed,
	max_sessions,
	bandwidth_price,
	notes
FROM
	relays;`
	rows, err := controller.pgsql.Query(query)
	if err != nil {
		return nil, fmt.Errorf("could not read relays: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := RelayData{}
		err := rows.Scan(
			&row.RelayId,
			&row.RelayName,
			&row.DatacenterId,
			&row.PublicIP,
			&row.PublicPort,
			&row.InternalIP,
			&row.InternalPort,
			&row.InternalGroup,
			&row.SSH_IP,
			&row.SSH_Port,
			&row.SSH_User,
			&row.PublicKeyBase64,
			&row.PrivateKeyBase64,
			&row.Version,
			&row.MRC,
			&row.PortSpeed,
			&row.MaxSessions,
			&row.BandwidthPrice,
			&row.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan relay row: %v\n", err)
		}
		relays = append(relays, row)
	}
	return relays, nil
}

func (controller *Controller) ReadRelay(relayId uint64) (RelayData, error) {
	relay := RelayData{}
	query := `
SELECT
	relay_id,
	relay_name,
	datacenter_id,
	public_ip,
	public_port,
	internal_ip,
	internal_port,
	internal_group,
	ssh_ip,
	ssh_port,
	ssh_user,
	public_key_base64,
	private_key_base64,
	version,
	mrc,
	port_speed,
	max_sessions,
	bandwidth_price,
	notes
FROM
	relays
WHERE
	relay_id = $1;`
	rows, err := controller.pgsql.Query(query, relayId)
	if err != nil {
		return relay, fmt.Errorf("could not read relay: %v\n", err)
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&relay.RelayId,
			&relay.RelayName,
			&relay.DatacenterId,
			&relay.PublicIP,
			&relay.PublicPort,
			&relay.InternalIP,
			&relay.InternalPort,
			&relay.InternalGroup,
			&relay.SSH_IP,
			&relay.SSH_Port,
			&relay.SSH_User,
			&relay.PublicKeyBase64,
			&relay.PrivateKeyBase64,
			&relay.Version,
			&relay.MRC,
			&relay.PortSpeed,
			&relay.MaxSessions,
			&relay.BandwidthPrice,
			&relay.Notes,
		)
		if err != nil {
			return relay, fmt.Errorf("could not scan relay row: %v\n", err)
		}
		relay.RelayId = relayId
		return relay, nil
	}
	return relay, fmt.Errorf("relay %x not found", relayId)
}

func (controller *Controller) UpdateRelay(relayData *RelayData) error {
	// IMPORTANT: Don't allow update of a relay without a valid public key!!!
	{
		data, err := base64.StdEncoding.DecodeString(relayData.PublicKeyBase64)
		if err != nil || len(data) != 32 {
			return fmt.Errorf("could not update relay: invalid public key\n")
		}
	}
	// IMPORTANT: Don't allow update of a relay without a valid private key!!!
	{
		data, err := base64.StdEncoding.DecodeString(relayData.PrivateKeyBase64)
		if err != nil || len(data) != 32 {
			return fmt.Errorf("could not update relay: invalid private key\n")
		}
	}
	// IMPORTANT: Cannot change relay id once created
	sql := `
UPDATE relays 
SET 
	relay_name = $1, 
	datacenter_id = $2,
	public_ip = $3,
	public_port = $4,
	internal_ip = $5,
	internal_port = $6,
	internal_group = $7,
	ssh_ip = $8,
	ssh_port = $9,
	ssh_user = $10,
	public_key_base64 = $11,
	private_key_base64 = $12,
	version = $13,
	mrc = $14,
	port_speed = $15,
	max_sessions = $16,
	bandwidth_price = $17,
	notes = $18
WHERE
	relay_id = $19;`
	_, err := controller.pgsql.Exec(sql,
		relayData.RelayName,
		relayData.DatacenterId,
		relayData.PublicIP,
		relayData.PublicPort,
		relayData.InternalIP,
		relayData.InternalPort,
		relayData.InternalGroup,
		relayData.SSH_IP,
		relayData.SSH_Port,
		relayData.SSH_User,
		relayData.PublicKeyBase64,
		relayData.PrivateKeyBase64,
		relayData.Version,
		relayData.MRC,
		relayData.PortSpeed,
		relayData.MaxSessions,
		relayData.BandwidthPrice,
		relayData.Notes,
		relayData.RelayId,
	)
	return err
}

func (controller *Controller) DeleteRelay(relayId uint64) error {
	sql := "DELETE FROM relays WHERE relay_id = $1;"
	_, err := controller.pgsql.Exec(sql, relayId)
	return err
}

// -----------------------------------------------------------------------

type BuyerDatacenterSettings struct {
	BuyerId            uint64 `json:"buyer_id"`
	DatacenterId       uint64 `json:"datacenter_id"`
	EnableAcceleration bool   `json:"enable_acceleration"`
}

func (controller *Controller) CreateBuyerDatacenterSettings(settings *BuyerDatacenterSettings) error {
	sql := "INSERT INTO buyer_datacenter_settings (buyer_id, datacenter_id, enable_acceleration) VALUES ($1, $2, $3);"
	_, err := controller.pgsql.Exec(sql, settings.BuyerId, settings.DatacenterId, settings.EnableAcceleration)
	return err
}

func (controller *Controller) ReadBuyerDatacenterSettingsList() ([]BuyerDatacenterSettings, error) {
	settings := make([]BuyerDatacenterSettings, 0)
	rows, err := controller.pgsql.Query("SELECT buyer_id, datacenter_id, enable_acceleration FROM buyer_datacenter_settings;")
	if err != nil {
		return nil, fmt.Errorf("could not read buyer datacenter settings: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := BuyerDatacenterSettings{}
		if err := rows.Scan(&row.BuyerId, &row.DatacenterId, &row.EnableAcceleration); err != nil {
			return nil, fmt.Errorf("could not scan buyer datacenter settings row: %v\n", err)
		}
		settings = append(settings, row)
	}
	return settings, nil
}

func (controller *Controller) ReadBuyerDatacenterSettings(buyerId uint64, datacenterId uint64) (BuyerDatacenterSettings, error) {
	settings := BuyerDatacenterSettings{}
	rows, err := controller.pgsql.Query("SELECT buyer_id, datacenter_id, enable_acceleration FROM buyer_datacenter_settings WHERE buyer_id = $1 and datacenter_id = $2;", buyerId, datacenterId)
	if err != nil {
		return settings, fmt.Errorf("could not read buyer datacenter settings: %v\n", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&settings.BuyerId, &settings.DatacenterId, &settings.EnableAcceleration); err != nil {
			return settings, fmt.Errorf("could not scan buyer datacenter settings row: %v\n", err)
		}
		return settings, nil
	} else {
		return settings, fmt.Errorf("buyer datacenter settings %x.%x not found", buyerId, datacenterId)
	}
}

func (controller *Controller) UpdateBuyerDatacenterSettings(settings *BuyerDatacenterSettings) error {
	// IMPORTANT: Cannot change buyer id or datacenter id once created
	sql := "UPDATE buyer_datacenter_settings SET enable_acceleration = $1 WHERE buyer_id = $2 AND datacenter_id = $3;"
	_, err := controller.pgsql.Exec(sql, settings.EnableAcceleration, settings.BuyerId, settings.DatacenterId)
	return err
}

func (controller *Controller) DeleteBuyerDatacenterSettings(buyerId uint64, datacenterId uint64) error {
	sql := "DELETE FROM buyer_datacenter_settings WHERE buyer_id = $1 AND datacenter_id = $2;"
	_, err := controller.pgsql.Exec(sql, buyerId, datacenterId)
	return err
}

// -----------------------------------------------------------------------

type RelayKeypairData struct {
	RelayKeypairId   uint64 `json:"relay_keypair_id"`
	PublicKeyBase64  string `json:"public_key_base64"`
	PrivateKeyBase64 string `json:"private_key_base64"`
}

func (controller *Controller) CreateRelayKeypair() (RelayKeypairData, error) {
	relayKeypairId := uint64(0)
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return RelayKeypairData{}, fmt.Errorf("could not generate relay keypair")
	}
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey[:])
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKey[:])
	sql := "INSERT INTO relay_keypairs (public_key_base64, private_key_base64) VALUES ($1, $2) RETURNING relay_keypair_id;"
	result := controller.pgsql.QueryRow(sql, publicKeyBase64, privateKeyBase64)
	if err := result.Scan(&relayKeypairId); err != nil {
		return RelayKeypairData{}, fmt.Errorf("could not insert relay keypair: %v\n", err)
	}
	return RelayKeypairData{RelayKeypairId: relayKeypairId, PublicKeyBase64: publicKeyBase64, PrivateKeyBase64: privateKeyBase64}, nil
}

func (controller *Controller) ReadRelayKeypairs() ([]RelayKeypairData, error) {
	relayKeypairs := make([]RelayKeypairData, 0)
	rows, err := controller.pgsql.Query("SELECT relay_keypair_id, public_key_base64, private_key_base64 FROM relay_keypairs;")
	if err != nil {
		return nil, fmt.Errorf("could not read relay keypairs: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := RelayKeypairData{}
		if err := rows.Scan(&row.RelayKeypairId, &row.PublicKeyBase64, &row.PrivateKeyBase64); err != nil {
			return nil, fmt.Errorf("could not scan relay keypair row: %v\n", err)
		}
		relayKeypairs = append(relayKeypairs, row)
	}
	return relayKeypairs, nil
}

func (controller *Controller) ReadRelayKeypair(relayKeypairId uint64) (RelayKeypairData, error) {
	relayKeypair := RelayKeypairData{}
	rows, err := controller.pgsql.Query("SELECT relay_keypair_id, public_key_base64, private_key_base64 FROM relay_keypairs WHERE relay_keypair_id = $1;", relayKeypairId)
	if err != nil {
		return relayKeypair, fmt.Errorf("could not read relay keypair: %v\n", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&relayKeypair.RelayKeypairId, &relayKeypair.PublicKeyBase64, &relayKeypair.PrivateKeyBase64); err != nil {
			return relayKeypair, fmt.Errorf("could not scan relay keypair row: %v\n", err)
		}
		return relayKeypair, nil
	}
	return relayKeypair, fmt.Errorf("relay keypair %x not found", relayKeypairId)
}

func (controller *Controller) UpdateRelayKeypair(relayKeypairData *RelayKeypairData) error {
	return fmt.Errorf("updating relay keypair is not supported")
}

func (controller *Controller) DeleteRelayKeypair(relayKeypairId uint64) error {
	sql := "DELETE FROM relay_keypairs WHERE relay_keypair_id = $1;"
	_, err := controller.pgsql.Exec(sql, relayKeypairId)
	return err
}

// -----------------------------------------------------------------------
