package synchronizer

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	elderberryValidium "github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	etrogValidium "github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/sequencer"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
)

func TestBatchSynchronizer_ResolveCommittee(t *testing.T) {
	t.Parallel()

	t.Run("error getting committee", func(t *testing.T) {
		t.Parallel()

		ethermanMock := mocks.NewEtherman(t)
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
		ethermanMock := mocks.NewEtherman(t)
		ethermanMock.On("GetCurrentDataCommittee").Return(committee, nil).Once()

		batchSyncronizer := &BatchSynchronizer{
			client: ethermanMock,
		}

		require.NoError(t, batchSyncronizer.resolveCommittee())
		require.Equal(t, len(committee.Members), batchSyncronizer.committee.Length())

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
	batchKey := types.BatchKey{
		Number: 1,
		Hash:   crypto.Keccak256Hash(data),
	}

	testFn := func(config testConfig) {
		clientMock := mocks.NewClient(t)
		ethermanMock := mocks.NewEtherman(t)
		sequencerMock := mocks.NewSequencerTracker(t)
		clientFactoryMock := mocks.NewClientFactory(t)

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
			committee:        NewCommitteeMapSafe(),
		}

		offChainData, err := batchSyncronizer.resolve(context.Background(), batchKey)
		if config.isErrorExpected {
			if config.errorString != "" {
				require.ErrorContains(t, err, config.errorString)
			} else {
				require.Error(t, err)
			}
		} else {
			require.NoError(t, err)
			require.Equal(t, batchKey.Hash, offChainData.Key)
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
			getSequenceBatchArgs: []interface{}{context.Background(), batchKey.Number},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(batchKey.Number),
				BatchL2Data: types.ArgBytes(data),
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
			getOffChainDataArgs:            [][]interface{}{{mock.Anything, batchKey.Hash}},
			getOffChainDataReturns:         [][]interface{}{{data, nil}},
			getSequenceBatchArgs:           []interface{}{context.Background(), batchKey.Number},
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
			getSequenceBatchArgs:           []interface{}{context.Background(), batchKey.Number},
			getSequenceBatchReturns:        []interface{}{nil, errors.New("error")},
			getCurrentDataCommitteeReturns: []interface{}{committee, nil},
			newArgs: [][]interface{}{
				{committee.Members[0].URL},
				{committee.Members[1].URL}},
			getOffChainDataArgs: [][]interface{}{
				{mock.Anything, batchKey.Hash},
				{mock.Anything, batchKey.Hash},
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
				{mock.Anything, batchKey.Hash},
				{mock.Anything, batchKey.Hash},
			},
			getOffChainDataReturns: [][]interface{}{
				{[]byte{0, 0, 0, 1}, nil}, // member doesn't have batch
				{[]byte{0, 0, 0, 1}, nil}, // member doesn't have batch
			},
			getSequenceBatchArgs:           []interface{}{context.Background(), batchKey.Number},
			getSequenceBatchReturns:        []interface{}{nil, errors.New("error")},
			getCurrentDataCommitteeReturns: []interface{}{committee, nil},
		})
	})
}

func TestBatchSynchronizer_HandleEvent(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		// etherman mock
		getTxArgs    []interface{}
		getTxReturns []interface{}
		// db mock
		listOffchainDataArgs         []interface{}
		listOffchainDataReturns      []interface{}
		storeMissingBatchKeysArgs    []interface{}
		storeMissingBatchKeysReturns []interface{}

		isErrorExpected bool
	}

	to := common.HexToAddress("0xFFFF")
	event := &etrogValidium.PolygonvalidiumetrogSequenceBatches{
		Raw: ethTypes.Log{
			TxHash: common.BytesToHash([]byte{0, 1, 2, 3}),
		},
		NumBatch: 10,
	}
	batchL2Data := []byte{1, 2, 3, 4, 5, 6}
	txHash := crypto.Keccak256Hash(batchL2Data)

	batchData := []etrogValidium.PolygonValidiumEtrogValidiumBatchData{
		{
			TransactionsHash: txHash,
		},
	}

	a, err := abi.JSON(strings.NewReader(etrogValidium.PolygonvalidiumetrogABI))
	require.NoError(t, err)

	methodDefinition, ok := a.Methods["sequenceBatchesValidium"]
	require.True(t, ok)

	data, err := methodDefinition.Inputs.Pack(batchData, common.HexToAddress("0xABCD"), []byte{22, 23, 24})
	require.NoError(t, err)

	tx := ethTypes.NewTx(
		&ethTypes.LegacyTx{
			Nonce:    0,
			GasPrice: big.NewInt(10_000),
			Gas:      21_000,
			To:       &to,
			Value:    ethgo.Ether(1),
			Data:     append(methodDefinition.ID, data...),
		})

	testFn := func(t *testing.T, config testConfig) {
		t.Helper()

		dbMock := mocks.NewDB(t)
		ethermanMock := mocks.NewEtherman(t)

		if config.getTxArgs != nil && config.getTxReturns != nil {
			ethermanMock.On("GetTx", config.getTxArgs...).Return(
				config.getTxReturns...).Once()
		}

		if config.listOffchainDataArgs != nil && config.listOffchainDataReturns != nil {
			dbMock.On("ListOffChainData", config.listOffchainDataArgs...).Return(
				config.listOffchainDataReturns...).Once()
		}

		if config.storeMissingBatchKeysArgs != nil && config.storeMissingBatchKeysReturns != nil {
			dbMock.On("StoreMissingBatchKeys", config.storeMissingBatchKeysArgs...).Return(
				config.storeMissingBatchKeysReturns...).Once()
		}

		batchSynronizer := &BatchSynchronizer{
			db:     dbMock,
			client: ethermanMock,
		}

		err := batchSynronizer.handleEvent(context.Background(), event)
		if config.isErrorExpected {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		dbMock.AssertExpectations(t)
		ethermanMock.AssertExpectations(t)
	}

	t.Run("could not get tx data", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getTxArgs:       []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:    []interface{}{nil, false, errors.New("error")},
			isErrorExpected: true,
		})
	})

	t.Run("invalid tx data", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getTxArgs: []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns: []interface{}{ethTypes.NewTx(
				&ethTypes.LegacyTx{
					Nonce:    0,
					GasPrice: big.NewInt(10_000),
					Gas:      21_000,
					To:       &to,
					Value:    ethgo.Ether(1),
					Data:     []byte{0, 1, 3, 4, 5, 6, 7}, //some invalid data
				},
			), true, nil},
			isErrorExpected: true,
		})
	})

	t.Run("Error getting offchain data", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getTxArgs:               []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:            []interface{}{tx, true, nil},
			listOffchainDataArgs:    []interface{}{mock.Anything, []common.Hash{txHash}},
			listOffchainDataReturns: []interface{}{nil, errors.New("error")},
			isErrorExpected:         true,
		})
	})

	t.Run("doesn't have batch in storage - successfully stored (Elderberry fork)", func(t *testing.T) {
		t.Parallel()

		a, err := abi.JSON(strings.NewReader(elderberryValidium.PolygonvalidiumetrogABI))
		require.NoError(t, err)

		methodDefinition, ok := a.Methods["sequenceBatchesValidium"]
		require.True(t, ok)

		data, err := methodDefinition.Inputs.Pack(batchData, uint64(10), uint64(20), common.HexToAddress("0xABCD"), []byte{22, 23, 24})
		require.NoError(t, err)

		localTx := ethTypes.NewTx(
			&ethTypes.LegacyTx{
				Nonce:    0,
				GasPrice: big.NewInt(10_000),
				Gas:      21_000,
				To:       &to,
				Value:    ethgo.Ether(1),
				Data:     append(methodDefinition.ID, data...),
			})

		testFn(t, testConfig{
			getTxArgs:               []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:            []interface{}{localTx, true, nil},
			listOffchainDataArgs:    []interface{}{mock.Anything, []common.Hash{txHash}},
			listOffchainDataReturns: []interface{}{nil, nil},
			storeMissingBatchKeysArgs: []interface{}{
				mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeMissingBatchKeysReturns: []interface{}{nil},
			isErrorExpected:              false,
		})
	})

	t.Run("doesn't have batch in storage - successfully stored (Etrog fork)", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getTxArgs:               []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:            []interface{}{tx, true, nil},
			listOffchainDataArgs:    []interface{}{mock.Anything, []common.Hash{txHash}},
			listOffchainDataReturns: []interface{}{nil, nil},
			storeMissingBatchKeysArgs: []interface{}{
				mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeMissingBatchKeysReturns: []interface{}{nil},
			isErrorExpected:              false,
		})
	})

	t.Run("doesn't have batch in storage - store fails", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			isErrorExpected:         true,
			listOffchainDataArgs:    []interface{}{mock.Anything, []common.Hash{txHash}},
			listOffchainDataReturns: []interface{}{nil, nil},
			storeMissingBatchKeysArgs: []interface{}{
				mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeMissingBatchKeysReturns: []interface{}{errors.New("error")},
			getTxArgs:                    []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:                 []interface{}{tx, true, nil},
		})
	})

	t.Run("have batch in storage already no error", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			isErrorExpected:      false,
			listOffchainDataArgs: []interface{}{mock.Anything, []common.Hash{txHash}},
			listOffchainDataReturns: []interface{}{
				[]types.OffChainData{
					{
						Key:   txHash,
						Value: batchL2Data,
					},
				}, nil,
			},
			getTxArgs:    []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns: []interface{}{tx, true, nil},
		})
	})
}

func TestBatchSynchronizer_ProcessMissingBatches(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dbMock := mocks.NewDB(t)
	dbMock.On("GetMissingBatchKeys", mock.Anything, mock.Anything).Return(
		[]types.BatchKey{}, nil)

	batchSynronizer := &BatchSynchronizer{
		db:    dbMock,
		retry: time.Millisecond * 100,
		stop:  make(chan struct{}),
	}
	go batchSynronizer.processMissingBatches(ctx)

	// Wait for the retry interval and then signal to stop
	time.Sleep(time.Millisecond * 200)
	batchSynronizer.stop <- struct{}{}
	dbMock.AssertExpectations(t)
}

func TestBatchSynchronizer_HandleMissingBatches(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		// db mock
		getMissingBatchKeysArgs       []interface{}
		getMissingBatchKeysReturns    []interface{}
		storeOffChainDataArgs         []interface{}
		storeOffChainDataReturns      []interface{}
		deleteMissingBatchKeysArgs    []interface{}
		deleteMissingBatchKeysReturns []interface{}
		// sequencer mocks
		getSequenceBatchArgs    []interface{}
		getSequenceBatchReturns []interface{}

		isErrorExpected bool
	}

	batchL2Data := []byte{1, 2, 3, 4, 5, 6}
	txHash := crypto.Keccak256Hash(batchL2Data)

	testFn := func(t *testing.T, config testConfig) {
		t.Helper()

		dbMock := mocks.NewDB(t)
		ethermanMock := mocks.NewEtherman(t)
		sequencerMock := mocks.NewSequencerTracker(t)

		if config.getMissingBatchKeysArgs != nil && config.getMissingBatchKeysReturns != nil {
			dbMock.On("GetMissingBatchKeys", config.getMissingBatchKeysArgs...).Return(
				config.getMissingBatchKeysReturns...).Once()
		}

		if config.storeOffChainDataArgs != nil && config.storeOffChainDataReturns != nil {
			dbMock.On("StoreOffChainData", config.storeOffChainDataArgs...).Return(
				config.storeOffChainDataReturns...).Once()
		}

		if config.deleteMissingBatchKeysArgs != nil && config.deleteMissingBatchKeysReturns != nil {
			dbMock.On("DeleteMissingBatchKeys", config.deleteMissingBatchKeysArgs...).Return(
				config.deleteMissingBatchKeysReturns...).Once()
		}

		if config.getSequenceBatchArgs != nil && config.getSequenceBatchReturns != nil {
			sequencerMock.On("GetSequenceBatch", config.getSequenceBatchArgs...).Return(
				config.getSequenceBatchReturns...).Once()
		}

		batchSynronizer := &BatchSynchronizer{
			db:        dbMock,
			client:    ethermanMock,
			sequencer: sequencerMock,
		}

		err := batchSynronizer.handleMissingBatches(context.Background())
		if config.isErrorExpected {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		dbMock.AssertExpectations(t)
		ethermanMock.AssertExpectations(t)
		sequencerMock.AssertExpectations(t)
	}

	t.Run("Could not get missing batch keys", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getMissingBatchKeysArgs:    []interface{}{mock.Anything, uint(100)},
			getMissingBatchKeysReturns: []interface{}{nil, errors.New("error")},
			isErrorExpected:            true,
		})
	})

	t.Run("No missing batch keys found", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getMissingBatchKeysArgs:    []interface{}{mock.Anything, uint(100)},
			getMissingBatchKeysReturns: []interface{}{nil, nil},
			isErrorExpected:            false,
		})
	})

	t.Run("Missing batch key found", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getMissingBatchKeysArgs: []interface{}{mock.Anything, uint(100)},
			getMissingBatchKeysReturns: []interface{}{
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				nil,
			},
			storeOffChainDataArgs: []interface{}{mock.Anything,
				[]types.OffChainData{{
					Key:   txHash,
					Value: batchL2Data,
				}},
			},
			storeOffChainDataReturns: []interface{}{nil},
			deleteMissingBatchKeysArgs: []interface{}{mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			deleteMissingBatchKeysReturns: []interface{}{nil},
			getSequenceBatchArgs:          []interface{}{context.Background(), uint64(10)},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(10),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
			isErrorExpected: false,
		})
	})

	t.Run("DB error while storing missing batch", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getMissingBatchKeysArgs: []interface{}{mock.Anything, uint(100)},
			getMissingBatchKeysReturns: []interface{}{
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				nil,
			},
			storeOffChainDataArgs: []interface{}{mock.Anything,
				[]types.OffChainData{{
					Key:   txHash,
					Value: batchL2Data,
				}},
			},
			storeOffChainDataReturns: []interface{}{errors.New("error")},
			getSequenceBatchArgs:     []interface{}{context.Background(), uint64(10)},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(10),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
			isErrorExpected: true,
		})
	})

	t.Run("DB error while deleting missing batch entries", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getMissingBatchKeysArgs: []interface{}{mock.Anything, uint(100)},
			getMissingBatchKeysReturns: []interface{}{
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				nil,
			},
			storeOffChainDataArgs: []interface{}{mock.Anything,
				[]types.OffChainData{{
					Key:   txHash,
					Value: batchL2Data,
				}},
			},
			storeOffChainDataReturns: []interface{}{nil},
			deleteMissingBatchKeysArgs: []interface{}{mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			deleteMissingBatchKeysReturns: []interface{}{errors.New("error")},
			getSequenceBatchArgs:          []interface{}{context.Background(), uint64(10)},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(10),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
			isErrorExpected: true,
		})
	})

	/*t.Run("Invalid tx data", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getTxArgs: []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns: []interface{}{ethTypes.NewTx(
				&ethTypes.LegacyTx{
					Nonce:    0,
					GasPrice: big.NewInt(10_000),
					Gas:      21_000,
					To:       &to,
					Value:    ethgo.Ether(1),
					Data:     []byte{0, 1, 3, 4, 5, 6, 7}, //some invalid data
				},
			), true, nil},
			isErrorExpected: true,
		})
	})

	t.Run("doesn't have batch in storage - successfully stored", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getTxArgs:            []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:         []interface{}{tx, true, nil},
			getSequenceBatchArgs: []interface{}{event.NumBatch},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(event.NumBatch),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
			beginStateTransactionArgs: []interface{}{mock.Anything},
			storeMissingBatchKeysArgs: []interface{}{mock.Anything,
				[]types.OffChainData{{
					Key:   txHash,
					Value: batchL2Data,
				}},
				mock.Anything,
			},
			storeMissingBatchKeysReturns: []interface{}{nil},
			commitReturns:                   []interface{}{nil},
			isErrorExpected:                 false,
		})
	})

	t.Run("doesn't have batch in storage - store fails", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			isErrorExpected: true,
			storeMissingBatchKeysArgs: []interface{}{mock.Anything,
				[]types.BatchKey{{
					Number: 1,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeMissingBatchKeysReturns: []interface{}{errors.New("error")},
			beginStateTransactionArgs:       []interface{}{mock.Anything},
			rollbackArgs:                    []interface{}{mock.Anything},
			getTxArgs:                       []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:                    []interface{}{tx, true, nil},
			getSequenceBatchArgs:            []interface{}{event.NumBatch},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(event.NumBatch),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
		})
	})
	})*/
}

func TestBatchSynchronizer_HandleReorgs(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		getLastProcessedBlockReturns   []interface{}
		storeLastProcessedBlockReturns []interface{}
		reorg                          BlockReorg
	}

	testFn := func(config testConfig) {
		dbMock := mocks.NewDB(t)

		dbMock.On("GetLastProcessedBlock", mock.Anything, string(L1SyncTask)).Return(config.getLastProcessedBlockReturns...).Once()
		if config.storeLastProcessedBlockReturns != nil {
			dbMock.On("StoreLastProcessedBlock", mock.Anything, mock.Anything, string(L1SyncTask)).
				Return(config.storeLastProcessedBlockReturns...).Once()
		}

		reorgChan := make(chan BlockReorg)
		batchSynchronizer := &BatchSynchronizer{
			db:     dbMock,
			stop:   make(chan struct{}),
			reorgs: reorgChan,
		}

		go batchSynchronizer.handleReorgs(context.Background())

		reorgChan <- config.reorg

		batchSynchronizer.stop <- struct{}{}

		dbMock.AssertExpectations(t)
	}

	t.Run("Getting last processed block fails", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getLastProcessedBlockReturns: []interface{}{uint64(0), errors.New("error")},
			reorg: BlockReorg{
				Number: 10,
			},
		})
	})

	t.Run("Reorg block higher than what we have in db", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getLastProcessedBlockReturns: []interface{}{uint64(5), nil},
			reorg: BlockReorg{
				Number: 10,
			},
		})
	})

	t.Run("Reorg block lower than what we have in db, but db throws error", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getLastProcessedBlockReturns:   []interface{}{uint64(15), nil},
			storeLastProcessedBlockReturns: []interface{}{errors.New("error")},
			reorg: BlockReorg{
				Number: 10,
			},
		})
	})

	t.Run("Reorg block lower than what we have in db, store the block in db", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getLastProcessedBlockReturns:   []interface{}{uint64(25), nil},
			storeLastProcessedBlockReturns: []interface{}{nil},
			reorg: BlockReorg{
				Number: 15,
			},
		})
	})
}
