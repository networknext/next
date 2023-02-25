package admin

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
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

type CustomerData struct {
	CustomerId   uint64 `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	CustomerCode string `json:"customer_code"`
	Live         bool   `json:"live"`
	Debug        bool   `json:"debug"`
}

func (controller *Controller) CreateCustomer(customerData *CustomerData) (uint64, error) {
	sql := "INSERT INTO customers (customer_id, customer_name, customer_code, live, debug) VALUES ($1, $2, $3, $4, $5) RETURNING customer_id;"
    result := controller.pgsql.QueryRow(sql, customerData.CustomerId, customerData.CustomerName, customerData.CustomerCode, customerData.Live, customerData.Debug)
    customerId := uint64(0)
	if err := result.Scan(&customerId); err != nil {
		return 0, fmt.Errorf("failed to scan insert customer result: %v\n", err)
	}
	return customerId, nil
}

func (controller *Controller) ReadCustomers() ([]CustomerData, error) {
	customers := make([]CustomerData, 0)
	rows, err := controller.pgsql.Query("SELECT customer_id, customer_name, customer_code, live, debug FROM customers;")
	if err != nil {
		return nil, fmt.Errorf("could not extract customers: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := CustomerData{}
		if err := rows.Scan(&row.CustomerId, &row.CustomerName, &row.CustomerCode, &row.Live, &row.Debug); err != nil {
			return nil, fmt.Errorf("failed to scan customer row: %v\n", err)
		}
		customers = append(customers, row)
	}
	return customers, nil
}

func (controller *Controller) UpdateCustomer(customerData *CustomerData) error {
	// IMPORTANT: Cannot change customer id once created
	sql := "UPDATE customers SET customer_name = $1, customer_code = $2, live = $3, debug = $4 WHERE customer_id = $5;"
	_, err := controller.pgsql.Exec(sql, customerData.CustomerName, customerData.CustomerCode, customerData.Live, customerData.Debug, customerData.CustomerId)
	return err
}

func (controller *Controller) DeleteCustomer(customerId uint64) error {
	sql := "DELETE FROM customers WHERE customer_id = $1;"
	_, err := controller.pgsql.Exec(sql, customerId)
	return err
}

// -----------------------------------------------------------------------

type RouteShaderData struct {
	RouteShaderId             uint64  `json:"route_shader_id"`
	Name                      string  `json:"name"`
	ABTest                    bool    `json:"ab_test"`
	AcceptableLatency         int     `json:"acceptable_latency"`
	AcceptablePacketLoss      float32 `json:"acceptable_packet_loss"`
	PacketLossSustained       float32 `json:"packet_loss_sustained"`
	AnalysisOnly              bool    `json:"analysis_only"`
	BandwidthEnvelopeUpKbps   int     `json:"bandwidth_envelope_up_kbps"`
	BandwidthEnvelopeDownKbps int     `json:"bandwidth_envelope_down_kbps"`
	DisableNetworkNext        bool    `json:"disable_network_next"`
	LatencyThreshold          int     `json:"latency_threshold"`
	Multipath                 bool    `json:"multipath"`
	ReduceLatency             bool    `json:"reduce_latency"`
	ReducePacketLoss          bool    `json:"reduce_packet_loss"`
	SelectionnPercent         float32 `json:"selection_percent"`
	MaxLatencyTradeOff        int     `json:"max_latency_trade_off"`
	MaxNextRTT                int     `json:"max_next_rtt"`
	RouteSwitchThreshold      int     `json:"route_switch_threshold"`
	RouteSelectThreshold      int     `json:"route_select_threshold"`
	RTTVeto_Default           int     `json:"rtt_veto_default"`
	RTTVeto_MultiPath         int     `json:"rtt_veto_multipath"`
	RTTVeto_PacketLoss        int     `json:"rtt_veto_packet_loss"`
	ForceNext                 bool    `json:"force_next"`
	RouteDiversity            int     `json:"route_diversity"`
}

func (controller *Controller) CreateRouteShader(routeShaderData *RouteShaderData) {
	// ...
}

func (controller *Controller) ReadRouteShaders() []RouteShaderData {
	// ...
	return nil
}

func (controller *Controller) UpdateRouteShader(routeShaderData *RouteShaderData) {
	// ...
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
	PublicKeyBase64 string `json:"public_key_base64"`
	CustomerId      uint64 `json:"customer_id"`
	RouteShaderId   uint64 `json:"route_shader_id"`
}

func (controller *Controller) CreateBuyer(buyerData *BuyerData) {
	// ...
}

func (controller *Controller) ReadBuyers() []BuyerData {
	// ...
	return nil
}

func (controller *Controller) UpdateBuyer(buyerData *BuyerData) error {
	// IMPORTANT: Cannot change buyer id once created
	sql := "UPDATE customers SET buyer_name = $1, public_key_base64 = $2, customer_id = $3, route_shader_id = $4 WHERE buyer_id = $5;"
	_, err := controller.pgsql.Exec(sql, buyerData.BuyerName, buyerData.PublicKeyBase64, buyerData.CustomerId, buyerData.RouteShaderId, buyerData.BuyerId)
	return err
}

func (controller *Controller) DeleteBuyer(buyerId uint64) error {
	sql := "DELETE FROM buyers WHERE buyers_id = $1;"
	_, err := controller.pgsql.Exec(sql, buyerId)
	return err
}

// -----------------------------------------------------------------------

type SellerData struct {
	SellerId   uint64 `json:"seller_id"`
	SellerName string `json:"seller_name"`
	CustomerId uint64 `json:"customer_id"`
}

func (controller *Controller) CreateSeller(sellerData *SellerData) {
	// ...
}

func (controller *Controller) ReadSellers() []SellerData {
	// ...
	return nil
}

func (controller *Controller) UpdateSeller(sellerData *SellerData) error {
	// IMPORTANT: Cannot change seller id once created
	sql := "UPDATE customers SET seller_name = $1, customer_id = $2 WHERE seller_id = $3;"
	_, err := controller.pgsql.Exec(sql, sellerData.SellerName, sellerData.CustomerId, sellerData.SellerId)
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
	Latitude       float32 `json:"latitude"`
	Longitude      float32 `json:"longitude"`
	SellerId       uint64  `json:"seller_id"`
	Notes          string  `json:"notes"`
}

func (controller *Controller) CreateDatacenter(datacenterData *SellerData) {
	// ...
}

func (controller *Controller) ReadDatacenters() []DatacenterData {
	// ...
	return nil
}

func (controller *Controller) UpdateDatacenter(datacenterData *DatacenterData) error {
	// IMPORTANT: Cannot change datacenter id once created
	sql := "UPDATE customers SET datacenter_name = $1, latitude = $2, longitude = $3, seller_id = $4, notes = $5 WHERE datacenter_id = $6;"
	_, err := controller.pgsql.Exec(sql, datacenterData.DatacenterName, datacenterData.Latitude, datacenterData.Longitude, datacenterData.SellerId, datacenterData.DatacenterId)
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
	InternalPort     int    `json:"internal_port`
	InternalGroup    string `json:"internal_group`
	SSH_IP           string `json:"ssh_ip"`
	SSH_Port         string `json:"ssh_port`
	SSH_User         string `json:"ssh_user`
	PublicKeyBase64  string `json:"public_key_base64"`
	PrivateKeyBase64 string `json:"private_key_base64"`
	Version          string `json:"version"`
	MRC              int    `json:"mrc"`
	PortSpeed        int    `json:"port_speed"`
	MaxSessions      int    `json:"max_sessions"`
	Notes            string `json:"notes"`
}

func (controller *Controller) CreateRelay(relayData *RelayData) {
	// ...
}

func (controller *Controller) ReadRelays() []RelayData {
	// ...
	return nil
}

func (controller *Controller) UpdateRelay(relayData *RelayData) error {
	// IMPORTANT: Cannot change relay id once created
	sql := `
UPDATE customers 
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
	notes = $17,
WHERE
	relay_id = $18;`
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

func (controller *Controller) CreateBuyerDatacenterSettings(settings *BuyerDatacenterSettings) {
	// ...
}

func (controller *Controller) ReadBuyerDatacenterSettings() []BuyerDatacenterSettings {
	// ...
	return nil
}

func (controller *Controller) UpdateBuyerDatacenterSettings(settings *BuyerDatacenterSettings) error {
	// IMPORTANT: Cannot change buyer id or datacenter id once created
	sql := "UPDATE buyer_datacenter_settings SET enable_acceleration = $1 WHERE buyer_id = $2, datacenter_id = $3;"
	_, err := controller.pgsql.Exec(sql, settings.EnableAcceleration, settings.BuyerId, settings.DatacenterId)
	return err
}

func (controller *Controller) DeleteBuyerDatacenterSettings(buyerId uint64, datacenterId uint64) error {
	sql := "DELETE FROM buyer_datacenter_settings WHERE relay_id = $1, datacenter_id = $2;"
	_, err := controller.pgsql.Exec(sql, buyerId, datacenterId)
	return err
}

// -----------------------------------------------------------------------
