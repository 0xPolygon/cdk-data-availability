package status

import (
	"context"
	"time"

	dataavailability "github.com/0xPolygon/cdk-data-availability"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/synchronizer"
	"github.com/0xPolygon/cdk-data-availability/types"
)

// APISTATUS is the namespace of the status service
const APISTATUS = "status"

// GapsDetector is an interface for detecting gaps in the offchain data
type GapsDetector interface {
	// Gaps returns a map of gaps in the offchain data
	Gaps() map[uint64]uint64
}

// Endpoints contains implementations for the "status" RPC endpoints
type Endpoints struct {
	db           db.DB
	startTime    time.Time
	gapsDetector GapsDetector
}

// NewEndpoints returns Endpoints
func NewEndpoints(db db.DB, gapsDetector GapsDetector) *Endpoints {
	return &Endpoints{
		db:           db,
		startTime:    time.Now(),
		gapsDetector: gapsDetector,
	}
}

// GetStatus returns the status of the service
func (s *Endpoints) GetStatus() (interface{}, rpc.Error) {
	ctx := context.Background()
	uptime := time.Since(s.startTime).String()

	rowCount, err := s.db.CountOffchainData(ctx)
	if err != nil {
		log.Errorf("failed to get the key count from the offchain_data table: %v", err)
	}

	backfillProgress, err := s.db.GetLastProcessedBlock(ctx, string(synchronizer.L1SyncTask))
	if err != nil {
		log.Errorf("failed to get last block processed by the synchronizer: %v", err)
	}

	return types.DACStatus{
		Version:               dataavailability.Version,
		Uptime:                uptime,
		KeyCount:              rowCount,
		BackfillProgress:      backfillProgress,
		OffchainDataGapsExist: len(s.gapsDetector.Gaps()) > 0,
	}, nil
}
