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
	sql := `INSERT INTO customers (customer_id, customer_name, customer_code, live, debug) VALUES ($1, $2, $3, $4, $5) RETURNING customer_id`
    result, err := controller.pgsql.QueryRow(sql, customerData.CustomerId, customerData.CustomerName, customerData.CustomerCode, customerData.Live, customerData.Debug)
    if err != nil {
    	return 0, fmt.Errorf("failed to insert customer: %v\n", err)
    }
    customerId := uint64(0)
	if err := result.Scan(&customerId); err != nil {
		return 0, fmt.Errorf("failed to scan insert customer result: %v\n", err)
	}
	return customerId, nil
}

func (controller *Controller) ReadCustomers() ([]CustomerData, error) {
	customers := make([]CustomerData, 0)
	rows, err := controller.pgsql.Query("SELECT customer_id, customer_name, customer_code, live, debug FROM customers")
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

func (controller *Controller) UpdateCustomer(customerData *CustomerData) {
	// ...
}

func (controller *Controller) DeleteCustomer(customerId uint64) {
	// ...
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

func (controller *Controller) DeleteRouteShader(routeShaderId uint64) {
	// ...
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

func (controller *Controller) UpdateBuyer(buyerData *BuyerData) {
	// ...
}

func (controller *Controller) DeleteBuyer(buyerId uint64) {
	// ...
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

func (controller *Controller) UpdateSeller(sellerData *SellerData) {
	// ...
}

func (controller *Controller) DeleteSeller(sellerId uint64) {
	// ...
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

func (controller *Controller) UpdateDatacenter(datacenterData *DatacenterData) {
	// ...
}

func (controller *Controller) DeleteDatacenter(datacenterId uint64) {
	// ...
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
	SSHIP            string `json:"ssh_ip"`
	SSHPort          string `json:"ssh_port`
	SSHUser          string `json:"ssh_user`
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

func (controller *Controller) UpdateRelay(relayData *RelayData) {
	// ...
}

func (controller *Controller) DeleteRelay(relayId uint64) {
	// ...
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

func (controller *Controller) UpdateBuyerDatacenterSettings(settings *BuyerDatacenterSettings) {
	// ...
}

func (controller *Controller) DeleteBuyerDatacenterSettings(relayId uint64) {
	// ...
}

// -----------------------------------------------------------------------
