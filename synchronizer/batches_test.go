package synchronizer

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/polygonvalidium"
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
	batchKey := batchKey{
		number: 1,
		hash:   crypto.Keccak256Hash(data),
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
				Number:      types.ArgUint64(batchKey.number),
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

func TestBatchSynchronizer_HandleEvent(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		// etherman mock
		getTxArgs    []interface{}
		getTxReturns []interface{}
		// db mock
		existsArgs                   []interface{}
		existsReturns                []interface{}
		beginStateTransactionArgs    []interface{}
		beginStateTransactionReturns []interface{}
		storeOffChainDataArgs        []interface{}
		storeOffChainDataReturns     []interface{}
		commitReturns                []interface{}
		rollbackArgs                 []interface{}
		// sequencer mocks
		getSequenceBatchArgs    []interface{}
		getSequenceBatchReturns []interface{}

		isErrorExpected bool
	}

	to := common.HexToAddress("0xFFFF")
	event := &polygonvalidium.PolygonvalidiumSequenceBatches{
		Raw: ethTypes.Log{
			TxHash: common.BytesToHash([]byte{0, 1, 2, 3}),
		},
		NumBatch: 10,
	}
	batchL2Data := []byte{1, 2, 3, 4, 5, 6}
	txHash := crypto.Keccak256Hash(batchL2Data)

	batchData := []polygonvalidium.PolygonValidiumEtrogValidiumBatchData{
		{
			TransactionsHash: txHash,
		},
	}

	a, err := abi.JSON(strings.NewReader(polygonvalidium.PolygonvalidiumABI))
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

	testFn := func(config testConfig) {
		dbMock := mocks.NewDB(t)
		txMock := mocks.NewTx(t)
		ethermanMock := mocks.NewEtherman(t)
		sequencerMock := mocks.NewSequencerTracker(t)

		if config.getTxArgs != nil && config.getTxReturns != nil {
			ethermanMock.On("GetTx", config.getTxArgs...).Return(
				config.getTxReturns...).Once()
		}

		if config.existsArgs != nil && config.existsReturns != nil {
			dbMock.On("Exists", config.existsArgs...).Return(
				config.existsReturns...).Once()
		}

		if config.getSequenceBatchArgs != nil && config.getSequenceBatchReturns != nil {
			sequencerMock.On("GetSequenceBatch", config.getSequenceBatchArgs...).Return(
				config.getSequenceBatchReturns...).Once()
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

		if config.storeOffChainDataArgs != nil && config.storeOffChainDataReturns != nil {
			dbMock.On("StoreOffChainData", config.storeOffChainDataArgs...).Return(
				config.storeOffChainDataReturns...).Once()
		}

		if config.commitReturns != nil {
			txMock.On("Commit", mock.Anything).Return(
				config.commitReturns...).Once()
		}

		if config.rollbackArgs != nil {
			txMock.On("Rollback", config.rollbackArgs...).Return(nil).Once()
		}

		batchSynronizer := &BatchSynchronizer{
			db:        dbMock,
			client:    ethermanMock,
			sequencer: sequencerMock,
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
		sequencerMock.AssertExpectations(t)
	}

	t.Run("Could not get tx data", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getTxArgs:       []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:    []interface{}{nil, false, errors.New("error")},
			isErrorExpected: true,
		})
	})

	t.Run("Invalid tx data", func(t *testing.T) {
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

	t.Run("has batch in storage", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getTxArgs:       []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:    []interface{}{tx, true, nil},
			existsArgs:      []interface{}{mock.Anything, common.Hash(batchData[0].TransactionsHash)},
			existsReturns:   []interface{}{true},
			isErrorExpected: false,
		})
	})

	t.Run("doesn't have batch in storage - successfully stored", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			getTxArgs:            []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:         []interface{}{tx, true, nil},
			existsArgs:           []interface{}{mock.Anything, txHash},
			existsReturns:        []interface{}{false},
			getSequenceBatchArgs: []interface{}{event.NumBatch},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(event.NumBatch),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
			beginStateTransactionArgs: []interface{}{mock.Anything},
			storeOffChainDataArgs: []interface{}{mock.Anything,
				[]types.OffChainData{{
					Key:   txHash,
					Value: batchL2Data,
				}},
				mock.Anything,
			},
			storeOffChainDataReturns: []interface{}{nil},
			commitReturns:            []interface{}{nil},
			isErrorExpected:          false,
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
			existsArgs:                   []interface{}{mock.Anything, txHash},
			existsReturns:                []interface{}{false},
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
			storeOffChainDataArgs: []interface{}{mock.Anything,
				[]types.OffChainData{{
					Key:   txHash,
					Value: batchL2Data,
				}},
				mock.Anything,
			},
			storeOffChainDataReturns:  []interface{}{errors.New("error")},
			beginStateTransactionArgs: []interface{}{mock.Anything},
			rollbackArgs:              []interface{}{mock.Anything},
			getTxArgs:                 []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:              []interface{}{tx, true, nil},
			existsArgs:                []interface{}{mock.Anything, txHash},
			existsReturns:             []interface{}{false},
			getSequenceBatchArgs:      []interface{}{event.NumBatch},
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
			storeOffChainDataArgs: []interface{}{mock.Anything,
				[]types.OffChainData{{
					Key:   txHash,
					Value: batchL2Data,
				}},
				mock.Anything,
			},
			storeOffChainDataReturns: []interface{}{nil},
			commitReturns:            []interface{}{errors.New("error")},
			getSequenceBatchArgs:     []interface{}{event.NumBatch},
			getSequenceBatchReturns: []interface{}{&sequencer.SeqBatch{
				Number:      types.ArgUint64(event.NumBatch),
				BatchL2Data: types.ArgBytes(batchL2Data),
			}, nil},
			getTxArgs:     []interface{}{mock.Anything, event.Raw.TxHash},
			getTxReturns:  []interface{}{tx, true, nil},
			existsArgs:    []interface{}{mock.Anything, txHash},
			existsReturns: []interface{}{false},
		})
	})
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

		dbMock.On("GetLastProcessedBlock", mock.Anything, l1SyncTask).Return(config.getLastProcessedBlockReturns...).Once()
		if config.commitReturns != nil {
			dbMock.On("BeginStateTransaction", mock.Anything).Return(txMock, nil).Once()
			dbMock.On("StoreLastProcessedBlock", mock.Anything, l1SyncTask, mock.Anything, txMock).Return(nil).Once()
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
