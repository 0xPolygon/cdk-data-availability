package synchronizer

import (
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/sequencer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"
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

func TestBatchSynchronizer_Resolve(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		// sequencer mocks
		getSequenceBatchArgs    []interface{}
		getSequenceBatchReturns []interface{}
		// etherman mocks
		getCurrentDataCommitteeReturns []interface{}
		// clientFactory mocks
		newArgs [][]interface{}
		// client mocks
		getOffChainDataArgs    [][]interface{}
		getOffChainDataReturns [][]interface{}

		isErrorExpected bool
		errorString     string
	}

	data := common.HexToHash("0xFFFF").Bytes()
	batchKey := batchKey{
		number: 1,
		hash:   crypto.Keccak256Hash(data),
	}

	testFn := func(config testConfig) {
		clientMock := new(mocks.ClientMock)
		ethermanMock := new(mocks.EthermanMock)
		sequencerMock := new(mocks.SequencerTrackerMock)
		clientFactoryMock := new(mocks.ClientFactoryMock)

		if config.getSequenceBatchArgs != nil && config.getSequenceBatchReturns != nil {
			sequencerMock.On("GetSequenceBatch", config.getSequenceBatchArgs...).Return(
				config.getSequenceBatchReturns...).Once()
		}

		if config.getCurrentDataCommitteeReturns != nil {
			ethermanMock.On("GetCurrentDataCommittee").Return(config.getCurrentDataCommitteeReturns...).Once()
		}

		if config.newArgs != nil {
			for _, args := range config.newArgs {
				clientFactoryMock.On("New", args...).Return(clientMock).Once()
			}
		}

		if config.getOffChainDataArgs != nil && config.getOffChainDataReturns != nil {
			for i, args := range config.getOffChainDataArgs {
				clientMock.On("GetOffChainData", args...).Return(
					config.getOffChainDataReturns[i]...).Once()
			}
		}

		batchSyncronizer := &BatchSynchronizer{
			client:           ethermanMock,
			sequencer:        sequencerMock,
			rpcClientFactory: clientFactoryMock,
		}

		offChainData, err := batchSyncronizer.resolve(batchKey)
		if config.isErrorExpected {
			if config.errorString != "" {
				require.ErrorContains(t, err, config.errorString)
			} else {
				require.Error(t, err)
			}
		} else {
			require.NoError(t, err)
			require.Equal(t, batchKey.hash, offChainData.Key)
			require.Equal(t, data, offChainData.Value)
		}

		clientMock.AssertExpectations(t)
		ethermanMock.AssertExpectations(t)
		sequencerMock.AssertExpectations(t)
		clientFactoryMock.AssertExpectations(t)
	}

	t.Run("Got data from sequencer", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getSequenceBatchArgs: []interface{}{batchKey.number},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      rpc.ArgUint64(batchKey.number),
				BatchL2Data: rpc.ArgBytes(data),
			}, nil},
		})
	})

	t.Run("Got data from a committee member", func(t *testing.T) {
		t.Parallel()

		committee := &etherman.DataCommittee{
			Members: []etherman.DataCommitteeMember{
				{
					Addr: common.HexToAddress("0x4321"),
					URL:  "http://url-22",
				},
				{
					Addr: common.HexToAddress("0x5321"),
					URL:  "http://url-22",
				},
			},
		}

		testFn(testConfig{
			isErrorExpected:                false,
			getOffChainDataArgs:            [][]interface{}{{mock.Anything, batchKey.hash}},
			getOffChainDataReturns:         [][]interface{}{{data, nil}},
			getSequenceBatchArgs:           []interface{}{batchKey.number},
			getSequenceBatchReturns:        []interface{}{nil, errors.New("error")},
			getCurrentDataCommitteeReturns: []interface{}{committee, nil},
			newArgs:                        [][]interface{}{{committee.Members[0].URL}},
		})
	})

	t.Run("No committee member has given batch - they return error", func(t *testing.T) {
		t.Parallel()

		committee := &etherman.DataCommittee{
			Members: []etherman.DataCommitteeMember{
				{
					Addr: common.HexToAddress("0x1234"),
					URL:  "http://url-1",
				},
				{
					Addr: common.HexToAddress("0x1235"),
					URL:  "http://url-2",
				},
			},
		}

		testFn(testConfig{
			getSequenceBatchArgs:           []interface{}{batchKey.number},
			getSequenceBatchReturns:        []interface{}{nil, errors.New("error")},
			getCurrentDataCommitteeReturns: []interface{}{committee, nil},
			newArgs: [][]interface{}{
				{committee.Members[0].URL},
				{committee.Members[1].URL}},
			getOffChainDataArgs: [][]interface{}{
				{mock.Anything, batchKey.hash},
				{mock.Anything, batchKey.hash},
			},
			getOffChainDataReturns: [][]interface{}{
				{nil, errors.New("error")}, // member doesn't have batch
				{nil, errors.New("error")}, // member doesn't have batch
			},
			isErrorExpected: true,
			errorString:     "no data found for number",
		})
	})

	t.Run("No committee member has given batch - they return another hash", func(t *testing.T) {
		t.Parallel()

		committee := &etherman.DataCommittee{
			Members: []etherman.DataCommitteeMember{
				{
					Addr: common.HexToAddress("0x123456"),
					URL:  "http://url-11",
				},
				{
					Addr: common.HexToAddress("0x12357"),
					URL:  "http://url-22",
				},
			},
		}

		testFn(testConfig{
			isErrorExpected: true,
			errorString:     "no data found for number",
			newArgs: [][]interface{}{
				{committee.Members[0].URL},
				{committee.Members[1].URL}},
			getOffChainDataArgs: [][]interface{}{
				{mock.Anything, batchKey.hash},
				{mock.Anything, batchKey.hash},
			},
			getOffChainDataReturns: [][]interface{}{
				{[]byte{0, 0, 0, 1}, nil}, // member doesn't have batch
				{[]byte{0, 0, 0, 1}, nil}, // member doesn't have batch
			},
			getSequenceBatchArgs:           []interface{}{batchKey.number},
			getSequenceBatchReturns:        []interface{}{nil, errors.New("error")},
			getCurrentDataCommitteeReturns: []interface{}{committee, nil},
		})
	})
}
