package synchronizer

import (
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_InitStartBlock(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		// store mocks
		getLastProcessedBlockArgs      []interface{}
		getLastProcessedBlockReturns   []interface{}
		beginStateTransactionArgs      []interface{}
		beginStateTransactionReturns   []interface{}
		storeLastProcessedBlockArgs    []interface{}
		storeLastProcessedBlockReturns []interface{}
		commitReturns                  []interface{}
		// eth client mocks
		blockByNumberArgs    []interface{}
		blockByNumberReturns []interface{}
		codeAtArgs           [][]interface{}
		codeAtReturns        [][]interface{}

		isErrorExpected bool
	}

	l1Config := config.L1Config{
		RpcURL: "ws://localhost:8080/ws",
		// RpcURL:                 "http://localhost:8081",
		PolygonValidiumAddress: "0xCDKValidium",
		DataCommitteeAddress:   "0xDAC",
		Timeout:                types.NewDuration(time.Minute),
		RetryPeriod:            types.NewDuration(time.Second * 2),
		BlockBatchSize:         10,
	}

	testFn := func(t *testing.T, config testConfig) {
		dbMock := mocks.NewDB(t)
		txMock := mocks.NewTx(t)
		emMock := mocks.NewEtherman(t)

		if config.getLastProcessedBlockArgs != nil && config.getLastProcessedBlockReturns != nil {
			dbMock.On("GetLastProcessedBlock", config.getLastProcessedBlockArgs...).Return(
				config.getLastProcessedBlockReturns...).Once()
		}

		if config.storeLastProcessedBlockArgs != nil && config.storeLastProcessedBlockReturns != nil {
			dbMock.On("StoreLastProcessedBlock", config.storeLastProcessedBlockArgs...).Return(
				config.storeLastProcessedBlockReturns...).Once()
		}

		if config.commitReturns != nil {
			txMock.On("Commit", mock.Anything).Return(config.commitReturns...).Once()
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

		if config.blockByNumberArgs != nil && config.blockByNumberReturns != nil {
			emMock.On("HeaderByNumber", config.blockByNumberArgs...).Return(
				config.blockByNumberReturns...).Once()
		}

		if config.codeAtArgs != nil && config.codeAtReturns != nil {
			for i, args := range config.codeAtArgs {
				emMock.On("CodeAt", args...).Return(
					config.codeAtReturns[i]...).Once()
			}
		}

		err := InitStartBlock(
			dbMock,
			emMock,
			l1Config.GenesisBlock,
			common.HexToAddress(l1Config.PolygonValidiumAddress),
		)
		if config.isErrorExpected {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		dbMock.AssertExpectations(t)
		txMock.AssertExpectations(t)
	}

	t.Run("GetLastProcessedBlock returns an error", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getLastProcessedBlockArgs:    []interface{}{mock.Anything, L1SyncTask},
			getLastProcessedBlockReturns: []interface{}{uint64(1), errors.New("can't get last processed block")},
			isErrorExpected:              true,
		})
	})

	t.Run("no need to resolve start block", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getLastProcessedBlockArgs:    []interface{}{mock.Anything, L1SyncTask},
			getLastProcessedBlockReturns: []interface{}{uint64(10), nil},
			isErrorExpected:              false,
		})
	})

	t.Run("can not get block from eth client", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getLastProcessedBlockArgs:    []interface{}{mock.Anything, L1SyncTask},
			getLastProcessedBlockReturns: []interface{}{uint64(0), nil},
			blockByNumberArgs:            []interface{}{mock.Anything, mock.Anything},
			blockByNumberReturns:         []interface{}{nil, errors.New("error")},
			isErrorExpected:              true,
		})
	})

	t.Run("BeginStateTransaction fails", func(t *testing.T) {
		t.Parallel()

		block := ethTypes.NewBlockWithHeader(&ethTypes.Header{
			Number: big.NewInt(0),
		})

		testFn(t, testConfig{
			getLastProcessedBlockArgs:    []interface{}{mock.Anything, L1SyncTask},
			getLastProcessedBlockReturns: []interface{}{uint64(0), nil},
			beginStateTransactionArgs:    []interface{}{mock.Anything},
			beginStateTransactionReturns: []interface{}{nil, errors.New("error")},
			blockByNumberArgs:            []interface{}{mock.Anything, mock.Anything},
			blockByNumberReturns:         []interface{}{block, nil},
			isErrorExpected:              true,
		})
	})

	t.Run("Store off-chain data fails", func(t *testing.T) {
		t.Parallel()

		block := ethTypes.NewBlockWithHeader(&ethTypes.Header{
			Number: big.NewInt(0),
		})

		testFn(t, testConfig{
			getLastProcessedBlockArgs:      []interface{}{mock.Anything, L1SyncTask},
			getLastProcessedBlockReturns:   []interface{}{uint64(0), nil},
			beginStateTransactionArgs:      []interface{}{mock.Anything},
			storeLastProcessedBlockArgs:    []interface{}{mock.Anything, L1SyncTask, uint64(0), mock.Anything},
			storeLastProcessedBlockReturns: []interface{}{errors.New("error")},
			blockByNumberArgs:              []interface{}{mock.Anything, mock.Anything},
			blockByNumberReturns:           []interface{}{block, nil},
			isErrorExpected:                true,
		})
	})

	t.Run("Commit fails", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			blockByNumberArgs: []interface{}{mock.Anything, mock.Anything},
			blockByNumberReturns: []interface{}{ethTypes.NewBlockWithHeader(&ethTypes.Header{
				Number: big.NewInt(0),
			}), nil},
			getLastProcessedBlockArgs:      []interface{}{mock.Anything, L1SyncTask},
			getLastProcessedBlockReturns:   []interface{}{uint64(0), nil},
			storeLastProcessedBlockArgs:    []interface{}{mock.Anything, L1SyncTask, uint64(0), mock.Anything},
			storeLastProcessedBlockReturns: []interface{}{nil},
			beginStateTransactionArgs:      []interface{}{mock.Anything},
			commitReturns:                  []interface{}{errors.New("error")},

			isErrorExpected: true,
		})
	})

	t.Run("Successful init", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			getLastProcessedBlockArgs:      []interface{}{mock.Anything, L1SyncTask},
			getLastProcessedBlockReturns:   []interface{}{uint64(0), nil},
			beginStateTransactionArgs:      []interface{}{mock.Anything},
			storeLastProcessedBlockArgs:    []interface{}{mock.Anything, L1SyncTask, uint64(2), mock.Anything},
			storeLastProcessedBlockReturns: []interface{}{nil},
			commitReturns:                  []interface{}{nil},
			blockByNumberArgs:              []interface{}{mock.Anything, mock.Anything},
			blockByNumberReturns: []interface{}{ethTypes.NewBlockWithHeader(&ethTypes.Header{
				Number: big.NewInt(3),
			}), nil},
			codeAtArgs: [][]interface{}{
				{mock.Anything, common.HexToAddress(l1Config.PolygonValidiumAddress), big.NewInt(1)},
				{mock.Anything, common.HexToAddress(l1Config.PolygonValidiumAddress), big.NewInt(2)},
			},
			codeAtReturns: [][]interface{}{
				{nil, errors.New("error")},
				{[]byte{1, 2, 3, 4, 5}, nil},
			},
			isErrorExpected: false,
		})
	})
}
