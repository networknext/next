package billing

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
	"github.com/golang/protobuf/proto"
)

type GoogleBigTableClient struct {
	Table *bigtable.Table
}

func (bt *GoogleBigTableClient) Bill(ctx context.Context, sessionID uint64, entry *Entry) error {
	data, err := proto.Marshal(entry)
	if err != nil {
		return err
	}

	mutation := bigtable.NewMutation()
	mutation.Set("data", "d", bigtable.Time(time.Now()), data)

	timeTag := entry.Timestamp - entry.TimestampStart
	rowKey := fmt.Sprintf("%016x#%s#%d", sessionID, entry.Request.BuyerID.Name, timeTag)

	return bt.Table.Apply(ctx, rowKey, mutation)
}
