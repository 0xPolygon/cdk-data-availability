package datacom

import (
	"crypto/ecdsa"
	"errors"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	cfgTypes "github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/sequencer"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDataCom_SignSequence(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		beginStateTransactionReturns []interface{}
		storeOffChainDataReturns     []interface{}
		rollbackReturns              []interface{}
		commitReturns                []interface{}
		sender                       *ecdsa.PrivateKey
		signer                       *ecdsa.PrivateKey
		expectedError                string
	}

	sequence := types.Sequence{
		types.ArgBytes([]byte{0, 1}),
		types.ArgBytes([]byte{2, 3}),
	}

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	otherPrivateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	testFn := func(cfg testConfig) {
		var (
			signer         = privateKey
			signedSequence *types.SignedSequence
			err            error
		)

		txMock := mocks.NewTx(t)
		dbMock := mocks.NewDB(t)
		if cfg.beginStateTransactionReturns != nil {
			dbMock.On("BeginStateTransaction", mock.Anything).Return(cfg.beginStateTransactionReturns...).Once()
		} else if cfg.storeOffChainDataReturns != nil {
			dbMock.On("BeginStateTransaction", mock.Anything).Return(txMock, nil).Once()
			dbMock.On("StoreOffChainData", mock.Anything, sequence.OffChainData(), txMock).Return(
				cfg.storeOffChainDataReturns...).Once()

			if cfg.rollbackReturns != nil {
				txMock.On("Rollback").Return(cfg.rollbackReturns...).Once()
			} else {
				txMock.On("Commit").Return(cfg.commitReturns...).Once()
			}
		}

		ethermanMock := mocks.NewEtherman(t)

		ethermanMock.On("TrustedSequencer").Return(crypto.PubkeyToAddress(otherPrivateKey.PublicKey), nil).Once()
		ethermanMock.On("TrustedSequencerURL").Return("http://some-url", nil).Once()

		sequencer, err := sequencer.NewTracker(config.L1Config{
			Timeout:     cfgTypes.Duration{Duration: time.Minute},
			RetryPeriod: cfgTypes.Duration{Duration: time.Second},
		}, ethermanMock)
		require.NoError(t, err)

		if cfg.sender != nil {
			signedSequence, err = sequence.Sign(cfg.sender)
			require.NoError(t, err)
		} else {
			signedSequence = &types.SignedSequence{
				Sequence:  sequence,
				Signature: []byte{},
			}
		}

		if cfg.signer != nil {
			signer = cfg.signer
		}

		dce := NewDataComEndpoints(dbMock, signer, sequencer)

		sig, err := dce.SignSequence(*signedSequence)
		if cfg.expectedError != "" {
			require.ErrorContains(t, err, cfg.expectedError)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, sig)
		}

		txMock.AssertExpectations(t)
		dbMock.AssertExpectations(t)
		ethermanMock.AssertExpectations(t)
	}

	t.Run("Failed to verify sender", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			expectedError: "failed to verify sender",
		})
	})

	t.Run("Unauthorized sender", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			sender:        privateKey,
			expectedError: "unauthorized",
		})
	})

	t.Run("Unauthorized sender", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			sender:        privateKey,
			expectedError: "unauthorized",
		})
	})

	t.Run("Fail to begin state transaction", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			sender:                       otherPrivateKey,
			expectedError:                "failed to connect to the state",
			beginStateTransactionReturns: []interface{}{nil, errors.New("error")},
		})
	})

	t.Run("Fail to store off chain data - rollback fails", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			sender:                   otherPrivateKey,
			expectedError:            "failed to rollback db transaction",
			storeOffChainDataReturns: []interface{}{errors.New("error")},
			rollbackReturns:          []interface{}{errors.New("rollback fails")},
		})
	})

	t.Run("Fail to store off chain data", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			sender:                   otherPrivateKey,
			expectedError:            "failed to store offchain data",
			storeOffChainDataReturns: []interface{}{errors.New("error")},
			rollbackReturns:          []interface{}{nil},
		})
	})

	t.Run("Fail to commit tx", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			sender:                   otherPrivateKey,
			expectedError:            "failed to commit db transaction",
			storeOffChainDataReturns: []interface{}{nil},
			commitReturns:            []interface{}{errors.New("error")},
		})
	})

	t.Run("Fail to sign sequence", func(t *testing.T) {
		t.Parallel()

		key, err := crypto.GenerateKey()
		require.NoError(t, err)

		key.D = common.Big0 // alter the key so that signing does not pass

		testFn(testConfig{
			sender:                   otherPrivateKey,
			signer:                   key,
			storeOffChainDataReturns: []interface{}{nil},
			commitReturns:            []interface{}{nil},
			expectedError:            "failed to sign",
		})
	})

	t.Run("Happy path - sequence signed", func(t *testing.T) {
		t.Parallel()

		testFn(testConfig{
			sender:                   otherPrivateKey,
			storeOffChainDataReturns: []interface{}{nil},
			commitReturns:            []interface{}{nil},
		})
	})
}
