package jsonrpc

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/looker"
	"github.com/networknext/backend/modules/transport/middleware"
)

type OpsService struct {
	Release   string
	BuildTime string

	Env string

	Storage storage.Storer

	LookerClient         *looker.LookerClient
	LookerDashboardCache []looker.LookerDashboard
}

type CurrentReleaseArgs struct{}

type CurrentReleaseReply struct {
	Release   string
	BuildTime string
}

func (s *OpsService) CurrentRelease(r *http.Request, args *CurrentReleaseArgs, reply *CurrentReleaseReply) error {
	reply.Release = s.Release
	reply.BuildTime = s.BuildTime
	return nil
}

type BuyersArgs struct{}

type BuyersReply struct {
	Buyers []buyer `json:"buyers"`
}

type buyer struct {
	CompanyName         string `json:"company_name"`
	CompanyCode         string `json:"company_code"`
	Alias               string `json:"alias"`
	ID                  uint64 `json:"id"`
	HexID               string `json:"hex_id"`
	Live                bool   `json:"live"`
	Debug               bool   `json:"debug"`
	Analytics           bool   `json:"analytics"`
	AnalysisOnly        bool   `json:"analysis_only"`
	Billing             bool   `json:"billing"`
	Trial               bool   `json:"trial"`
	ExoticLocationFee   string `json:"exotic_location_fee"`
	StandardLocationFee string `json:"standard_location_fee"`
	LookerSeats         int64  `json:"looker_seats"`
}

func (s *OpsService) Buyers(r *http.Request, args *BuyersArgs, reply *BuyersReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("Buyers(): %v", err.Error())
		return &err
	}

	for _, b := range s.Storage.Buyers(r.Context()) {
		c, err := s.Storage.Customer(r.Context(), b.CompanyCode)
		if err != nil {
			err = fmt.Errorf("Buyers() could not find Customer %s for %s: %v", b.CompanyCode, b.String(), err)
			core.Error("%v", err)
			return err
		}

		reply.Buyers = append(reply.Buyers, buyer{
			ID:                  b.ID,
			HexID:               b.HexID,
			CompanyName:         c.Name,
			CompanyCode:         b.CompanyCode,
			Alias:               b.Alias,
			Live:                b.Live,
			Debug:               b.Debug,
			Analytics:           b.Analytics,
			AnalysisOnly:        b.RouteShader.AnalysisOnly,
			Billing:             b.Billing,
			Trial:               b.Trial,
			ExoticLocationFee:   fmt.Sprintf("%f", b.ExoticLocationFee),
			StandardLocationFee: fmt.Sprintf("%f", b.StandardLocationFee),
			LookerSeats:         b.LookerSeats,
		})
	}

	sort.Slice(reply.Buyers, func(i int, j int) bool {
		return reply.Buyers[i].CompanyName < reply.Buyers[j].CompanyName
	})

	return nil
}

// Remove this functions and use addNewBuyerAccount VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV

type JSAddBuyerArgs struct {
	CompanyCode         string `json:"company_code"`
	Alias               string `json:"alias"`
	Live                bool   `json:"live"`
	Debug               bool   `json:"debug"`
	Analytics           bool   `json:"analytics"`
	Billing             bool   `json:"billing"`
	Trial               bool   `json:"trial"`
	ExoticLocationFee   string `json:"exoticLocationFee"`
	StandardLocationFee string `json:"standardLocationFee"`
	PublicKey           string `json:"publicKey"`
	LookerSeats         string `json:"looker_seats"`
}

type JSAddBuyerReply struct{}

func (s *OpsService) JSAddBuyer(r *http.Request, args *JSAddBuyerArgs, reply *JSAddBuyerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("JSAddBuyer(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	publicKey, err := base64.StdEncoding.DecodeString(args.PublicKey)
	if err != nil {
		core.Error("%v", err)
		return err
	}

	if len(publicKey) != crypto.KeySize+8 {
		core.Error("%v", err)
		return err
	}

	exoticLocationFee, err := strconv.ParseFloat(args.ExoticLocationFee, 64)
	if err != nil {
		core.Error("%v", err)
		return err
	}

	standardLocationFee, err := strconv.ParseFloat(args.StandardLocationFee, 64)
	if err != nil {
		core.Error("%v", err)
		return err
	}

	lookerSeats, err := strconv.ParseInt(args.LookerSeats, 10, 64)
	if err != nil {
		core.Error("%v", err)
		return err
	}

	// slice the public key here instead of in the clients
	buyer := routing.Buyer{
		CompanyCode:         args.CompanyCode,
		Alias:               args.Alias,
		ID:                  binary.LittleEndian.Uint64(publicKey[:8]),
		Live:                args.Live,
		Debug:               args.Debug,
		Analytics:           args.Analytics,
		Billing:             args.Billing,
		Trial:               args.Trial,
		ExoticLocationFee:   exoticLocationFee,
		StandardLocationFee: standardLocationFee,
		PublicKey:           publicKey[8:],
		LookerSeats:         lookerSeats,
	}

	return s.Storage.AddBuyer(ctx, buyer)
}

// Remove this functions and use addNewBuyerAccount ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type AddNewBuyerAccountArgs struct {
	ParentCustomerCode  string  `json:"customer_code"`
	PublicKey           string  `json:"public_key"`
	ExoticLocationFee   float32 `json:"exotic_location_fee"`
	StandardLocationFee float32 `json:"standard_location_fee"`
	LookerSeats         int32   `json:"looker_seats"`
	Live                bool    `json:"live"`
	Debug               bool    `json:"debug"`
	Trial               bool    `json:"trial"`
	Billing             bool    `json:"billing"`
	Analytics           bool    `json:"analytics"`
	Advanced            bool    `json:"advanced"`
}

type AddNewBuyerAccountReply struct{}

func (s *OpsService) AddNewBuyerAccount(r *http.Request, args *AddNewBuyerAccountArgs, reply *AddNewBuyerAccountReply) error {
	ctx := r.Context()

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AddNewBuyerAccount(): %v", err.Error())
		return &err
	}

	if args.ParentCustomerCode == "" {
		err := fmt.Errorf("AddNewBuyerAccount(): Parent customer code is required")
		core.Error("%v", err)
		return err
	}

	// Check if a buyer is assigned to this customer already
	buyer, err := s.Storage.BuyerWithCompanyCode(ctx, args.ParentCustomerCode)
	if buyer.CompanyCode != "" || buyer.ID != 0 {
		err := fmt.Errorf("AddNewBuyerAccount() Customer account is already assigned to a buyer")
		core.Error("%v", err)
		return err
	}

	if args.PublicKey == "" {
		err := fmt.Errorf("AddNewBuyerAccount() A public key is required")
		core.Error("%v", err)
		return err
	}

	byteKey, err := base64.StdEncoding.DecodeString(args.PublicKey)
	if err != nil {
		err = fmt.Errorf("AddNewBuyerAccount() Failed to decode public key string")
		core.Error("%v", err)
		return err
	}

	buyerID := binary.LittleEndian.Uint64(byteKey[0:8])

	// Create new buyer
	err = s.Storage.AddBuyer(ctx, routing.Buyer{
		CompanyCode: args.ParentCustomerCode,
		ID:          buyerID,
		Live:        args.Live,
		Analytics:   args.Analytics,
		Billing:     args.Billing,
		Trial:       args.Trial,
		Debug:       args.Debug,
		// Advanced: args.Advanced,
		PublicKey:           byteKey[8:],
		LookerSeats:         int64(args.LookerSeats),
		ExoticLocationFee:   float64(args.ExoticLocationFee),
		StandardLocationFee: float64(args.StandardLocationFee),
	})
	if err != nil {
		err = fmt.Errorf("AddNewBuyerAccount() Failed to add new buyer account: %v", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type UpdateBuyerAccountArgs struct {
	Alias               string  `json:"alias"`
	CustomerCode        string  `json:"customer_code"`
	ExoticLocationFee   float32 `json:"exotic_location_fee"`
	StandardLocationFee float32 `json:"standard_location_fee"`
	LookerSeats         int32   `json:"looker_seats"`
	Live                bool    `json:"live"`
	Debug               bool    `json:"debug"`
	Trial               bool    `json:"trial"`
	Billing             bool    `json:"billing"`
	Analytics           bool    `json:"analytics"`
	Advanced            bool    `json:"advanced"`
}

type UpdateBuyerAccountReply struct{}

func (s *OpsService) UpdateBuyerAccount(r *http.Request, args *UpdateBuyerAccountArgs, reply *UpdateBuyerAccountReply) error {
	ctx := r.Context()

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateBuyerAccount(): %v", err.Error())
		return &err
	}

	buyer, err := s.Storage.BuyerWithCompanyCode(ctx, args.CustomerCode)
	if err != nil {
		core.Error("UpdateBuyerAccount(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	// TODO: Update functions should be using database ID here

	wasError := false
	if buyer.ExoticLocationFee != float64(args.ExoticLocationFee) && (args.ExoticLocationFee >= 0 && args.ExoticLocationFee < 100000) {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "ExoticLocationFee", float64(args.ExoticLocationFee)); err != nil {
			core.Error("UpdateBuyerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.StandardLocationFee != float64(args.StandardLocationFee) && (args.StandardLocationFee >= 0 && args.StandardLocationFee < 100000) {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "StandardLocationFee", float64(args.StandardLocationFee)); err != nil {
			core.Error("UpdateBuyerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.LookerSeats != int64(args.LookerSeats) && (args.LookerSeats >= 0 && args.LookerSeats < 1000) {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "LookerSeats", int64(args.LookerSeats)); err != nil {
			core.Error("UpdateBuyerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.Live != args.Live {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "Live", args.Live); err != nil {
			core.Error("UpdateBuyerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.Debug != args.Debug {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "Debug", args.Debug); err != nil {
			core.Error("UpdateBuyerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.Analytics != args.Analytics {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "Analytics", args.Analytics); err != nil {
			core.Error("UpdateBuyerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.Billing != args.Billing {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "Billing", args.Billing); err != nil {
			core.Error("UpdateBuyerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.Trial != args.Trial {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "Trial", args.Trial); err != nil {
			core.Error("UpdateBuyerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.Alias != args.Alias && args.Alias != "" {
		if err := s.Storage.UpdateBuyer(ctx, buyer.ID, "Alias", args.Alias); err != nil {
			core.Error("UpdateBuyerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if wasError {
		core.Error("UpdateBuyerAccount(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type RemoveBuyerArgs struct {
	ID string
}

type RemoveBuyerReply struct{}

func (s *OpsService) RemoveBuyer(r *http.Request, args *RemoveBuyerArgs, reply *RemoveBuyerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RemoveBuyer(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	buyerID, err := strconv.ParseUint(args.ID, 16, 64)
	if err != nil {
		err = fmt.Errorf("RemoveBuyer() could not convert buyer ID %s to uint64: %v", args.ID, err)
		core.Error("%v", err)
		return err
	}

	return s.Storage.RemoveBuyer(ctx, buyerID)
}

type buyerDatacenterMap struct {
	Alias string `json:"alias"`
	Name  string `json:"name"`
	HexID string `json:"hex_id"`
}

type FetchBuyerInformationArgs struct {
	CustomerCode string `json:"customer_code"`
}

type FetchBuyerInformationReply struct {
	ID                  string               `json:"id"`
	Alias               string               `json:"alias"`
	Advanced            bool                 `json:"advanced"`
	Live                bool                 `json:"live"`
	Debug               bool                 `json:"debug"`
	Analytics           bool                 `json:"analytics"`
	Billing             bool                 `json:"billing"`
	Trial               bool                 `json:"trial"`
	ExoticLocationFee   float32              `json:"exotic_location_fee"`
	StandardLocationFee float32              `json:"standard_location_fee"`
	PublicKey           string               `json:"public_key"`
	LookerSeats         int32                `json:"looker_seats"`
	RouteShader         core.RouteShader     `json:"route_shader"`
	InternalConfig      core.InternalConfig  `json:"internal_config"`
	MappedDatacenters   []buyerDatacenterMap `json:"mapped_datacenters"`
}

func (s *OpsService) FetchBuyerInformation(r *http.Request, args *FetchBuyerInformationArgs, reply *FetchBuyerInformationReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("FetchBuyerInformation(): %v", err.Error())
		return &err
	}

	if args.CustomerCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "CustomerCode"
		core.Error("FetchBuyerInformation(): %v", err.Error())
		return &err
	}

	ctx := r.Context()

	buyer, err := s.Storage.BuyerWithCompanyCode(ctx, args.CustomerCode)
	if err == nil {
		reply.ID = buyer.HexID
		reply.Alias = buyer.Alias
		// reply.Advanced = buyer.Advanced
		reply.Live = buyer.Live
		reply.Debug = buyer.Debug
		reply.Analytics = buyer.Analytics
		reply.Billing = buyer.Billing
		reply.Trial = buyer.Trial
		reply.ExoticLocationFee = float32(buyer.ExoticLocationFee)
		reply.StandardLocationFee = float32(buyer.StandardLocationFee)
		reply.PublicKey = buyer.EncodedPublicKey()
		reply.LookerSeats = int32(buyer.LookerSeats)

		reply.RouteShader = buyer.RouteShader
		reply.InternalConfig = buyer.InternalConfig

		reply.MappedDatacenters = make([]buyerDatacenterMap, 0)

		buyerMaps := s.Storage.GetDatacenterMapsForBuyer(ctx, buyer.ID)
		datacenters := s.Storage.Datacenters(ctx)

		for _, datacenter := range datacenters {
			_, ok := buyerMaps[datacenter.ID]
			if ok {
				reply.MappedDatacenters = append(reply.MappedDatacenters, buyerDatacenterMap{
					Alias: datacenter.AliasName,
					Name:  datacenter.Name,
					HexID: fmt.Sprintf("%016x", datacenter.ID),
				})
			}
		}
	} else {
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		core.Error("FetchBuyerInformation(): %v", err.Error())
		return &err
	}

	return nil
}

type UpdateBuyerRouteShaderArgs struct {
	CustomerCode string           `json:"customer_code"`
	RouteShader  core.RouteShader `json:"route_shader"`
}

type UpdateBuyerRouteShaderReply struct{}

func (s *OpsService) UpdateBuyerRouteShader(r *http.Request, args *UpdateBuyerRouteShaderArgs, reply *UpdateBuyerRouteShaderReply) error {
	ctx := r.Context()

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateBuyerRouteShader(): %v", err.Error())
		return &err
	}

	buyer, err := s.Storage.BuyerWithCompanyCode(ctx, args.CustomerCode)
	if err != nil {
		core.Error("UpdateBuyerRouteShader(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	// Check if the buyer actually has a route shader
	if _, err := s.Storage.RouteShader(ctx, buyer.ID); err != nil {
		if err := s.Storage.AddRouteShader(ctx, core.NewRouteShader(), buyer.ID); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}
	}

	// TODO: Update functions should be using database ID here

	wasError := false
	if buyer.RouteShader.ABTest != args.RouteShader.ABTest {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "ABTest", args.RouteShader.ABTest); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.DisableNetworkNext != args.RouteShader.DisableNetworkNext {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "DisableNetworkNext", args.RouteShader.DisableNetworkNext); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.AnalysisOnly != args.RouteShader.AnalysisOnly {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "AnalysisOnly", args.RouteShader.AnalysisOnly); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.Multipath != args.RouteShader.Multipath {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "Multipath", args.RouteShader.Multipath); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.ProMode != args.RouteShader.ProMode {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "ProMode", args.RouteShader.ProMode); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.ReduceLatency != args.RouteShader.ReduceLatency {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "ReduceLatency", args.RouteShader.ReduceLatency); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.ReducePacketLoss != args.RouteShader.ReducePacketLoss {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "ReducePacketLoss", args.RouteShader.ReducePacketLoss); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.ReduceJitter != args.RouteShader.ReduceJitter {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "ReduceJitter", args.RouteShader.ReduceJitter); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.AcceptableLatency != args.RouteShader.AcceptableLatency && (args.RouteShader.AcceptableLatency >= 0 && args.RouteShader.AcceptableLatency < 1024) {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "AcceptableLatency", args.RouteShader.AcceptableLatency); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.AcceptablePacketLoss != args.RouteShader.AcceptablePacketLoss && (args.RouteShader.AcceptablePacketLoss >= 0 && args.RouteShader.AcceptablePacketLoss <= 100) {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "AcceptablePacketLoss", args.RouteShader.AcceptablePacketLoss); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.BandwidthEnvelopeUpKbps != args.RouteShader.BandwidthEnvelopeUpKbps && (args.RouteShader.BandwidthEnvelopeUpKbps >= 0 && args.RouteShader.BandwidthEnvelopeUpKbps < 10000) {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "BandwidthEnvelopeUpKbps", args.RouteShader.BandwidthEnvelopeUpKbps); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.BandwidthEnvelopeDownKbps != args.RouteShader.BandwidthEnvelopeDownKbps && (args.RouteShader.BandwidthEnvelopeDownKbps >= 0 && args.RouteShader.BandwidthEnvelopeDownKbps < 10000) {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "BandwidthEnvelopeDownKbps", args.RouteShader.BandwidthEnvelopeDownKbps); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.LatencyThreshold != args.RouteShader.LatencyThreshold && (args.RouteShader.LatencyThreshold >= 0 && args.RouteShader.LatencyThreshold < 1024) {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "LatencyThreshold", args.RouteShader.LatencyThreshold); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.PacketLossSustained != args.RouteShader.PacketLossSustained && (args.RouteShader.PacketLossSustained >= 0 && args.RouteShader.PacketLossSustained <= 100) {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "PacketLossSustained", args.RouteShader.PacketLossSustained); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.RouteShader.SelectionPercent != args.RouteShader.SelectionPercent && (args.RouteShader.SelectionPercent >= 0 && args.RouteShader.SelectionPercent <= 100) {
		if err := s.Storage.UpdateRouteShader(ctx, buyer.ID, "SelectionPercent", args.RouteShader.SelectionPercent); err != nil {
			core.Error("UpdateBuyerRouteShader(): %v", err.Error())
			wasError = true
		}
	}

	if wasError {
		core.Error("UpdateBuyerRouteShader(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type UpdateBuyerInternalConfigArgs struct {
	CustomerCode   string              `json:"customer_code"`
	InternalConfig core.InternalConfig `json:"internal_config"`
}

type UpdateBuyerInternalConfigReply struct{}

func (s *OpsService) UpdateBuyerInternalConfig(r *http.Request, args *UpdateBuyerInternalConfigArgs, reply *UpdateBuyerInternalConfigReply) error {
	ctx := r.Context()

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
		return &err
	}

	buyer, err := s.Storage.BuyerWithCompanyCode(ctx, args.CustomerCode)
	if err != nil {
		core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	// TODO: Update functions should be using database ID here

	wasError := false
	if buyer.InternalConfig.EnableVanityMetrics != args.InternalConfig.EnableVanityMetrics {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "EnableVanityMetrics", args.InternalConfig.EnableVanityMetrics); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.ForceNext != args.InternalConfig.ForceNext {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "ForceNext", args.InternalConfig.ForceNext); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.LargeCustomer != args.InternalConfig.LargeCustomer {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "LargeCustomer", args.InternalConfig.LargeCustomer); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.HighFrequencyPings != args.InternalConfig.HighFrequencyPings {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "HighFrequencyPings", args.InternalConfig.HighFrequencyPings); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.Uncommitted != args.InternalConfig.Uncommitted {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "Uncommitted", args.InternalConfig.Uncommitted); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.TryBeforeYouBuy != args.InternalConfig.TryBeforeYouBuy {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "TryBeforeYouBuy", args.InternalConfig.TryBeforeYouBuy); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.MaxLatencyTradeOff != args.InternalConfig.MaxLatencyTradeOff && (args.InternalConfig.MaxLatencyTradeOff >= 0 && args.InternalConfig.MaxLatencyTradeOff < 1024) {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "MaxLatencyTradeOff", args.InternalConfig.MaxLatencyTradeOff); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.MaxRTT != args.InternalConfig.MaxRTT && (args.InternalConfig.MaxRTT >= 0 && args.InternalConfig.MaxRTT < 1024) {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "MaxRTT", args.InternalConfig.MaxRTT); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.ReducePacketLossMinSliceNumber != args.InternalConfig.ReducePacketLossMinSliceNumber && (args.InternalConfig.ReducePacketLossMinSliceNumber >= 0) {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "ReducePacketLossMinSliceNumber", args.InternalConfig.ReducePacketLossMinSliceNumber); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.RouteDiversity != args.InternalConfig.RouteDiversity && (args.InternalConfig.RouteDiversity >= 0) {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "RouteDiversity", args.InternalConfig.RouteDiversity); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.RouteSelectThreshold != args.InternalConfig.RouteSelectThreshold && (args.InternalConfig.RouteSelectThreshold >= 0) {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "RouteSelectThreshold", args.InternalConfig.RouteSelectThreshold); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.RouteSwitchThreshold != args.InternalConfig.RouteSwitchThreshold && (args.InternalConfig.RouteSwitchThreshold >= 0) {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "RouteSwitchThreshold", args.InternalConfig.RouteSwitchThreshold); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.MultipathOverloadThreshold != args.InternalConfig.MultipathOverloadThreshold && (args.InternalConfig.MultipathOverloadThreshold >= 0) {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "MultipathOverloadThreshold", args.InternalConfig.MultipathOverloadThreshold); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.RTTVeto_Default != args.InternalConfig.RTTVeto_Default {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "RTTVeto_Default", args.InternalConfig.RTTVeto_Default); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.RTTVeto_Multipath != args.InternalConfig.RTTVeto_Multipath {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "RTTVeto_Multipath", args.InternalConfig.RTTVeto_Multipath); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if buyer.InternalConfig.RTTVeto_PacketLoss != args.InternalConfig.RTTVeto_PacketLoss {
		if err := s.Storage.UpdateInternalConfig(ctx, buyer.ID, "RTTVeto_PacketLoss", args.InternalConfig.RTTVeto_PacketLoss); err != nil {
			core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
			wasError = true
		}
	}

	if wasError {
		core.Error("UpdateBuyerInternalConfig(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type SellersArgs struct{}

type SellersReply struct {
	Sellers []seller
}

type seller struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	EgressPriceNibblins routing.Nibblin `json:"egressPriceNibblins"`
	Secret              bool            `json:"secret"`
}

func (s *OpsService) Sellers(r *http.Request, args *SellersArgs, reply *SellersReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("Sellers(): %v", err.Error())
		return &err
	}

	for _, localSeller := range s.Storage.Sellers(r.Context()) {

		reply.Sellers = append(reply.Sellers, seller{
			ID:                  localSeller.ID,
			Name:                localSeller.Name,
			EgressPriceNibblins: localSeller.EgressPriceNibblinsPerGB,
			Secret:              localSeller.Secret,
		})
	}

	sort.Slice(reply.Sellers, func(i int, j int) bool {
		return reply.Sellers[i].Name < reply.Sellers[j].Name
	})

	return nil
}

type CustomersArgs struct{}

type CustomersReply struct {
	Customers []customer `json:"customers"`
}

type customer struct {
	Name                   string `json:"name"`
	Code                   string `json:"code"`
	AutomaticSignInDomains string `json:"automaticSigninDomains"`
	BuyerID                string `json:"buyer_id"`
	Buyer                  buyer  `json:"buyer,omitempty"`
	SellerID               string `json:"seller_id"`
	Seller                 seller `json:"seller,omitempty"`
}

func (s *OpsService) Customers(r *http.Request, args *CustomersArgs, reply *CustomersReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("Customers(): %v", err.Error())
		return &err
	}

	customers := s.Storage.Customers(r.Context())
	for _, c := range customers {

		buyerID := ""

		// TODO both of these support functions should be
		// removed or modified to check by FK
		buyer, _ := s.Storage.BuyerWithCompanyCode(r.Context(), c.Code)
		seller, _ := s.Storage.SellerWithCompanyCode(r.Context(), c.Code)

		if buyer.ID != 0 {
			buyerID = fmt.Sprintf("%016x", buyer.ID)
		}

		customerEntry := customer{
			Name:                   c.Name,
			Code:                   c.Code,
			AutomaticSignInDomains: c.AutomaticSignInDomains,
			BuyerID:                buyerID,
			SellerID:               seller.ID,
		}

		if buyerID != "" {
			customerEntry.Buyer.Analytics = buyer.Analytics
			customerEntry.Buyer.AnalysisOnly = buyer.RouteShader.AnalysisOnly
			customerEntry.Buyer.Billing = buyer.Billing
			customerEntry.Buyer.Debug = buyer.Debug
			customerEntry.Buyer.Live = buyer.Live
		}

		reply.Customers = append(reply.Customers, customerEntry)
	}

	sort.Slice(reply.Customers, func(i int, j int) bool {
		return reply.Customers[i].Name < reply.Customers[j].Name
	})
	return nil
}

// TODO: Remove these functions and use AddNewCustomerAccount VVVVVVVVVV

type JSAddCustomerArgs struct {
	Code                   string `json:"code"`
	Name                   string `json:"name"`
	AutomaticSignInDomains string `json:"automaticSignInDomains"`
}

type JSAddCustomerReply struct{}

func (s *OpsService) JSAddCustomer(r *http.Request, args *JSAddCustomerArgs, reply *JSAddCustomerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("JSAddCustomer(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	customer := routing.Customer{
		Name:                   args.Name,
		Code:                   args.Code,
		AutomaticSignInDomains: args.AutomaticSignInDomains,
	}

	if err := s.Storage.AddCustomer(ctx, customer); err != nil {
		err = fmt.Errorf("AddCustomer() error: %w", err)
		core.Error("%v", err)
		return err
	}
	return nil
}

type AddCustomerArgs struct {
	Customer routing.Customer
}

type AddCustomerReply struct{}

func (s *OpsService) AddCustomer(r *http.Request, args *AddCustomerArgs, reply *AddCustomerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AddCustomer(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddCustomer(ctx, args.Customer); err != nil {
		err = fmt.Errorf("AddCustomer() error: %w", err)
		core.Error("%v", err)
		return err
	}
	return nil
}

// TODO: Remove these functions and use AddNewCustomerAccount ^^^^^^^^^^^^^^

type AddNewCustomerAccountArgs struct {
	Name    string   `json:"name"`
	Code    string   `json:"code"`
	Domains []string `json:"domains"`
}

type AddNewCustomerAccountReply struct{}

func (s *OpsService) AddNewCustomerAccount(r *http.Request, args *AddNewCustomerAccountArgs, reply *AddNewCustomerAccountReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AddNewCustomerAccount(): %v", err.Error())
		return &err
	}

	if args.Name == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Name"
		core.Error("AddNewCustomerAccount(): %v", err.Error())
		return &err
	}

	if args.Code == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Code"
		core.Error("AddNewCustomerAccount(): %v", err.Error())
		return &err
	}

	customer := routing.Customer{
		Name:                   args.Name,
		Code:                   args.Code,
		AutomaticSignInDomains: strings.Join(args.Domains, ", "),
	}

	if err := s.Storage.AddCustomer(r.Context(), customer); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		core.Error("AddNewCustomerAccount(): %v", err.Error())
		return &err
	}

	return nil
}

type UpdateCustomerAccountArgs struct {
	ID      int32    `json:"id"`
	Name    string   `json:"name"`
	Domains []string `json:"domains"`
}

type UpdateCustomerAccountReply struct{}

func (s *OpsService) UpdateCustomerAccount(r *http.Request, args *UpdateCustomerAccountArgs, reply *UpdateCustomerAccountReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateNewCustomerAccount(): %v", err.Error())
		return &err
	}

	ctx := r.Context()

	customer, err := s.Storage.CustomerByID(ctx, int64(args.ID))
	if err != nil {
		core.Error("UpdateCustomerAccount(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err

	}

	// TODO: Figure out if we want to support changing customer code - would require deleting and reconstructing customer account
	// TODO: These update functions really should be using the database ID

	wasError := false
	if customer.Name != args.Name && args.Name != "" {
		if err := s.Storage.UpdateCustomer(ctx, customer.Code, "Name", args.Name); err != nil {
			core.Error("UpdateCustomerAccount(): %v", err.Error())
			wasError = true
		}
	}

	domains := strings.Join(args.Domains, ", ")

	if customer.AutomaticSignInDomains != domains {
		if err := s.Storage.UpdateCustomer(ctx, customer.Code, "AutomaticSigninDomains", domains); err != nil {
			core.Error("UpdateCustomerAccount(): %v", err.Error())
			wasError = true
		}
	}

	if wasError {
		core.Error("UpdateCustomerAccount(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type FetchCustomerInformationArgs struct {
	CustomerCode string `json:"customer_code"`
}

type FetchCustomerInformationReply struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	BuyerID       string `json:"buyer_id"`
	IsLive        bool   `json:"is_live"`
	AnalyticsOnly bool   `json:"analytics_only"`
	HasAnalytics  bool   `json:"premium_analytics"`
	HasBilling    bool   `json:"show_billing"`
	PublicKey     string `json:"public_key"`
	SellerID      string `json:"seller_id"`
	DatabaseID    int32  `json:"id"`
	Domains       string `json:"domains"`
}

func (s *OpsService) FetchCustomerInformation(r *http.Request, args *FetchCustomerInformationArgs, reply *FetchCustomerInformationReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("Customer(): %v", err.Error())
		return &err
	}

	ctx := r.Context()

	customer, err := s.Storage.Customer(ctx, args.CustomerCode)
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		core.Error("FetchCustomerInformation(): %v", err.Error())
		return &err
	}

	reply.Name = customer.Name
	reply.Code = customer.Code
	reply.DatabaseID = int32(customer.DatabaseID)
	reply.Domains = customer.AutomaticSignInDomains

	buyer, err := s.Storage.BuyerWithCompanyCode(ctx, customer.Code)
	if err == nil && buyer.CompanyCode != "" {
		reply.BuyerID = buyer.HexID
		reply.IsLive = buyer.Live
		reply.AnalyticsOnly = buyer.RouteShader.AnalysisOnly
		reply.PublicKey = buyer.EncodedPublicKey()
		reply.HasAnalytics = buyer.Analytics
		reply.HasBilling = buyer.Billing
	}

	seller, err := s.Storage.SellerWithCompanyCode(ctx, customer.Code)
	if err == nil && seller.CompanyCode != "" {
		reply.SellerID = seller.CompanyCode
	}

	return nil
}

type CustomerArg struct {
	CustomerID string
}

type CustomerReply struct {
	Customer routing.Customer
}

func (s *OpsService) Customer(r *http.Request, arg *CustomerArg, reply *CustomerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("Customer(): %v", err.Error())
		return &err
	}

	var c routing.Customer
	var err error

	if c, err = s.Storage.Customer(r.Context(), arg.CustomerID); err != nil {
		err = fmt.Errorf("Customer() error: %w", err)
		core.Error("%v", err)
		return err
	}
	reply.Customer = c

	return nil
}

type SellerArg struct {
	ID string
}

type SellerReply struct {
	Seller routing.Seller
}

func (s *OpsService) Seller(r *http.Request, arg *SellerArg, reply *SellerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("Seller(): %v", err.Error())
		return &err
	}

	var seller routing.Seller
	var err error
	if seller, err = s.Storage.Seller(r.Context(), arg.ID); err != nil {
		err = fmt.Errorf("Seller() error: %w", err)
		core.Error("%v", err)
		return err
	}

	reply.Seller = seller
	return nil

}

type JSAddSellerArgs struct {
	ShortName    string `json:"shortName"`
	Secret       bool   `json:"secret"`
	IngressPrice int64  `json:"ingressPrice"` // nibblins
	EgressPrice  int64  `json:"egressPrice"`  // nibblins
}

type JSAddSellerReply struct{}

func (s *OpsService) JSAddSeller(r *http.Request, args *JSAddSellerArgs, reply *JSAddSellerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("JSAddSeller(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	seller := routing.Seller{
		ID:                       args.ShortName,
		ShortName:                args.ShortName,
		CompanyCode:              args.ShortName,
		Secret:                   args.Secret,
		EgressPriceNibblinsPerGB: routing.Nibblin(args.EgressPrice),
	}

	if err := s.Storage.AddSeller(ctx, seller); err != nil {
		err = fmt.Errorf("AddSeller() error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type AddSellerArgs struct {
	Seller routing.Seller
}

type AddSellerReply struct{}

func (s *OpsService) AddSeller(r *http.Request, args *AddSellerArgs, reply *AddSellerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AddSeller(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddSeller(ctx, args.Seller); err != nil {
		err = fmt.Errorf("AddSeller() error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type RemoveSellerArgs struct {
	ID string
}

type RemoveSellerReply struct{}

func (s *OpsService) RemoveSeller(r *http.Request, args *RemoveSellerArgs, reply *RemoveSellerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RemoveSeller(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.RemoveSeller(ctx, args.ID); err != nil {
		err = fmt.Errorf("RemoveSeller() error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type SetCustomerLinkArgs struct {
	CustomerName string
	BuyerID      uint64
	SellerID     string
}

type SetCustomerLinkReply struct{}

func (s *OpsService) SetCustomerLink(r *http.Request, args *SetCustomerLinkArgs, reply *SetCustomerLinkReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("SetCustomerLink(): %v", err.Error())
		return &err
	}

	if args.CustomerName == "" {
		err := errors.New("SetCustomerLink() error: customer name empty")
		core.Error("%v", err)
		return err
	}

	if args.BuyerID == 0 && args.SellerID == "" {
		err := errors.New("SetCustomerLink() error: invalid paramters - both buyer ID and seller ID are empty")
		core.Error("%v", err)
		return err
	}

	if args.BuyerID != 0 && args.SellerID != "" {
		err := errors.New("SetCustomerLink() error: invalid paramters - both buyer ID and seller ID are given which is not allowed")
		core.Error("%v", err)
		return err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	buyerID := args.BuyerID
	sellerID := args.SellerID

	if buyerID != 0 {
		// We're trying to update the link to the buyer ID, so get the existing seller ID so it doesn't change
		var err error
		sellerID, err = s.Storage.SellerIDFromCustomerName(ctx, args.CustomerName)
		if err != nil {
			err = fmt.Errorf("SetCustomerLink() error: %w", err)
			core.Error("%v", err)
			return err
		}
	}

	if sellerID != "" {
		// We're trying to update the link to the seller ID, so get the existing buyer ID so it doesn't change
		var err error
		buyerID, err = s.Storage.BuyerIDFromCustomerName(ctx, args.CustomerName)
		if err != nil {
			err = fmt.Errorf("SetCustomerLink() error: %w", err)
			core.Error("%v", err)
			return err
		}
	}

	if err := s.Storage.SetCustomerLink(ctx, args.CustomerName, buyerID, sellerID); err != nil {
		err = fmt.Errorf("SetCustomerLink() error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type RelaysArgs struct {
	Regex string `json:"name"`
}

type RelaysReply struct {
	Relays []relay `json:"relays"`
}

type relay struct {
	ID                            uint64                `json:"id"`
	HexID                         string                `json:"hexID"`
	DatacenterHexID               string                `json:"datacenterHexID"`
	BillingSupplier               string                `json:"billingSupplier"`
	SignedID                      int64                 `json:"signed_id"`
	Name                          string                `json:"name"`
	Addr                          string                `json:"addr"`
	InternalAddr                  string                `json:"internalAddr"`
	Latitude                      float64               `json:"latitude"`
	Longitude                     float64               `json:"longitude"`
	NICSpeedMbps                  int32                 `json:"nicSpeedMbps"`
	IncludedBandwidthGB           int32                 `json:"includedBandwidthGB"`
	MaxBandwidthMbps              int32                 `json:"maxBandwidthMbps"`
	State                         string                `json:"state"`
	ManagementAddr                string                `json:"management_addr"`
	SSHUser                       string                `json:"ssh_user"`
	SSHPort                       int64                 `json:"ssh_port"`
	MaxSessionCount               uint32                `json:"maxSessionCount"`
	PublicKey                     string                `json:"public_key"`
	Version                       string                `json:"relay_version"`
	SellerName                    string                `json:"seller_name"`
	EgressPriceOverride           routing.Nibblin       `json:"egressPriceOverride"`
	MRC                           routing.Nibblin       `json:"monthlyRecurringChargeNibblins"`
	Overage                       routing.Nibblin       `json:"overage"`
	BWRule                        routing.BandWidthRule `json:"bandwidthRule"`
	ContractTerm                  int32                 `json:"contractTerm"`
	StartDate                     time.Time             `json:"startDate"`
	EndDate                       time.Time             `json:"endDate"`
	Type                          routing.MachineType   `json:"machineType"`
	Notes                         string                `json:"notes"`
	DestFirst                     bool                  `json:"dest_first"`
	InternalAddressClientRoutable bool                  `json:"internal_address_client_routable"`
	DatabaseID                    int64
	DatacenterID                  uint64
}

func (s *OpsService) Relays(r *http.Request, args *RelaysArgs, reply *RelaysReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("Relays(): %v", err.Error())
		return &err
	}

	for _, r := range s.Storage.Relays(r.Context()) {
		relay := relay{
			ID:                            r.ID,
			HexID:                         fmt.Sprintf("%016x", r.ID),
			DatacenterHexID:               fmt.Sprintf("%016x", r.Datacenter.ID),
			BillingSupplier:               r.BillingSupplier,
			SignedID:                      r.SignedID,
			Name:                          r.Name,
			Addr:                          r.Addr.String(),
			Latitude:                      float64(r.Datacenter.Location.Latitude),
			Longitude:                     float64(r.Datacenter.Location.Longitude),
			NICSpeedMbps:                  r.NICSpeedMbps,
			IncludedBandwidthGB:           r.IncludedBandwidthGB,
			MaxBandwidthMbps:              r.MaxBandwidthMbps,
			ManagementAddr:                r.ManagementAddr,
			SSHUser:                       r.SSHUser,
			SSHPort:                       r.SSHPort,
			State:                         r.State.String(),
			PublicKey:                     base64.StdEncoding.EncodeToString(r.PublicKey),
			MaxSessionCount:               r.MaxSessions,
			SellerName:                    r.Seller.Name,
			EgressPriceOverride:           r.EgressPriceOverride,
			MRC:                           r.MRC,
			Overage:                       r.Overage,
			BWRule:                        r.BWRule,
			ContractTerm:                  r.ContractTerm,
			StartDate:                     r.StartDate,
			EndDate:                       r.EndDate,
			Type:                          r.Type,
			Notes:                         r.Notes,
			Version:                       r.Version,
			DestFirst:                     r.DestFirst,
			InternalAddressClientRoutable: r.InternalAddressClientRoutable,
			DatabaseID:                    r.DatabaseID,
		}

		if addrStr := r.InternalAddr.String(); addrStr != ":0" {
			relay.InternalAddr = addrStr
		}

		reply.Relays = append(reply.Relays, relay)
	}

	if args.Regex != "" {
		var filtered []relay

		// first check for an exact match
		for idx := range reply.Relays {
			relay := &reply.Relays[idx]
			if relay.Name == args.Regex {
				filtered = append(filtered, *relay)
				break
			}
		}

		// if no relay found, attempt to see if the query matches any seller names
		if len(filtered) == 0 {
			for idx := range reply.Relays {
				relay := &reply.Relays[idx]
				if args.Regex == relay.SellerName {
					filtered = append(filtered, *relay)
				}
			}
		}

		// if still no matches are found, match by regex
		if len(filtered) == 0 {
			for idx := range reply.Relays {
				relay := &reply.Relays[idx]
				if match, err := regexp.Match(args.Regex, []byte(relay.Name)); match && err == nil {
					filtered = append(filtered, *relay)
					continue
				} else if err != nil {
					return err
				}
			}
		}

		reply.Relays = filtered
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].Name < reply.Relays[j].Name
	})

	return nil
}

type RelayEgressPriceOverrideArgs struct {
	SellerShortName string `json:"sellerShortName"`
}

type RelayEgressPriceOverrideReply struct {
	Relays []relay `json:"relays"`
}

func (s *OpsService) RelaysWithEgressPriceOverride(r *http.Request, args *RelayEgressPriceOverrideArgs, reply *RelayEgressPriceOverrideReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RelaysWithEgressPriceOverride(): %v", err.Error())
		return &err
	}

	for _, r := range s.Storage.Relays(r.Context()) {

		if !(r.Seller.ShortName == args.SellerShortName && r.EgressPriceOverride > 0) {
			continue
		}

		relay := relay{
			ID:                            r.ID,
			HexID:                         fmt.Sprintf("%016x", r.ID),
			DatacenterHexID:               fmt.Sprintf("%016x", r.Datacenter.ID),
			BillingSupplier:               r.BillingSupplier,
			SignedID:                      r.SignedID,
			Name:                          r.Name,
			Addr:                          r.Addr.String(),
			Latitude:                      float64(r.Datacenter.Location.Latitude),
			Longitude:                     float64(r.Datacenter.Location.Longitude),
			NICSpeedMbps:                  r.NICSpeedMbps,
			IncludedBandwidthGB:           r.IncludedBandwidthGB,
			MaxBandwidthMbps:              r.MaxBandwidthMbps,
			ManagementAddr:                r.ManagementAddr,
			SSHUser:                       r.SSHUser,
			SSHPort:                       r.SSHPort,
			State:                         r.State.String(),
			PublicKey:                     base64.StdEncoding.EncodeToString(r.PublicKey),
			MaxSessionCount:               r.MaxSessions,
			SellerName:                    r.Seller.Name,
			EgressPriceOverride:           r.EgressPriceOverride,
			MRC:                           r.MRC,
			Overage:                       r.Overage,
			BWRule:                        r.BWRule,
			ContractTerm:                  r.ContractTerm,
			StartDate:                     r.StartDate,
			EndDate:                       r.EndDate,
			Type:                          r.Type,
			Notes:                         r.Notes,
			Version:                       r.Version,
			DestFirst:                     r.DestFirst,
			InternalAddressClientRoutable: r.InternalAddressClientRoutable,
			DatabaseID:                    r.DatabaseID,
		}

		if addrStr := r.InternalAddr.String(); addrStr != ":0" {
			relay.InternalAddr = addrStr
		}

		reply.Relays = append(reply.Relays, relay)
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].Name < reply.Relays[j].Name
	})

	return nil
}

type AddRelayArgs struct {
	Relay routing.Relay `json:"Relay"`
}

type AddRelayReply struct{}

func (s *OpsService) AddRelay(r *http.Request, args *AddRelayArgs, reply *AddRelayReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AddRelay(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	existingRelay, err := s.Storage.Relay(ctx, args.Relay.ID)
	if err == nil && existingRelay.ID == args.Relay.ID {
		// The relay exists and should be reserrected
		err := JSONRPCErrorCodes[int(ERROR_RELAY_NEEDS_RESURRECTION)]
		core.Error("AddRelay(): %v", err.Error())
		return &err
	}

	if err := s.Storage.AddRelay(ctx, args.Relay); err != nil {
		err = fmt.Errorf("AddRelay() error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type JSAddRelayArgs struct {
	Name                          string `json:"name"`
	Addr                          string `json:"addr"`
	InternalAddr                  string `json:"internal_addr"`
	PublicKey                     string `json:"public_key"`
	SellerID                      string `json:"seller"`
	DatacenterID                  string `json:"datacenter"`
	NICSpeedMbps                  int64  `json:"nicSpeedMbps"`
	IncludedBandwidthGB           int64  `json:"includedBandwidthGB"`
	MaxBandwidthMbps              int64  `json:"maxBandwidthMbps"`
	ManagementAddr                string `json:"management_addr"`
	SSHUser                       string `json:"ssh_user"`
	SSHPort                       int64  `json:"ssh_port"`
	MaxSessions                   int64  `json:"max_sessions"`
	EgressPriceOverride           int64  `json:"egressPriceOverride"`
	MRC                           int64  `json:"monthlyRecurringChargeNibblins"`
	Overage                       int64  `json:"overage"`
	BWRule                        int64  `json:"bandwidthRule"`
	ContractTerm                  int64  `json:"contractTerm"`
	StartDate                     string `json:"startDate"`
	EndDate                       string `json:"endDate"`
	Type                          int64  `json:"machineType"`
	Notes                         string `json:"notes"`
	BillingSupplier               string `json:"billingSupplier"`
	Version                       string `json:"relay_version"`
	DestFirst                     bool   `json:"dest_first"`
	InternalAddressClientRoutable bool   `json:"internal_address_client_routable"`
}

type JSAddRelayReply struct{}

func (s *OpsService) JSAddRelay(r *http.Request, args *JSAddRelayArgs, reply *JSAddRelayReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("JSAddRelay(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	rid := crypto.HashID(args.Addr)

	existingRelay, err := s.Storage.Relay(ctx, rid)
	if err == nil && existingRelay.ID == rid {
		// The relay exists and should be reserrected
		err := JSONRPCErrorCodes[int(ERROR_RELAY_NEEDS_RESURRECTION)]
		err.Data = fmt.Sprintf("%016x", existingRelay.ID)
		core.Error("JSAddRelay(): %v", err.Error())
		return &err
	}

	addr, err := net.ResolveUDPAddr("udp", args.Addr)
	if err != nil {
		core.Error("%v", err)
		return err
	}

	dcID, err := strconv.ParseUint(args.DatacenterID, 16, 64)
	if err != nil {
		core.Error("%v", err)
		return err
	}

	var datacenter routing.Datacenter
	if datacenter, err = s.Storage.Datacenter(r.Context(), dcID); err != nil {
		err = fmt.Errorf("Datacenter() error: %w", err)
		core.Error("%v", err)
		return err
	}

	publicKey, err := base64.StdEncoding.DecodeString(args.PublicKey)
	if err != nil {
		err = fmt.Errorf("could not decode base64 public key %s: %v", args.PublicKey, err)
		core.Error("%v", err)
		return err
	}

	relay := routing.Relay{
		ID:                            rid,
		Name:                          args.Name,
		Addr:                          *addr,
		PublicKey:                     publicKey,
		Datacenter:                    datacenter,
		NICSpeedMbps:                  int32(args.NICSpeedMbps),
		IncludedBandwidthGB:           int32(args.IncludedBandwidthGB),
		MaxBandwidthMbps:              int32(args.MaxBandwidthMbps),
		State:                         routing.RelayStateEnabled,
		ManagementAddr:                args.ManagementAddr,
		SSHUser:                       args.SSHUser,
		SSHPort:                       args.SSHPort,
		MaxSessions:                   uint32(args.MaxSessions),
		EgressPriceOverride:           routing.Nibblin(args.EgressPriceOverride),
		MRC:                           routing.Nibblin(args.MRC),
		Overage:                       routing.Nibblin(args.Overage),
		BWRule:                        routing.BandWidthRule(args.BWRule),
		ContractTerm:                  int32(args.ContractTerm),
		Type:                          routing.MachineType(args.Type),
		Notes:                         args.Notes,
		BillingSupplier:               args.BillingSupplier,
		Version:                       args.Version,
		DestFirst:                     args.DestFirst,
		InternalAddressClientRoutable: args.InternalAddressClientRoutable,
	}

	var internalAddr *net.UDPAddr
	if args.InternalAddr != "" {
		internalAddr, err = net.ResolveUDPAddr("udp", args.InternalAddr)
		if err != nil {
			core.Error("%v", err)
			return err
		}
		relay.InternalAddr = *internalAddr
	}

	if args.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", args.StartDate)
		if err != nil {
			core.Error("%v", err)
			return err
		}
		relay.StartDate = startDate
	}

	if args.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", args.EndDate)
		if err != nil {
			core.Error("%v", err)
			return err
		}
		relay.EndDate = endDate
	}

	if err := s.Storage.AddRelay(ctx, relay); err != nil {
		err = fmt.Errorf("AddRelay() error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type RemoveRelayArgs struct {
	RelayID uint64
}

type RemoveRelayReply struct{}

func (s *OpsService) RemoveRelay(r *http.Request, args *RemoveRelayArgs, reply *RemoveRelayReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RemoveRelay(): %v", err.Error())
		return &err
	}

	relay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("RemoveRelay() Storage.Relay error: %w", err)
		core.Error("%v", err)
		return err
	}

	// Rather than actually removing the relay from postgres, just
	// rename it and set it to the decomissioned state
	relay.State = routing.RelayStateDecommissioned

	// want: “$(relayname)-removed-$(date-time-of-removal)”
	shortDate := time.Now().Format("2006-01-02")
	shortTime := time.Now().Format("15:04:05")
	relay.Name = fmt.Sprintf("%s-removed-%s-%s", relay.Name, shortDate, shortTime)

	relay.Addr = net.UDPAddr{} // clear the address to 0 when removed

	if err = s.Storage.SetRelay(r.Context(), relay); err != nil {
		err = fmt.Errorf("RemoveRelay() Storage.SetRelay error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type RelayNameUpdateArgs struct {
	RelayID   uint64 `json:"relay_id"`
	RelayName string `json:"relay_name"`
}

type RelayNameUpdateReply struct {
}

func (s *OpsService) RelayNameUpdate(r *http.Request, args *RelayNameUpdateArgs, reply *RelayNameUpdateReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RelayNameUpdate(): %v", err.Error())
		return &err
	}

	relay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayNameUpdate() Storage.Relay error: %w", err)
		core.Error("%v", err)
		return err
	}

	relay.Name = args.RelayName
	if err = s.Storage.SetRelay(r.Context(), relay); err != nil {
		err = fmt.Errorf("Storage.SetRelay error: %w", err)
		return err
	}

	return nil
}

type RelayStateUpdateArgs struct {
	RelayID    uint64             `json:"relay_id"`
	RelayState routing.RelayState `json:"relay_state"`
}

type RelayStateUpdateReply struct {
}

func (s *OpsService) RelayStateUpdate(r *http.Request, args *RelayStateUpdateArgs, reply *RelayStateUpdateReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RelayStateUpdate(): %v", err.Error())
		return &err
	}

	relay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayStateUpdate() Storage.Relay error: %w", err)
		core.Error("%v", err)
		return err
	}

	relay.State = args.RelayState
	if err = s.Storage.SetRelay(r.Context(), relay); err != nil {
		err = fmt.Errorf("RelayStateUpdate() Storage.SetRelay error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type RelayPublicKeyUpdateArgs struct {
	RelayID        uint64 `json:"relay_id"`
	RelayPublicKey string `json:"relay_public_key"`
}

type RelayPublicKeyUpdateReply struct {
}

func (s *OpsService) RelayPublicKeyUpdate(r *http.Request, args *RelayPublicKeyUpdateArgs, reply *RelayPublicKeyUpdateReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RelayPublicKeyUpdate(): %v", err.Error())
		return &err
	}

	relay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayPublicKeyUpdate()")
		return err
	}

	relay.PublicKey, err = base64.StdEncoding.DecodeString(args.RelayPublicKey)

	if err != nil {
		err = fmt.Errorf("RelayPublicKeyUpdate() could not decode relay public key: %v", err)
		core.Error("%v", err)
		return err
	}

	if err = s.Storage.SetRelay(r.Context(), relay); err != nil {
		err = fmt.Errorf("RelayPublicKeyUpdate() SetRelay error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type DatacenterArg struct {
	ID uint64
}

type DatacenterReply struct {
	Datacenter routing.Datacenter
}

func (s *OpsService) Datacenter(r *http.Request, arg *DatacenterArg, reply *DatacenterReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("Datacenter(): %v", err.Error())
		return &err
	}

	var datacenter routing.Datacenter
	var err error
	if datacenter, err = s.Storage.Datacenter(r.Context(), arg.ID); err != nil {
		err = fmt.Errorf("Datacenter() error: %w", err)
		core.Error("%v", err)
		return err
	}

	reply.Datacenter = datacenter
	return nil

}

/*
 	TODO: These functions will eventually be renamed but is being used instead of the existing functions to avoid breaking datacenters page in admin tool
	 and to avoid continuing the pattern of two separate endpoints (next and admin tool)
*/

type mappableDatacenter struct {
	Alias string `json:"alias"`
	Name  string `json:"name"`
	HexID string `json:"hex_id"`
}

type DatacenterListArgs struct{}

type DatacenterListReply struct {
	Datacenters []mappableDatacenter `json:"datacenters"`
}

func (s *OpsService) DatacenterList(r *http.Request, args *DatacenterListArgs, reply *DatacenterListReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("DatacenterList(): %v", err.Error())
		return &err
	}

	allDatacenters := s.Storage.Datacenters(r.Context())

	reply.Datacenters = make([]mappableDatacenter, len(allDatacenters))

	for i, datacenter := range allDatacenters {
		reply.Datacenters[i] = mappableDatacenter{
			Alias: datacenter.AliasName,
			Name:  datacenter.Name,
			HexID: fmt.Sprintf("%016x", datacenter.ID),
		}
	}

	return nil
}

type BuyerDatacenterMapArgs struct {
	BuyerHexID      string `json:"buyer_hex_id"`
	DatacenterHexID string `json:"datacenter_hex_id"`
}

type BuyerDatacenterMapReply struct{}

func (s *OpsService) AddBuyerDatacenterMap(r *http.Request, args *BuyerDatacenterMapArgs, reply *BuyerDatacenterMapReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RemoveBuyerDatacenterMap(): %v", err.Error())
		return &err
	}

	ctx := r.Context()

	buyerID, err := strconv.ParseUint(args.BuyerHexID, 16, 64)
	if err != nil {
		return fmt.Errorf("Unable to parse Buyer ID: %s", args.BuyerHexID)
	}

	datacenterID, err := strconv.ParseUint(args.DatacenterHexID, 16, 64)
	if err != nil {
		return fmt.Errorf("Unable to parse Datacenter ID: %s", args.BuyerHexID)
	}

	dcMap := routing.DatacenterMap{
		BuyerID:      buyerID,
		DatacenterID: datacenterID,
	}

	return s.Storage.AddDatacenterMap(ctx, dcMap)
}

func (s *OpsService) RemoveBuyerDatacenterMap(r *http.Request, args *BuyerDatacenterMapArgs, reply *BuyerDatacenterMapReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RemoveBuyerDatacenterMap(): %v", err.Error())
		return &err
	}

	ctx := r.Context()

	buyerID, err := strconv.ParseUint(args.BuyerHexID, 16, 64)
	if err != nil {
		return fmt.Errorf("Unable to parse Buyer ID: %s", args.BuyerHexID)
	}

	datacenterID, err := strconv.ParseUint(args.DatacenterHexID, 16, 64)
	if err != nil {
		return fmt.Errorf("Unable to parse Datacenter ID: %s", args.BuyerHexID)
	}

	dcMap := routing.DatacenterMap{
		BuyerID:      buyerID,
		DatacenterID: datacenterID,
	}

	return s.Storage.RemoveDatacenterMap(ctx, dcMap)
}

// --------------------------------------------------------------------------------------------------------------------------------------------

type DatacentersArgs struct {
	Name string `json:"name"`
}

type DatacentersReply struct {
	Datacenters []datacenter
}

type datacenter struct {
	Name         string  `json:"name"`
	HexID        string  `json:"hexID"`
	ID           uint64  `json:"id"`
	SignedID     int64   `json:"signed_id"`
	Latitude     float32 `json:"latitude"`
	Longitude    float32 `json:"longitude"`
	SupplierName string  `json:"supplierName"`
}

func (s *OpsService) Datacenters(r *http.Request, args *DatacentersArgs, reply *DatacentersReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("Datacenters(): %v", err.Error())
		return &err
	}

	for _, d := range s.Storage.Datacenters(r.Context()) {
		reply.Datacenters = append(reply.Datacenters, datacenter{
			Name:      d.Name,
			HexID:     fmt.Sprintf("%016x", d.ID),
			ID:        d.ID,
			Latitude:  d.Location.Latitude,
			Longitude: d.Location.Longitude,
		})
	}

	if args.Name != "" {
		var filtered []datacenter
		for idx := range reply.Datacenters {
			if strings.Contains(reply.Datacenters[idx].Name, args.Name) {
				filtered = append(filtered, reply.Datacenters[idx])
			}
		}
		reply.Datacenters = filtered
	}

	sort.Slice(reply.Datacenters, func(i int, j int) bool {
		return reply.Datacenters[i].Name < reply.Datacenters[j].Name
	})

	return nil
}

type AddDatacenterArgs struct {
	Datacenter routing.Datacenter
}

type AddDatacenterReply struct{}

func (s *OpsService) AddDatacenter(r *http.Request, args *AddDatacenterArgs, reply *AddDatacenterReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AddDatacenter(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddDatacenter(ctx, args.Datacenter); err != nil {
		err = fmt.Errorf("AddDatacenter() error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type JSAddDatacenterArgs struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	SellerID  string  `json:"sellerID"`
}

type JSAddDatacenterReply struct{}

func (s *OpsService) JSAddDatacenter(r *http.Request, args *JSAddDatacenterArgs, reply *JSAddDatacenterReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("JSAddDatacenter(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	dcID := crypto.HashID(args.Name)

	seller, err := s.Storage.Seller(r.Context(), args.SellerID)
	if err != nil {
		core.Error("%v", err)
		return err
	}

	datacenter := routing.Datacenter{
		Name: args.Name,
		ID:   dcID,
		Location: routing.Location{
			Latitude:  float32(args.Latitude),
			Longitude: float32(args.Longitude),
		},
		SellerID: seller.DatabaseID,
	}

	if err := s.Storage.AddDatacenter(ctx, datacenter); err != nil {
		err = fmt.Errorf("AddDatacenter() error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type RemoveDatacenterArgs struct {
	Name string
}

type RemoveDatacenterReply struct{}

func (s *OpsService) RemoveDatacenter(r *http.Request, args *RemoveDatacenterArgs, reply *RemoveDatacenterReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RemoveDatacenter(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	id := crypto.HashID(args.Name)

	if err := s.Storage.RemoveDatacenter(ctx, id); err != nil {
		err = fmt.Errorf("RemoveDatacenter() error: %w", err)
		core.Error("%v", err)
		return err
	}

	return nil
}

type ListDatacenterMapsArgs struct {
	DatacenterID uint64
}

type ListDatacenterMapsReply struct {
	DatacenterMaps []DatacenterMapsFull
}

// A zero DatacenterID returns a list of all maps.
func (s *OpsService) ListDatacenterMaps(r *http.Request, args *ListDatacenterMapsArgs, reply *ListDatacenterMapsReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("ListDatacenterMaps(): %v", err.Error())
		return &err
	}

	dcm := s.Storage.ListDatacenterMaps(r.Context(), args.DatacenterID)

	var replySlice []DatacenterMapsFull
	for _, dcMap := range dcm {
		buyer, err := s.Storage.Buyer(r.Context(), dcMap.BuyerID)
		if err != nil {
			err = fmt.Errorf("ListDatacenterMaps() could not parse buyer: %w", err)
			core.Error("%v", err)
			return err
		}
		datacenter, err := s.Storage.Datacenter(r.Context(), dcMap.DatacenterID)
		if err != nil {
			err = fmt.Errorf("ListDatacenterMaps() could not parse datacenter: %w", err)
			core.Error("%v", err)
			return err
		}

		company, err := s.Storage.Customer(r.Context(), buyer.CompanyCode)
		if err != nil {
			err = fmt.Errorf("ListDatacenterMaps() failed to find buyer company: %w", err)
			core.Error("%v", err)
			return err
		}

		dcmFull := DatacenterMapsFull{
			DatacenterName: datacenter.Name,
			DatacenterID:   fmt.Sprintf("%016x", dcMap.DatacenterID),
			BuyerName:      company.Name,
			BuyerID:        fmt.Sprintf("%016x", dcMap.BuyerID),
		}

		replySlice = append(replySlice, dcmFull)
	}

	reply.DatacenterMaps = replySlice

	return nil
}

type RouteSelectionArgs struct {
	SourceRelays      []string `json:"src_relays"`
	DestinationRelays []string `json:"dest_relays"`
	RTT               float64  `json:"rtt"`
	RouteHash         uint64   `json:"route_hash"`
}

type RouteSelectionReply struct {
	Routes []routing.Route `json:"routes"`
}

type CheckRelayIPAddressArgs struct {
	IpAddress string `json:"ipAddress"`
	HexID     string `json:"hexID"`
}

type CheckRelayIPAddressReply struct {
	Valid bool `json:"valid"`
}

// CheckRelayIPAddress is used by the Admin tool when recommissioning a relay to ensure the
// selected IP address "matches" the HexID (which was derived from its original IP address).
func (s *OpsService) CheckRelayIPAddress(r *http.Request, args *CheckRelayIPAddressArgs, reply *CheckRelayIPAddressReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CheckRelayIPAddress(): %v", err.Error())
		return &err
	}

	internalIDFromHexID, err := strconv.ParseUint(args.HexID, 16, 64)
	if err != nil {
		reply.Valid = false
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", args.IpAddress)
	if err != nil {
		reply.Valid = false
		return err
	}

	internalIdFromIpAddress := crypto.HashID(addr.String())
	if internalIDFromHexID != internalIdFromIpAddress {
		reply.Valid = false
		return fmt.Errorf("CheckRelayIPAddress(): internal ID from Hex ID (%016x) does not match internal ID from IP Address (%016x)", internalIDFromHexID, internalIdFromIpAddress)
	}

	reply.Valid = true
	return nil
}

type UpdateRelayArgs struct {
	RelayID    uint64      `json:"relayID"`    // used by next tool
	HexRelayID string      `json:"hexRelayID"` // used by javascript clients
	Field      string      `json:"field"`
	Value      interface{} `json:"value"`
}

type UpdateRelayReply struct{}

func (s *OpsService) UpdateRelay(r *http.Request, args *UpdateRelayArgs, reply *UpdateRelayReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateRelay(): %v", err.Error())
		return &err
	}

	relayID := args.RelayID
	var err error
	if args.HexRelayID != "" {
		relayID, err = strconv.ParseUint(args.HexRelayID, 16, 64)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() failed to parse HexRelayID %s: %w", args.HexRelayID, err)
			core.Error("%v", err)
			return err
		}
	}
	err = s.Storage.UpdateRelay(r.Context(), relayID, args.Field, args.Value)
	if err != nil {
		err = fmt.Errorf("UpdateRelay() failed to modify relay record for field %s with value %v: %w", args.Field, args.Value, err)
		core.Error("%v", err)
		return err
	}
	return nil
}

type GetRelayArgs struct {
	RelayID uint64
}

type GetRelayReply struct {
	Relay relay
}

func (s *OpsService) GetRelay(r *http.Request, args *GetRelayArgs, reply *GetRelayReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("GetRelay(): %v", err.Error())
		return &err
	}

	routingRelay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("Error retrieving relay ID %016x: %v", args.RelayID, err)
		core.Error("%v", err)
		return err
	}

	relay := relay{
		ID:                            routingRelay.ID,
		SignedID:                      routingRelay.SignedID,
		Name:                          routingRelay.Name,
		Addr:                          routingRelay.Addr.String(),
		InternalAddr:                  routingRelay.InternalAddr.String(),
		Latitude:                      float64(routingRelay.Datacenter.Location.Latitude),
		Longitude:                     float64(routingRelay.Datacenter.Location.Longitude),
		NICSpeedMbps:                  routingRelay.NICSpeedMbps,
		IncludedBandwidthGB:           routingRelay.IncludedBandwidthGB,
		MaxBandwidthMbps:              routingRelay.MaxBandwidthMbps,
		ManagementAddr:                routingRelay.ManagementAddr,
		SSHUser:                       routingRelay.SSHUser,
		SSHPort:                       routingRelay.SSHPort,
		State:                         routingRelay.State.String(),
		PublicKey:                     base64.StdEncoding.EncodeToString(routingRelay.PublicKey),
		MaxSessionCount:               routingRelay.MaxSessions,
		SellerName:                    routingRelay.Seller.Name,
		EgressPriceOverride:           routingRelay.EgressPriceOverride,
		MRC:                           routingRelay.MRC,
		Overage:                       routingRelay.Overage,
		BWRule:                        routingRelay.BWRule,
		ContractTerm:                  routingRelay.ContractTerm,
		StartDate:                     routingRelay.StartDate,
		EndDate:                       routingRelay.EndDate,
		Type:                          routingRelay.Type,
		Notes:                         routingRelay.Notes,
		DatabaseID:                    routingRelay.DatabaseID,
		DatacenterID:                  routingRelay.Datacenter.ID,
		BillingSupplier:               routingRelay.BillingSupplier,
		Version:                       routingRelay.Version,
		DestFirst:                     routingRelay.DestFirst,
		InternalAddressClientRoutable: routingRelay.InternalAddressClientRoutable,
	}

	reply.Relay = relay

	return nil
}

type ModifyRelayFieldArgs struct {
	RelayID uint64
	Field   string
	Value   string
}

type ModifyRelayFieldReply struct{}

func (s *OpsService) ModifyRelayField(r *http.Request, args *ModifyRelayFieldArgs, reply *ModifyRelayFieldReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("ModifyRelayField(): %v", err.Error())
		return &err
	}

	// sort out the value type here (comes from the next tool and javascript UI as a string)
	switch args.Field {
	// sent to storer as float64
	case "NICSpeedMbps", "IncludedBandwidthGB", "MaxBandwidthMbps", "ContractTerm", "SSHPort", "MaxSessions":
		newfloat, err := strconv.ParseFloat(args.Value, 64)
		if err != nil {
			return fmt.Errorf("Value: %v is not a valid numeric type", args.Value)
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, newfloat)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			core.Error("%v", err)
			return err
		}

	// net.UDPAddr, time.Time - all sent to storer as strings
	case "Addr", "InternalAddr", "ManagementAddr", "SSHUser", "StartDate", "EndDate", "BillingSupplier", "Version":
		err := s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			core.Error("%v", err)
			return err
		}

	// sent to storer as bool
	case "DestFirst", "InternalAddressClientRoutable":
		newBool, err := strconv.ParseBool(args.Value)
		if err != nil {
			return fmt.Errorf("Value: %v is not a valid boolean type", args.Value)
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, newBool)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			core.Error("%v", err)
			return err
		}

	// relay.PublicKey
	case "PublicKey":
		newPublicKey := string(args.Value)
		err := s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, newPublicKey)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			core.Error("%v", err)
			return err
		}

	// routing.RelayState
	case "State":

		state, err := routing.ParseRelayState(args.Value)
		if err != nil {
			err := fmt.Errorf("value '%s' is not a valid relay state", args.Value)
			core.Error("%v", err)
			return err
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, float64(state))
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			core.Error("%v", err)
			return err
		}

	// nibblins (received as USD, sent to storer as float64)
	case "EgressPriceOverride", "MRC", "Overage":
		newValue, err := strconv.ParseFloat(args.Value, 64)
		if err != nil {
			err = fmt.Errorf("value '%s' is not a valid float64 number: %v", args.Value, err)
			core.Error("%v", err)
			return err
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, newValue)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			core.Error("%v", err)
			return err
		}

	// routing.BandwidthRule
	case "BWRule":

		bwRule, err := routing.ParseBandwidthRule(args.Value)
		if err != nil {
			err := fmt.Errorf("value '%s' is not a valid bandwidth rule", args.Value)
			core.Error("%v", err)
			return err
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, float64(bwRule))
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			core.Error("%v", err)
			return err
		}

		// routing.MachineType
	case "Type":

		machineType, err := routing.ParseMachineType(args.Value)
		if err != nil {
			err := fmt.Errorf("value '%s' is not a valid machine type", args.Value)
			core.Error("%v", err)
			return err
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, float64(machineType))
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			core.Error("%v", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist on the Relay type", args.Field)
	}

	return nil
}

type UpdateCustomerArgs struct {
	CustomerID string `json:"customerCode"`
	Field      string `json:"field"`
	Value      string `json:"value"`
}

type UpdateCustomerReply struct{}

func (s *OpsService) UpdateCustomer(r *http.Request, args *UpdateCustomerArgs, reply *UpdateCustomerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateCustomer(): %v", err.Error())
		return &err
	}

	// sort out the value type here (comes from the next tool and javascript UI as a string)
	switch args.Field {
	case "Name", "AutomaticSigninDomains":
		err := s.Storage.UpdateCustomer(r.Context(), args.CustomerID, args.Field, args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateCustomer() error updating record for customer %s: %v", args.CustomerID, err)
			core.Error("%v", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the Customer type", args.Field)
	}

	return nil
}

type RemoveCustomerArgs struct {
	CustomerCode string `json:"customerCode"`
}

type RemoveCustomerReply struct{}

func (s *OpsService) RemoveCustomer(r *http.Request, args *RemoveCustomerArgs, reply *RemoveCustomerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("RemoveCustomer(): %v", err.Error())
		return &err
	}

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	return s.Storage.RemoveCustomer(ctx, args.CustomerCode)
}

type UpdateSellerArgs struct {
	SellerID string `json:"shortName"`
	Field    string `json:"field"`
	Value    string `json:"value"`
}

type UpdateSellerReply struct{}

func (s *OpsService) UpdateSeller(r *http.Request, args *UpdateSellerArgs, reply *UpdateSellerReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateSeller(): %v", err.Error())
		return &err
	}

	// sort out the value type here (comes from the next tool and javascript UI as a string)
	switch args.Field {
	case "ShortName":
		err := s.Storage.UpdateSeller(r.Context(), args.SellerID, args.Field, args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateSeller() error updating record for seller %s: %v", args.SellerID, err)
			core.Error("%v", err)
			return err
		}
	case "Secret":
		secret, err := strconv.ParseBool(args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateSeller() value '%s' is not a valid Secret/boolean: %v", args.Value, err)
			core.Error("%v", err)
			return err
		}
		err = s.Storage.UpdateSeller(r.Context(), args.SellerID, args.Field, secret)
		if err != nil {
			err = fmt.Errorf("UpdateSeller() error updating record for seller %s: %v", args.SellerID, err)
			core.Error("%v", err)
			return err
		}
	case "EgressPrice", "EgressPriceNibblinsPerGB":
		newValue, err := strconv.ParseFloat(args.Value, 64)
		if err != nil {
			err = fmt.Errorf("UpdateSeller() value '%s' is not a valid price: %v", args.Value, err)
			core.Error("%v", err)
			return err
		}

		if args.Field == "EgressPrice" {
			args.Field = "EgressPriceNibblinsPerGB"
		}

		err = s.Storage.UpdateSeller(r.Context(), args.SellerID, args.Field, newValue)
		if err != nil {
			err = fmt.Errorf("UpdateSeller() error updating field for seller %s: %v", args.SellerID, err)
			core.Error("%v", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the Seller type", args.Field)
	}

	return nil
}

type ResetSellerEgressPriceOverrideArgs struct {
	SellerID string `json:"shortName"`
	Field    string `json:"field"`
}

type ResetSellerEgressPriceOverrideReply struct{}

func (s *OpsService) ResetSellerEgressPriceOverride(r *http.Request, args *ResetSellerEgressPriceOverrideArgs, reply *ResetSellerEgressPriceOverrideReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("ResetSellerEgressPriceOverride(): %v", err.Error())
		return &err
	}

	// Iterate through relays and reset egress price override for this seller's relays
	relays := s.Storage.Relays(r.Context())

	for _, relay := range relays {

		switch args.Field {
		case "EgressPriceOverride":
			if relay.Seller.ShortName == args.SellerID {
				err := s.Storage.UpdateRelay(r.Context(), relay.ID, args.Field, float64(0))
				if err != nil {
					err = fmt.Errorf("ResetSellerEgressPriceOverride() error updating %s for seller %s: %v", args.Field, args.SellerID, err)
					core.Error("%v", err)
					return err
				}
			}
		default:
			return fmt.Errorf("Field '%s' is not a valid Relay type for resetting seller egress price override", args.Field)
		}
	}

	return nil
}

type UpdateDatacenterArgs struct {
	HexDatacenterID string      `json:"hexDatacenterID"`
	Field           string      `json:"field"`
	Value           interface{} `json:"value"`
}

type UpdateDatacenterReply struct{}

func (s *OpsService) UpdateDatacenter(r *http.Request, args *UpdateDatacenterArgs, reply *UpdateDatacenterReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateDatacenter(): %v", err.Error())
		return &err
	}

	dcID, err := strconv.ParseUint(args.HexDatacenterID, 16, 64)
	if err != nil {
		err = fmt.Errorf("UpdateDatacenter() failed to parse hex datacenter ID: %v", err)
		core.Error("%v", err)
		return err
	}

	switch args.Field {
	case "Latitude", "Longitude":
		floatValue, ok := args.Value.(float64)
		if !ok {
			err = fmt.Errorf("UpdateDatacenter() value '%v' is not a valid float32 type", args.Value)
			core.Error("%v", err)
			return err
		}

		newValue := float32(floatValue)
		err := s.Storage.UpdateDatacenter(r.Context(), dcID, args.Field, newValue)
		if err != nil {
			err = fmt.Errorf("UpdateDatacenter() error updating record for datacenter %s: %v", args.HexDatacenterID, err)
			core.Error("%v", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the Datacenter type", args.Field)
	}

	return nil
}

type FetchAnalyticsDashboardCategoriesArgs struct{}

type FetchAnalyticsDashboardCategoriesReply struct {
	Categories    []looker.AnalyticsDashboardCategory            `json:"categories"`
	SubCategories map[string][]looker.AnalyticsDashboardCategory `json:"sub_categories"`
}

func (s *OpsService) FetchAnalyticsDashboardCategories(r *http.Request, args *FetchAnalyticsDashboardCategoriesArgs, reply *FetchAnalyticsDashboardCategoriesReply) error {
	reply.Categories = make([]looker.AnalyticsDashboardCategory, 0)
	reply.SubCategories = make(map[string][]looker.AnalyticsDashboardCategory)

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("FetchAnalyticsDashboardCategories(): %v", err.Error())
		return &err
	}

	ctx := r.Context()

	categories, err := s.Storage.GetAnalyticsDashboardCategories(ctx)
	if err != nil {
		core.Error("FetchAnalyticsDashboardCategories(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	// Filter out all parent categories and build a map to their children
	for _, category := range categories {
		if category.ParentCategoryID == -1 {
			reply.Categories = append(reply.Categories, category)

			_, ok := reply.SubCategories[category.Label]
			if !ok {
				reply.SubCategories[category.Label] = make([]looker.AnalyticsDashboardCategory, 0)
			}

			subCategories, err := s.Storage.GetAnalyticsDashboardSubCategoriesByCategoryID(ctx, category.ID)
			if err != nil {
				core.Error("FetchAnalyticsDashboardCategory(): %v", err.Error())
			}

			reply.SubCategories[category.Label] = subCategories
		}
	}

	return nil
}

type FetchAnalyticsDashboardCategoryArgs struct {
	ID int32 `json:"id"`
}

type FetchAnalyticsDashboardCategoryReply struct {
	Category      looker.AnalyticsDashboardCategory   `json:"category"`
	SubCategories []looker.AnalyticsDashboardCategory `json:"sub_categories"`
}

func (s *OpsService) FetchAnalyticsDashboardCategory(r *http.Request, args *FetchAnalyticsDashboardCategoryArgs, reply *FetchAnalyticsDashboardCategoryReply) error {
	reply.SubCategories = make([]looker.AnalyticsDashboardCategory, 0)

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("FetchAnalyticsDashboardCategory(): %v", err.Error())
		return &err
	}

	ctx := r.Context()

	category, err := s.Storage.GetAnalyticsDashboardCategoryByID(ctx, int64(args.ID))
	if err != nil {
		core.Error("FetchAnalyticsDashboardCategory(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	reply.Category = category

	// If the category is a parent category, look for sub categories (-1 determines a parent category)
	if category.ParentCategoryID < 0 {
		subCategories, err := s.Storage.GetAnalyticsDashboardSubCategoriesByCategoryID(ctx, category.ID)
		if err != nil {
			core.Error("FetchAnalyticsDashboardCategory(): %v", err.Error())
		}

		reply.SubCategories = subCategories
		return nil
	}

	return nil
}

type AddAnalyticsDashboardCategoryArgs struct {
	Order            int32  `json:"order"`
	Label            string `json:"label"`
	ParentCategoryID int32  `json:"parent_category_id"`
}

type AddAnalyticsDashboardCategoryReply struct{}

func (s *OpsService) AddAnalyticsDashboardCategory(r *http.Request, args *AddAnalyticsDashboardCategoryArgs, reply *AddAnalyticsDashboardCategoryReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AddAnalyticsDashboardCategory(): %v", err.Error())
		return &err
	}

	if args.Label == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Label"
		core.Error("AddAnalyticsDashboardCategory(): %v: Label is required", err.Error())
		return &err
	}

	if err := s.Storage.AddAnalyticsDashboardCategory(r.Context(), args.Order, args.Label, int64(args.ParentCategoryID)); err != nil {
		core.Error("AddAnalyticsDashboardCategory(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type DeleteAnalyticsDashboardCategoryArgs struct {
	ID int32 `json:"id"`
}

type DeleteAnalyticsDashboardCategoryReply struct{}

func (s *OpsService) DeleteAnalyticsDashboardCategory(r *http.Request, args *DeleteAnalyticsDashboardCategoryArgs, reply *DeleteAnalyticsDashboardCategoryReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("DeleteAnalyticsDashboardCategory(): %v", err.Error())
		return &err
	}

	ctx := r.Context()

	// Remove dashboards with this category to take care of FK issues
	dashboards, err := s.Storage.GetAnalyticsDashboardsByCategoryID(ctx, int64(args.ID))
	if err != nil {
		core.Error("DeleteAnalyticsDashboardCategory(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	for _, d := range dashboards {
		if err := s.Storage.RemoveAnalyticsDashboardByID(ctx, d.ID); err != nil {
			core.Error("DeleteAnalyticsDashboardCategory(): %v", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}
	}

	subCategories, err := s.Storage.GetAnalyticsDashboardSubCategoriesByCategoryID(ctx, int64(args.ID))
	if err != nil {
		core.Error("DeleteAnalyticsDashboardCategory(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	for _, subCategory := range subCategories {
		subDashboards, err := s.Storage.GetAnalyticsDashboardsByCategoryID(ctx, subCategory.ID)
		if err != nil {
			core.Error("DeleteAnalyticsDashboardCategory(): %v", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}

		for _, d := range subDashboards {
			if err := s.Storage.RemoveAnalyticsDashboardByID(ctx, d.ID); err != nil {
				core.Error("DeleteAnalyticsDashboardCategory(): %v", err.Error())
				err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
				return &err
			}
		}

		if err := s.Storage.RemoveAnalyticsDashboardCategoryByID(ctx, subCategory.ID); err != nil {
			core.Error("DeleteAnalyticsDashboardCategory(): %v", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}
	}

	if err := s.Storage.RemoveAnalyticsDashboardCategoryByID(ctx, int64(args.ID)); err != nil {
		core.Error("DeleteAnalyticsDashboardCategory(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type UpdateAnalyticsDashboardCategoryArgs struct {
	ID      int32  `json:"id"`
	Order   int32  `json:"order"`
	Label   string `json:"label"`
	Premium bool   `json:"premium"`
	Admin   bool   `json:"admin"`
	Seller  bool   `json:"seller"`
}

type UpdateAnalyticsDashboardCategoryReply struct{}

func (s *OpsService) UpdateAnalyticsDashboardCategory(r *http.Request, args *UpdateAnalyticsDashboardCategoryArgs, reply *UpdateAnalyticsDashboardCategoryReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateAnalyticsDashboardCategory(): %v", err.Error())
		return &err
	}

	if args.Label == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Label"
		core.Error("UpdateAnalyticsDashboardCategory(): %v: Label is required", err.Error())
		return &err
	}

	ctx := r.Context()

	category, err := s.Storage.GetAnalyticsDashboardCategoryByID(ctx, int64(args.ID))
	if err != nil {
		core.Error("UpdateAnalyticsDashboardCategory(): %v: Name is required", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err

	}

	wasError := false
	if category.Label != args.Label {
		if err := s.Storage.UpdateAnalyticsDashboardCategoryByID(ctx, int64(args.ID), "Label", args.Label); err != nil {
			core.Error("UpdateAnalyticsDashboardCategory(): %v", err.Error())
			wasError = true
		}
	}

	if category.Order != args.Order {
		if err := s.Storage.UpdateAnalyticsDashboardCategoryByID(ctx, int64(args.ID), "Order", args.Order); err != nil {
			core.Error("UpdateAnalyticsDashboardCategory(): %v", err.Error())
			wasError = true
		}
	}

	if wasError {
		core.Error("UpdateAnalyticsDashboardCategory(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type AnalyticsDashboard struct {
	ID          int32  `json:"id"`
	Order       int32  `json:"order"`
	Category    string `json:"category"`
	SubCategory string `json:"sub_category"`
	Customer    string `json:"customer"`
	LookerID    int32  `json:"looker_id"`
	Name        string `json:"name"`
	AdminOnly   bool   `json:"admin_only"`
	Premium     bool   `json:"premium"`
}

type FetchAnalyticsDashboardListArgs struct {
	CustomerCode string `json:"customer_code"`
}

type FetchAnalyticsDashboardListReply struct {
	Dashboards []AnalyticsDashboard `json:"dashboards"`
}

func (s *OpsService) FetchAnalyticsDashboardList(r *http.Request, args *FetchAnalyticsDashboardListArgs, reply *FetchAnalyticsDashboardListReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("FetchAnalyticsDashboardList(): %v", err.Error())
		return &err
	}

	ctx := r.Context()

	reply.Dashboards = make([]AnalyticsDashboard, 0)

	dashboards, err := s.Storage.GetAnalyticsDashboards(ctx)
	if err != nil {
		core.Error("FetchAnalyticsDashboardList(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	for _, dashboard := range dashboards {
		// TODO: This can be replaced with a better storage call
		if args.CustomerCode == "" || dashboard.CustomerCode == args.CustomerCode {
			parentCategory := looker.AnalyticsDashboardCategory{}
			childCategory := looker.AnalyticsDashboardCategory{}

			// If the dashboard has a valid parent category ID, find with ID and use that as the parent otherwise use the dashboard's category
			if dashboard.Category.ParentCategoryID != -1 {
				parentCategory, err = s.Storage.GetAnalyticsDashboardCategoryByID(ctx, dashboard.Category.ParentCategoryID)
				if err != nil {
					core.Error("FetchAnalyticsDashboardList(): %v", err.Error())
					err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
					return &err
				}
				childCategory = dashboard.Category
			} else {
				parentCategory = dashboard.Category
			}

			reply.Dashboards = append(reply.Dashboards, AnalyticsDashboard{
				ID:          int32(dashboard.ID),
				Order:       dashboard.Order,
				LookerID:    int32(dashboard.LookerID),
				Category:    parentCategory.Label,
				SubCategory: childCategory.Label,
				Customer:    dashboard.CustomerCode,
				Name:        dashboard.Name,
				AdminOnly:   dashboard.Admin,
				Premium:     dashboard.Premium,
			})
		}
	}

	return nil
}

type FetchAnalyticsDashboardInformationArgs struct {
	ID int32 `json:"id"`
}

type FetchAnalyticsDashboardInformationReply struct {
	ID           int32                             `json:"id"`
	Order        int32                             `json:"order"`
	Name         string                            `json:"name"`
	AdminOnly    bool                              `json:"admin_only"`
	Premium      bool                              `json:"premium"`
	CustomerCode string                            `json:"customer_code"`
	Category     looker.AnalyticsDashboardCategory `json:"category"`
	LookerID     int32                             `json:"looker_id"`
}

func (s *OpsService) FetchAnalyticsDashboardInformation(r *http.Request, args *FetchAnalyticsDashboardInformationArgs, reply *FetchAnalyticsDashboardInformationReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("FetchAnalyticsDashboardInformation(): %v", err.Error())
		return &err
	}

	dashboard, err := s.Storage.GetAnalyticsDashboardByID(r.Context(), int64(args.ID))
	if err != nil {
		core.Error("FetchAnalyticsDashboardInformation(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	reply.ID = int32(dashboard.ID)
	reply.Order = dashboard.Order
	reply.Name = dashboard.Name
	reply.AdminOnly = dashboard.Admin
	reply.Premium = dashboard.Premium
	reply.CustomerCode = dashboard.CustomerCode
	reply.Category = dashboard.Category
	reply.LookerID = int32(dashboard.LookerID)

	return nil
}

type AddAnalyticsDashboardArgs struct {
	Order        int32  `json:"order"`
	Name         string `json:"name"`
	AdminOnly    bool   `json:"admin_only"`
	Premium      bool   `json:"premium"`
	LookerID     int32  `json:"looker_id"`
	Discovery    bool   `json:"discovery"`
	CustomerCode string `json:"customer_code"`
	CategoryID   int32  `json:"category_id"`
}

type AddAnalyticsDashboardReply struct{}

func (s *OpsService) AddAnalyticsDashboard(r *http.Request, args *AddAnalyticsDashboardArgs, reply *AddAnalyticsDashboardReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AddAnalyticsDashboard(): %v", err.Error())
		return &err
	}

	if args.Name == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Name"
		core.Error("AddAnalyticsDashboard(): %v: Name is required", err.Error())
		return &err
	}

	if args.CustomerCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "CustomerCode"
		core.Error("AddAnalyticsDashboard(): %v: Name is required", err.Error())
		return &err
	}

	ctx := r.Context()

	customer, err := s.Storage.Customer(ctx, args.CustomerCode)
	if err != nil {
		core.Error("AddAnalyticsDashboard(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	if err := s.Storage.AddAnalyticsDashboard(ctx, args.Order, args.Name, args.AdminOnly, args.Premium, int64(args.LookerID), customer.DatabaseID, int64(args.CategoryID)); err != nil {
		core.Error("AddAnalyticsDashboard(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type DeleteAnalyticsDashboardArgs struct {
	ID int32 `json:"id"`
}

type DeleteAnalyticsDashboardReply struct{}

func (s *OpsService) DeleteAnalyticsDashboard(r *http.Request, args *DeleteAnalyticsDashboardArgs, reply *DeleteAnalyticsDashboardReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("DeleteAnalyticsDashboard(): %v", err.Error())
		return &err
	}

	if err := s.Storage.RemoveAnalyticsDashboardByID(r.Context(), int64(args.ID)); err != nil {
		core.Error("DeleteAnalyticsDashboard(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type UpdateAnalyticsDashboardArgs struct {
	ID         int32  `json:"id"`
	Order      int32  `json:"order"`
	Name       string `json:"name"`
	LookerID   int32  `json:"looker_id"`
	CategoryID int32  `json:"category_id"`
	AdminOnly  bool   `json:"admin_only"`
	Premium    bool   `json:"premium"`
}

type UpdateAnalyticsDashboardReply struct{}

// This function accomplishes "bulk" update for a list of customer codes and their corresponding dashboards. The three main use cases are:
// 1. New Dashboard (new customer code that wasn't there originally)
// 2. Dashboard removal (customer code was removed from the original list)
// 3. Dashboard update (customer code and dashboard ID are still available and match up with DB values)
func (s *OpsService) UpdateAnalyticsDashboard(r *http.Request, args *UpdateAnalyticsDashboardArgs, reply *UpdateAnalyticsDashboardReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateAnalyticsDashboards(): %v", err.Error())
		return &err
	}

	if args.Name == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Name"
		core.Error("UpdateAnalyticsDashboards(): %v: Name is required", err.Error())
		return &err
	}

	ctx := r.Context()
	id := int64(args.ID)

	dashboard, err := s.Storage.GetAnalyticsDashboardByID(ctx, id)
	if err != nil {
		core.Error("UpdateAnalyticsDashboards(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	wasError := false
	if dashboard.Name != args.Name {
		if err := s.Storage.UpdateAnalyticsDashboardByID(ctx, id, "Name", args.Name); err != nil {
			core.Error("UpdateAnalyticsDashboards(): %v", err.Error())
			wasError = true
		}
	}

	if dashboard.Order != args.Order {
		if err := s.Storage.UpdateAnalyticsDashboardByID(ctx, id, "Order", args.Order); err != nil {
			core.Error("UpdateAnalyticsDashboards(): %v", err.Error())
			wasError = true
		}
	}

	if dashboard.LookerID != int64(args.LookerID) {
		if err := s.Storage.UpdateAnalyticsDashboardByID(ctx, id, "LookerID", int64(args.LookerID)); err != nil {
			fmt.Println(err)
			core.Error("UpdateAnalyticsDashboards(): %v", err.Error())
			wasError = true
		}
	}

	if int64(dashboard.Category.ID) != int64(args.CategoryID) {
		if err := s.Storage.UpdateAnalyticsDashboardByID(ctx, id, "Category", int64(args.CategoryID)); err != nil {
			core.Error("UpdateAnalyticsDashboards(): %v", err.Error())
			wasError = true
		}
	}

	if dashboard.Admin != args.AdminOnly {
		if err := s.Storage.UpdateAnalyticsDashboardByID(ctx, id, "Admin", args.AdminOnly); err != nil {
			core.Error("UpdateAnalyticsDashboards(): %v", err.Error())
			wasError = true
		}
	}

	if dashboard.Premium != args.Premium {
		if err := s.Storage.UpdateAnalyticsDashboardByID(ctx, id, "Premium", args.Premium); err != nil {
			core.Error("UpdateAnalyticsDashboards(): %v", err.Error())
			wasError = true
		}
	}

	if wasError {
		core.Error("UpdateAnalyticsDashboards(): There was an issue updating the dashboard")
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type AdminDashboard struct {
	URL  string `json:"url"`
	Live bool   `json:"live"`
}

type FetchAdminDashboardsArgs struct {
	CompanyCode string `json:"company_code"`
}

type FetchAdminDashboardsReply struct {
	Dashboards map[string][]AdminDashboard `json:"dashboards"`
	MainTabs   []string                    `json:"tabs"`
	SubTabs    map[string][]string         `json:"sub_tabs"`
}

// TODO: turn this back on later this week (Friday Aug 20th 2021 - Waiting on Tapan to finalize dash and add automatic buyer filtering)
func (s *OpsService) FetchAdminDashboards(r *http.Request, args *FetchAdminDashboardsArgs, reply *FetchAdminDashboardsReply) error {
	ctx := r.Context()
	reply.Dashboards = make(map[string][]AdminDashboard, 0)

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("FetchAdminDashboards(): %v", err.Error())
		return &err
	}

	user := ctx.Value(middleware.Keys.UserKey)
	if user == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("FetchUsageDashboard(): %v", err.Error())
		return &err
	}

	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)
	requestID, ok := claims["sub"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("FetchUsageDashboard(): %v: Failed to parse user ID", err.Error())
		return &err
	}

	customerCode := args.CompanyCode

	dashboards, err := s.Storage.GetAnalyticsDashboards(ctx)
	if err != nil {
		core.Error("FetchAdminDashboards(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	categories := make([]looker.AnalyticsDashboardCategory, 0)
	subCategories := make(map[string][]looker.AnalyticsDashboardCategory, 0)

	for _, dashboard := range dashboards {
		if dashboard.CustomerCode == customerCode {
			dashCustomerCode := customerCode

			if s.Env == "local" {
				dashCustomerCode = "twenty-four-entertainment"
			}

			// If the dashboard is assigned to a parent category, assign it to the normal label / dashboard system
			if dashboard.Category.ParentCategoryID < 0 {
				parentSubCategories, err := s.Storage.GetAnalyticsDashboardSubCategoriesByCategoryID(ctx, dashboard.Category.ID)
				if err != nil {
					core.Error("FetchAdminDashboards(): %v", err.Error())
					continue
				}

				// If a category has a dashboard assigned to it before a sub category is, this get a little weird. We will prioritize the sub tabs over an individual dashboard
				if len(parentSubCategories) > 0 {
					// TODO: Not logging an error here because this has the potential to be really spammy. There is a better fix here at the admin tool level
					continue
				}

				_, ok := reply.Dashboards[dashboard.Category.Label]
				if !ok {
					reply.Dashboards[dashboard.Category.Label] = make([]AdminDashboard, 0)
					categories = append(categories, dashboard.Category)
				}

				url, err := s.LookerClient.BuildGeneralPortalLookerURLWithDashID(fmt.Sprintf("%d", dashboard.LookerID), dashCustomerCode, requestID, r.Header.Get("Origin"))
				if err != nil {
					core.Error("FetchAdminDashboards(): %v", err.Error())
					continue
				}

				reply.Dashboards[dashboard.Category.Label] = append(reply.Dashboards[dashboard.Category.Label], AdminDashboard{
					URL:  url,
					Live: !dashboard.Admin,
				})
			} else {
				// Find the parent category information
				parentCategory, err := s.Storage.GetAnalyticsDashboardCategoryByID(ctx, dashboard.Category.ParentCategoryID)
				if err != nil {
					core.Error("FetchAdminDashboards(): %v", err.Error())
					continue
				}

				categoryLabel := fmt.Sprintf("%s/%s", parentCategory.Label, dashboard.Category.Label)

				if _, ok := subCategories[parentCategory.Label]; !ok {
					subCategories[parentCategory.Label] = make([]looker.AnalyticsDashboardCategory, 0)
					categories = append(categories, parentCategory)
				}

				subCategories[parentCategory.Label] = append(subCategories[parentCategory.Label], dashboard.Category)

				// Setup the usual system for the parent category
				if _, ok := reply.Dashboards[categoryLabel]; !ok {
					reply.Dashboards[categoryLabel] = make([]AdminDashboard, 0)
				}

				url, err := s.LookerClient.BuildGeneralPortalLookerURLWithDashID(fmt.Sprintf("%d", dashboard.LookerID), dashCustomerCode, requestID, r.Header.Get("Origin"))
				if err != nil {
					continue
				}

				reply.Dashboards[categoryLabel] = append(reply.Dashboards[categoryLabel], AdminDashboard{
					URL:  url,
					Live: !dashboard.Admin,
				})
			}
		}
	}

	reply.MainTabs = make([]string, 0)
	reply.SubTabs = make(map[string][]string)

	sort.Slice(categories, func(i int, j int) bool {
		return categories[i].Order > categories[j].Order
	})

	for _, category := range categories {
		reply.MainTabs = append(reply.MainTabs, category.Label)

		// If this category has sub categories, sort them and then add them back to the map
		if subTabs, ok := subCategories[category.Label]; ok {
			sort.Slice(subTabs, func(i int, j int) bool {
				return subTabs[i].Order > subTabs[j].Order
			})

			if _, ok := reply.SubTabs[category.Label]; !ok {
				reply.SubTabs[category.Label] = make([]string, 0)
			}

			for _, subTab := range subTabs {
				reply.SubTabs[category.Label] = append(reply.SubTabs[category.Label], subTab.Label)
			}
		}
	}

	return nil
}

type FetchAllLookerDashboardsArgs struct{}

type FetchAllLookerDashboardsReply struct {
	Dashboards []looker.LookerDashboard `json:"dashboards"`
}

func (s *OpsService) FetchAllLookerDashboards(r *http.Request, args *FetchAllLookerDashboardsArgs, reply *FetchAllLookerDashboardsReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("FetchAllLookerDashboards(): %v", err.Error())
		return &err
	}

	reply.Dashboards = s.LookerDashboardCache
	return nil
}

func (s *OpsService) RefreshLookerDashboardCache() error {
	dashboards, err := s.LookerClient.FetchCurrentLookerDashboards()
	if err != nil {
		return err
	}

	s.LookerDashboardCache = dashboards
	return nil
}
