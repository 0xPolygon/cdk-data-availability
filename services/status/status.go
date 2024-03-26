package status

import (
	"context"
	"time"

	dataavailability "github.com/0xPolygon/cdk-data-availability"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/synchronizer"
)

// APISTATUS is the namespace of the status service
const APISTATUS = "status"

type Status struct {
	Uptime           string
	Version          string
	KeyCount         uint64
	BackfillProgress uint64
}

// StatusEndpoints contains implementations for the "status" RPC endpoints
type StatusEndpoints struct {
	db        db.DB
	startTime time.Time
}

// NewStatusEndpoints returns StatusEndpoints
func NewStatusEndpoints(db db.DB) *StatusEndpoints {
	return &StatusEndpoints{
		db:        db,
		startTime: time.Now(),
	}
}

// GetStatus returns the status of the service
func (s *StatusEndpoints) GetStatus() (interface{}, rpc.Error) {
	ctx := context.Background()
	uptime := time.Since(s.startTime).String()

	var rowCount uint64
	err := s.db.GetOffchainDataRowCount(ctx, &rowCount)
	if err != nil {
		log.Errorf("failed to get the key count from the offchain_data table: %v", err)
	}

	backfillProgress, err := s.db.GetLastProcessedBlock(ctx, synchronizer.L1SyncTask)
	if err != nil {
		log.Errorf("failed to get last block processed by the synchronizer: %v", err)
	}

	status := Status{
		Version:          dataavailability.Version,
		Uptime:           uptime,
		KeyCount:         rowCount,
		BackfillProgress: backfillProgress,
	}

	return status, nil
}
