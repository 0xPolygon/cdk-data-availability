package synchronizer

import (
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestBatchSynchronizer_ResolveCommittee(t *testing.T) {
	t.Parallel()

	t.Run("error getting committee", func(t *testing.T) {
		t.Parallel()

		ethermanMock := new(mocks.EthermanMock)
		ethermanMock.On("GetCurrentDataCommittee").Return(nil, errors.New("error")).Once()

		batchSyncronizer := &BatchSynchronizer{
			client: ethermanMock,
		}

		require.Error(t, batchSyncronizer.resolveCommittee())

		ethermanMock.AssertExpectations(t)
	})

	t.Run("resolving committee successful", func(t *testing.T) {
		t.Parallel()

		committee := &etherman.DataCommittee{
			Members: []etherman.DataCommitteeMember{
				{
					Addr: common.HexToAddress("0x123312415"),
					URL:  "http://url-1",
				},
				{
					Addr: common.HexToAddress("0x123312416"),
					URL:  "http://url-2",
				},
				{
					Addr: common.HexToAddress("0x123312417"),
					URL:  "http://url-3",
				},
			},
		}
		ethermanMock := new(mocks.EthermanMock)
		ethermanMock.On("GetCurrentDataCommittee").Return(committee, nil).Once()

		batchSyncronizer := &BatchSynchronizer{
			client: ethermanMock,
		}

		require.NoError(t, batchSyncronizer.resolveCommittee())
		require.Len(t, batchSyncronizer.committee, len(committee.Members))

		ethermanMock.AssertExpectations(t)
	})
}
