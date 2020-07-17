package billing

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LocalBiller struct {
	Logger log.Logger

	submitted uint64
}

func (local *LocalBiller) Bill(ctx context.Context, entry *BillingEntry) error {
	local.submitted++

	if local.Logger == nil {
		return errors.New("no logger for local biller, can't display entry")
	}

	level.Info(local.Logger).Log("msg", "submitted billing entry")

	output := fmt.Sprintf("%#v", entry)
	level.Debug(local.Logger).Log("entry", output)

	// values, _, err := entry.Save()
	// if err != nil {
	// 	return err
	// }

	// bytes, err := json.MarshalIndent(values, "", "\t")
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(string(bytes))
	return nil
}

func (local *LocalBiller) NumSubmitted() uint64 {
	return local.submitted
}

func (local *LocalBiller) NumQueued() uint64 {
	return 0
}

func (local *LocalBiller) NumFlushed() uint64 {
	return local.submitted
}
