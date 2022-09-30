package common

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/networknext/backend/modules-old/backend"
	"github.com/networknext/backend/modules-old/routing"
	"github.com/networknext/backend/modules/core"
)

// ------------------------------------------------------------------------

type BinType string

const BIN_DATABASE BinType = "city"
const BIN_OVERLAY BinType = "isp"

type BinFile struct {
	Name string
	Path string
	Type BinType
}

func (binFile *BinFile) Validate(fileRef *os.File) error {
	switch binFile.Type {
	case BIN_DATABASE:
		return validateDatabaseFile(fileRef, &routing.DatabaseBinWrapper{})
	case BIN_OVERLAY:
		return validateOverlayFile(fileRef, &routing.OverlayBinWrapper{})
	default:
		return errors.New("unknown bin file type")
	}
}

// todo: revisit these as the old modules are cleaned up
func validateDatabaseFile(databaseFile *os.File, databaseNew *routing.DatabaseBinWrapper) error {
	if err := backend.DecodeBinWrapper(databaseFile, databaseNew); err != nil {
		core.Error("validateDatabaseFile() failed to decode database file: %v", err)
		return err
	}

	if databaseNew.IsEmpty() {
		// Don't want to use an empty bin wrapper
		// so early out here and use existing array and hash
		err := fmt.Errorf("new database file is empty, keeping previous values")
		core.Error(err.Error())
		return err
	}

	return nil
}

// todo: revisit these as the old modules are cleaned up
func validateOverlayFile(overlayFile *os.File, overlayNew *routing.OverlayBinWrapper) error {
	if err := backend.DecodeOverlayWrapper(overlayFile, overlayNew); err != nil {
		core.Error("validateOverlayFile() failed to decode database file: %v", err)
		return err
	}

	return nil
}

// ------------------------------------------------------------------------

type MaxmindType string

const MAXMIND_CITY MaxmindType = "city"
const MAXMIND_ISP MaxmindType = "isp"

type MaxmindFile struct {
	Name        string
	Path        string
	Type        MaxmindType
	DownloadURL string
	Env         string
}

func (mmdbFile *MaxmindFile) Validate(ctx context.Context) error {
	switch mmdbFile.Type {
	case MAXMIND_ISP:
		return validateISPFile(ctx, mmdbFile.Env, mmdbFile.Path)
	case MAXMIND_CITY:
		return validateISPFile(ctx, mmdbFile.Env, mmdbFile.Path)
	default:
		return errors.New("unknown maxmind file type")
	}
}

// todo: revisit these as the old modules are cleaned up
func validateISPFile(ctx context.Context, env string, ispStorageName string) error {
	mmdb := &routing.MaxmindDB{
		IspFile:   ispStorageName,
		IsStaging: env == "staging",
	}

	// Validate the ISP file
	if err := mmdb.OpenISP(ctx); err != nil {
		return err
	}

	if err := mmdb.ValidateISP(); err != nil {
		return err
	}

	return nil
}

// todo: revisit these as the old modules are cleaned up
func validateCityFile(ctx context.Context, env string, cityStorageName string) error {
	mmdb := &routing.MaxmindDB{
		CityFile:  cityStorageName,
		IsStaging: env == "staging",
	}

	// Validate the City file
	if err := mmdb.OpenCity(ctx); err != nil {
		return err
	}

	if err := mmdb.ValidateCity(); err != nil {
		return err
	}

	return nil
}
