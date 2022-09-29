package common

import (
	"context"
	"fmt"
	"os"

	"github.com/networknext/backend/modules-old/backend"
	"github.com/networknext/backend/modules-old/routing"
	"github.com/networknext/backend/modules/core"
)

func ValidateISPFile(ctx context.Context, env string, ispStorageName string) error {
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

func ValidateCityFile(ctx context.Context, env string, cityStorageName string) error {
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

func ValidateDatabaseFile(databaseFile *os.File, databaseNew *routing.DatabaseBinWrapper) error {
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

func ValidateOverlayFile(overlayFile *os.File, overlayNew *routing.OverlayBinWrapper) error {
	if err := backend.DecodeOverlayWrapper(overlayFile, overlayNew); err != nil {
		core.Error("validateOverlayFile() failed to decode database file: %v", err)
		return err
	}

	return nil
}
