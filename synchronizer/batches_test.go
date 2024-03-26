package synchronizer

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/etherman"
	elderberryValidium "github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/elderberry/polygonvalidium"
	etrogValidium "github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/etrog/polygonvalidium"
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
			getSequenceBatchArgs: []interface{}{batchKey.Number},
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
			getSequenceBatchArgs:           []interface{}{batchKey.Number},
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
			getSequenceBatchArgs:           []interface{}{batchKey.Number},
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
			getSequenceBatchArgs:           []interface{}{batchKey.Number},
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
		beginStateTransactionArgs       []interface{}
		beginStateTransactionReturns    []interface{}
		storeUnresolvedBatchKeysArgs    []interface{}
		storeUnresolvedBatchKeysReturns []interface{}
		commitReturns                   []interface{}
		rollbackArgs                    []interface{}

		isErrorExpected bool
	}

	to := common.HexToAddress("0xFFFF")
	event := &etrogValidium.PolygonvalidiumSequenceBatches{
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

	a, err := abi.JSON(strings.NewReader(etrogValidium.PolygonvalidiumABI))
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
		dbMock := mocks.NewDB(t)
		txMock := mocks.NewTx(t)
		ethermanMock := mocks.NewEtherman(t)

		if config.getTxArgs != nil && config.getTxReturns != nil {
			ethermanMock.On("GetTx", config.getTxArgs...).Return(
				config.getTxReturns...).Once()
		}

		if config.beginStateTransactionArgs != nil {
			var returnArgs []interface{}
			if config.beginStateTransactionReturns != nil {
				returnArgs = config.beginStateTransactionReturns
			} else {
				returnArgs = append(returnArgs, txMock, nil)
			}

			dbMock.On("BeginStateTransaction", config.beginStateTransactionArgs...).Return(
				returnArgs...).Once()
		}

		if config.storeUnresolvedBatchKeysArgs != nil && config.storeUnresolvedBatchKeysReturns != nil {
			dbMock.On("StoreUnresolvedBatchKeys", config.storeUnresolvedBatchKeysArgs...).Return(
				config.storeUnresolvedBatchKeysReturns...).Once()
		}

		if config.commitReturns != nil {
			txMock.On("Commit", mock.Anything).Return(
				config.commitReturns...).Once()
		}

		if config.rollbackArgs != nil {
			txMock.On("Rollback", config.rollbackArgs...).Return(nil).Once()
		}

		batchSynronizer := &BatchSynchronizer{
			db:     dbMock,
			client: ethermanMock,
		}

		err := batchSynronizer.handleEvent(event)
		if config.isErrorExpected {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		dbMock.AssertExpectations(t)
		txMock.AssertExpectations(t)
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

	t.Run("doesn't have batch in storage - successfully stored (Elderberry fork)", func(t *testing.T) {
		t.Parallel()

		a, err := abi.JSON(strings.NewReader(elderberryValidium.PolygonvalidiumABI))
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
			getTxArgs:                 []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:              []interface{}{localTx, true, nil},
			beginStateTransactionArgs: []interface{}{mock.Anything},
			storeUnresolvedBatchKeysArgs: []interface{}{
				mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeUnresolvedBatchKeysReturns: []interface{}{nil},
			commitReturns:                   []interface{}{nil},
			isErrorExpected:                 false,
		})
	})

	t.Run("doesn't have batch in storage - successfully stored (Etrog fork)", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getTxArgs:                 []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:              []interface{}{tx, true, nil},
			beginStateTransactionArgs: []interface{}{mock.Anything},
			storeUnresolvedBatchKeysArgs: []interface{}{
				mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeUnresolvedBatchKeysReturns: []interface{}{nil},
			commitReturns:                   []interface{}{nil},
			isErrorExpected:                 false,
		})
	})

	t.Run("doesn't have batch in storage - begin state transaction fails", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			isErrorExpected:              true,
			beginStateTransactionArgs:    []interface{}{mock.Anything},
			beginStateTransactionReturns: []interface{}{nil, errors.New("error")},
			getTxArgs:                    []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:                 []interface{}{tx, true, nil},
		})
	})

	t.Run("doesn't have batch in storage - store fails", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			isErrorExpected: true,
			storeUnresolvedBatchKeysArgs: []interface{}{
				mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeUnresolvedBatchKeysReturns: []interface{}{errors.New("error")},
			beginStateTransactionArgs:       []interface{}{mock.Anything},
			rollbackArgs:                    []interface{}{mock.Anything},
			getTxArgs:                       []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:                    []interface{}{tx, true, nil},
		})
	})

	t.Run("doesn't have batch in storage - commit fails", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			isErrorExpected:           true,
			beginStateTransactionArgs: []interface{}{mock.Anything},
			storeUnresolvedBatchKeysArgs: []interface{}{
				mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeUnresolvedBatchKeysReturns: []interface{}{nil},
			commitReturns:                   []interface{}{errors.New("error")},
			getTxArgs:                       []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:                    []interface{}{tx, true, nil},
		})
	})
}

func TestBatchSynchronizer_HandleUnresolvedBatches(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		// db mock
		getUnresolvedBatchKeysArgs         []interface{}
		getUnresolvedBatchKeysReturns      []interface{}
		existsArgs                         []interface{}
		existsReturns                      []interface{}
		storeBeginStateTransactionArgs     []interface{}
		storeBeginStateTransactionReturns  []interface{}
		storeOffChainDataArgs              []interface{}
		storeOffChainDataReturns           []interface{}
		storeCommitReturns                 []interface{}
		storeRollbackArgs                  []interface{}
		deleteBeginStateTransactionArgs    []interface{}
		deleteBeginStateTransactionReturns []interface{}
		deleteUnresolvedBatchKeysArgs      []interface{}
		deleteUnresolvedBatchKeysReturns   []interface{}
		deleteCommitReturns                []interface{}
		deleteRollbackArgs                 []interface{}
		// sequencer mocks
		getSequenceBatchArgs    []interface{}
		getSequenceBatchReturns []interface{}

		isErrorExpected bool
	}

	batchL2Data := []byte{1, 2, 3, 4, 5, 6}
	txHash := crypto.Keccak256Hash(batchL2Data)

	testFn := func(t *testing.T, config testConfig) {
		dbMock := mocks.NewDB(t)
		storeTxMock := mocks.NewTx(t)
		deleteTxMock := mocks.NewTx(t)
		ethermanMock := mocks.NewEtherman(t)
		sequencerMock := mocks.NewSequencerTracker(t)

		if config.getUnresolvedBatchKeysArgs != nil && config.getUnresolvedBatchKeysReturns != nil {
			dbMock.On("GetUnresolvedBatchKeys", config.getUnresolvedBatchKeysArgs...).Return(
				config.getUnresolvedBatchKeysReturns...).Once()
		}

		if config.existsArgs != nil && config.existsReturns != nil {
			dbMock.On("Exists", config.existsArgs...).Return(
				config.existsReturns...).Once()
		}

		if config.storeBeginStateTransactionArgs != nil {
			var returnArgs []interface{}
			if config.storeBeginStateTransactionReturns != nil {
				returnArgs = config.storeBeginStateTransactionArgs
			} else {
				returnArgs = append(returnArgs, storeTxMock, nil)
			}

			dbMock.On("BeginStateTransaction", config.storeBeginStateTransactionArgs...).Return(
				returnArgs...).Once()
		}

		if config.storeOffChainDataArgs != nil && config.storeOffChainDataReturns != nil {
			dbMock.On("StoreOffChainData", config.storeOffChainDataArgs...).Return(
				config.storeOffChainDataReturns...).Once()
		}

		if config.storeCommitReturns != nil {
			storeTxMock.On("Commit", mock.Anything).Return(
				config.storeCommitReturns...).Once()
		}

		if config.storeRollbackArgs != nil {
			storeTxMock.On("Rollback", config.storeRollbackArgs...).Return(nil).Once()
		}

		if config.deleteBeginStateTransactionArgs != nil {
			var returnArgs []interface{}
			if config.deleteBeginStateTransactionReturns != nil {
				returnArgs = config.deleteBeginStateTransactionArgs
			} else {
				returnArgs = append(returnArgs, deleteTxMock, nil)
			}

			dbMock.On("BeginStateTransaction", config.deleteBeginStateTransactionArgs...).Return(
				returnArgs...).Once()
		}

		if config.deleteUnresolvedBatchKeysArgs != nil && config.deleteUnresolvedBatchKeysReturns != nil {
			dbMock.On("DeleteUnresolvedBatchKeys", config.deleteUnresolvedBatchKeysArgs...).Return(
				config.deleteUnresolvedBatchKeysReturns...).Once()
		}

		if config.deleteCommitReturns != nil {
			deleteTxMock.On("Commit", mock.Anything).Return(
				config.deleteCommitReturns...).Once()
		}

		if config.deleteRollbackArgs != nil {
			deleteTxMock.On("Rollback", config.deleteRollbackArgs...).Return(nil).Once()
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

		err := batchSynronizer.handleUnresolvedBatches()
		if config.isErrorExpected {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		dbMock.AssertExpectations(t)
		storeTxMock.AssertExpectations(t)
		deleteTxMock.AssertExpectations(t)
		ethermanMock.AssertExpectations(t)
		sequencerMock.AssertExpectations(t)
	}

	t.Run("Could not get unresolved batch keys", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getUnresolvedBatchKeysArgs:    []interface{}{mock.Anything},
			getUnresolvedBatchKeysReturns: []interface{}{nil, errors.New("error")},
			isErrorExpected:               true,
		})
	})

	t.Run("No unresolved batch keys found", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getUnresolvedBatchKeysArgs:    []interface{}{mock.Anything},
			getUnresolvedBatchKeysReturns: []interface{}{nil, nil},
			isErrorExpected:               false,
		})
	})

	t.Run("Unresolved batch key already resolved", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getUnresolvedBatchKeysArgs: []interface{}{mock.Anything},
			getUnresolvedBatchKeysReturns: []interface{}{
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				nil,
			},
			existsArgs:                      []interface{}{mock.Anything, txHash},
			existsReturns:                   []interface{}{true},
			deleteBeginStateTransactionArgs: []interface{}{mock.Anything},
			deleteUnresolvedBatchKeysArgs: []interface{}{mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			deleteUnresolvedBatchKeysReturns: []interface{}{nil},
			deleteCommitReturns:              []interface{}{nil},
			isErrorExpected:                  false,
		})
	})

	t.Run("Unresolved batch key found", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getUnresolvedBatchKeysArgs: []interface{}{mock.Anything},
			getUnresolvedBatchKeysReturns: []interface{}{
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				nil,
			},
			existsArgs:                     []interface{}{mock.Anything, txHash},
			existsReturns:                  []interface{}{false},
			storeBeginStateTransactionArgs: []interface{}{mock.Anything},
			storeOffChainDataArgs: []interface{}{mock.Anything,
				[]types.OffChainData{{
					Key:   txHash,
					Value: batchL2Data,
				}},
				mock.Anything,
			},
			storeOffChainDataReturns:        []interface{}{nil},
			storeCommitReturns:              []interface{}{nil},
			deleteBeginStateTransactionArgs: []interface{}{mock.Anything},
			deleteUnresolvedBatchKeysArgs: []interface{}{mock.Anything,
				[]types.BatchKey{{
					Number: 10,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			deleteUnresolvedBatchKeysReturns: []interface{}{nil},
			deleteCommitReturns:              []interface{}{nil},
			getSequenceBatchArgs:             []interface{}{uint64(10)},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(10),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
			isErrorExpected: false,
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
			storeUnresolvedBatchKeysArgs: []interface{}{mock.Anything,
				[]types.OffChainData{{
					Key:   txHash,
					Value: batchL2Data,
				}},
				mock.Anything,
			},
			storeUnresolvedBatchKeysReturns: []interface{}{nil},
			commitReturns:                   []interface{}{nil},
			isErrorExpected:                 false,
		})
	})

	t.Run("doesn't have batch in storage - begin state transaction fails", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			isErrorExpected:              true,
			beginStateTransactionArgs:    []interface{}{mock.Anything},
			beginStateTransactionReturns: []interface{}{nil, errors.New("error")},
			getTxArgs:                    []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:                 []interface{}{tx, true, nil},
			getSequenceBatchArgs:         []interface{}{event.NumBatch},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(event.NumBatch),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
		})
	})

	t.Run("doesn't have batch in storage - store fails", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			isErrorExpected: true,
			storeUnresolvedBatchKeysArgs: []interface{}{mock.Anything,
				[]types.BatchKey{{
					Number: 1,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeUnresolvedBatchKeysReturns: []interface{}{errors.New("error")},
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

	t.Run("doesn't have batch in storage - commit fails", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			isErrorExpected:           true,
			beginStateTransactionArgs: []interface{}{mock.Anything},
			storeUnresolvedBatchKeysArgs: []interface{}{mock.Anything,
				[]types.BatchKey{{
					Number: 1,
					Hash:   txHash,
				}},
				mock.Anything,
			},
			storeUnresolvedBatchKeysReturns: []interface{}{nil},
			commitReturns:                   []interface{}{errors.New("error")},
			getSequenceBatchArgs:            []interface{}{event.NumBatch},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(event.NumBatch),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
			getTxArgs:    []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns: []interface{}{tx, true, nil},
		})
	})*/
}

func TestBatchSyncronizer_HandleReorgs(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		getLastProcessedBlockReturns []interface{}
		commitReturns                []interface{}
		reorg                        BlockReorg
	}

	testFn := func(config testConfig) {
		dbMock := mocks.NewDB(t)
		txMock := mocks.NewTx(t)

		dbMock.On("GetLastProcessedBlock", mock.Anything, L1SyncTask).Return(config.getLastProcessedBlockReturns...).Once()
		if config.commitReturns != nil {
			dbMock.On("BeginStateTransaction", mock.Anything).Return(txMock, nil).Once()
			dbMock.On("StoreLastProcessedBlock", mock.Anything, L1SyncTask, mock.Anything, txMock).Return(nil).Once()
			txMock.On("Commit").Return(config.commitReturns...).Once()
		}

		reorgChan := make(chan BlockReorg)
		batchSynchronizer := &BatchSynchronizer{
			db:     dbMock,
			stop:   make(chan struct{}),
			reorgs: reorgChan,
		}

		go batchSynchronizer.handleReorgs()

		reorgChan <- config.reorg

		batchSynchronizer.stop <- struct{}{}

		dbMock.AssertExpectations(t)
		txMock.AssertExpectations(t)
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
			getLastProcessedBlockReturns: []interface{}{uint64(15), nil},
			commitReturns:                []interface{}{errors.New("error")},
			reorg: BlockReorg{
				Number: 10,
			},
		})
	})

	t.Run("Reorg block lower than what we have in db, store the block in db", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getLastProcessedBlockReturns: []interface{}{uint64(25), nil},
			commitReturns:                []interface{}{nil},
			reorg: BlockReorg{
				Number: 15,
			},
		})
	})
}
