package datacom

import (
	"context"
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
		storeOffChainDataReturns []interface{}
		sender                   *ecdsa.PrivateKey
		signer                   *ecdsa.PrivateKey
		expectedError            string
	}

	sequence := types.Sequence{
		types.ArgBytes([]byte{0, 1}),
		types.ArgBytes([]byte{2, 3}),
	}

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	otherPrivateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	testFn := func(t *testing.T, cfg testConfig) {
		t.Helper()

		var (
			signer         = privateKey
			signedSequence *types.SignedSequence
			err            error
		)

		dbMock := mocks.NewDB(t)

		if len(cfg.storeOffChainDataReturns) > 0 {
			dbMock.On("StoreOffChainData", mock.Anything, sequence.OffChainData()).Return(
				cfg.storeOffChainDataReturns...).Once()
		}

		ethermanMock := mocks.NewEtherman(t)

		ethermanMock.On("TrustedSequencer", mock.Anything).Return(crypto.PubkeyToAddress(otherPrivateKey.PublicKey), nil).Once()
		ethermanMock.On("TrustedSequencerURL", mock.Anything).Return("http://some-url", nil).Once()

		sqr := sequencer.NewTracker(config.L1Config{
			Timeout:     cfgTypes.Duration{Duration: time.Minute},
			RetryPeriod: cfgTypes.Duration{Duration: time.Second},
		}, ethermanMock)

		sqr.Start(context.Background())

		if cfg.sender != nil {
			signature, err := sequence.Sign(cfg.sender)
			require.NoError(t, err)
			signedSequence = &types.SignedSequence{
				Sequence:  sequence,
				Signature: signature,
			}
		} else {
			signedSequence = &types.SignedSequence{
				Sequence:  sequence,
				Signature: []byte{},
			}
		}

		if cfg.signer != nil {
			signer = cfg.signer
		}

		dce := NewEndpoints(dbMock, signer, sqr)

		sig, err := dce.SignSequence(*signedSequence)
		if cfg.expectedError != "" {
			require.ErrorContains(t, err, cfg.expectedError)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, sig)
		}

		sqr.Stop()

		dbMock.AssertExpectations(t)
		ethermanMock.AssertExpectations(t)
	}

	t.Run("Failed to verify sender", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			expectedError: "failed to verify sender",
		})
	})

	t.Run("Unauthorized sender", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:        privateKey,
			expectedError: "unauthorized",
		})
	})

	t.Run("Unauthorized sender", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:        privateKey,
			expectedError: "unauthorized",
		})
	})

	t.Run("Fail to store off chain data", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:                   otherPrivateKey,
			expectedError:            "failed to store offchain data",
			storeOffChainDataReturns: []interface{}{errors.New("error")},
		})
	})

	t.Run("Fail to sign sequence", func(t *testing.T) {
		t.Parallel()

		key, err := crypto.GenerateKey()
		require.NoError(t, err)

		key.D = common.Big0 // alter the key so that signing does not pass

		testFn(t, testConfig{
			sender:                   otherPrivateKey,
			signer:                   key,
			storeOffChainDataReturns: []interface{}{nil},
			expectedError:            "failed to sign",
		})
	})

	t.Run("Happy path - sequence signed", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:                   otherPrivateKey,
			storeOffChainDataReturns: []interface{}{nil},
		})
	})
}
