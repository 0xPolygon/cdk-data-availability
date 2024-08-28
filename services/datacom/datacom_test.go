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
		sequence                 types.Sequence
		expectedError            string
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
			dbMock.On("StoreOffChainData", mock.Anything, cfg.sequence.OffChainData()).Return(
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
			signature, err := cfg.sequence.Sign(cfg.sender)
			require.NoError(t, err)
			signedSequence = &types.SignedSequence{
				Sequence:  cfg.sequence,
				Signature: signature,
			}
		} else {
			signedSequence = &types.SignedSequence{
				Sequence:  cfg.sequence,
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
			sequence: types.Sequence{
				types.ArgBytes{0, 1},
				types.ArgBytes{2, 3},
			},
		})
	})

	t.Run("Unauthorized sender", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:        privateKey,
			expectedError: "unauthorized",
			sequence: types.Sequence{
				types.ArgBytes{0, 1},
				types.ArgBytes{2, 3},
			},
		})
	})

	t.Run("Fail to store off chain data", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:                   otherPrivateKey,
			expectedError:            "failed to store offchain data",
			storeOffChainDataReturns: []interface{}{errors.New("error")},
			sequence: types.Sequence{
				types.ArgBytes{0, 1},
				types.ArgBytes{2, 3},
			},
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
			sequence: types.Sequence{
				types.ArgBytes{0, 1},
				types.ArgBytes{2, 3},
			},
		})
	})

	t.Run("Happy path - sequence signed", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:                   otherPrivateKey,
			storeOffChainDataReturns: []interface{}{nil},
			sequence: types.Sequence{
				types.ArgBytes{0, 1},
				types.ArgBytes{2, 3},
			},
		})
	})
}

func TestDataCom_SignSequenceBanana(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		storeOffChainDataReturns []interface{}
		sender                   *ecdsa.PrivateKey
		signer                   *ecdsa.PrivateKey
		sequence                 types.SequenceBanana
		expectedError            string
	}

	sequenceSignerKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	trustedSequencerKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	unknownKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	testFn := func(t *testing.T, cfg testConfig) {
		t.Helper()

		var (
			signer         = sequenceSignerKey
			signedSequence *types.SignedSequenceBanana
			err            error
		)

		dbMock := mocks.NewDB(t)

		if len(cfg.storeOffChainDataReturns) > 0 {
			dbMock.On("StoreOffChainData", mock.Anything, cfg.sequence.OffChainData()).Return(
				cfg.storeOffChainDataReturns...).Once()
		}

		ethermanMock := mocks.NewEtherman(t)

		ethermanMock.On("TrustedSequencer", mock.Anything).Return(crypto.PubkeyToAddress(trustedSequencerKey.PublicKey), nil).Once()
		ethermanMock.On("TrustedSequencerURL", mock.Anything).Return("http://some-url", nil).Once()

		sqr := sequencer.NewTracker(config.L1Config{
			Timeout:     cfgTypes.Duration{Duration: time.Minute},
			RetryPeriod: cfgTypes.Duration{Duration: time.Second},
		}, ethermanMock)

		sqr.Start(context.Background())

		if cfg.sender != nil {
			signature, err := cfg.sequence.Sign(cfg.sender)
			require.NoError(t, err)
			signedSequence = &types.SignedSequenceBanana{
				Sequence:  cfg.sequence,
				Signature: signature,
			}
		} else {
			signedSequence = &types.SignedSequenceBanana{
				Sequence:  cfg.sequence,
				Signature: []byte{},
			}
		}

		if cfg.signer != nil {
			signer = cfg.signer
		}

		dce := NewEndpoints(dbMock, signer, sqr)

		sig, err := dce.SignSequenceBanana(*signedSequence)
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
			sequence:      types.SequenceBanana{},
		})
	})

	t.Run("Unauthorized sender", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:        sequenceSignerKey,
			expectedError: "unauthorized",
			sequence:      types.SequenceBanana{},
			signer:        unknownKey,
		})
	})

	t.Run("Fail to store off chain data", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:                   trustedSequencerKey,
			expectedError:            "failed to store offchain data",
			storeOffChainDataReturns: []interface{}{errors.New("error")},
			sequence:                 types.SequenceBanana{},
		})
	})

	t.Run("Fail to sign sequence", func(t *testing.T) {
		t.Parallel()

		key, err := crypto.GenerateKey()
		require.NoError(t, err)

		key.D = common.Big0 // alter the key so that signing does not pass

		testFn(t, testConfig{
			sender:                   trustedSequencerKey,
			signer:                   key,
			storeOffChainDataReturns: []interface{}{nil},
			expectedError:            "failed to sign",
			sequence:                 types.SequenceBanana{},
		})
	})

	t.Run("Happy path - sequence signed", func(t *testing.T) {
		t.Parallel()

		testFn(t, testConfig{
			sender:                   trustedSequencerKey,
			storeOffChainDataReturns: []interface{}{nil},
			sequence:                 types.SequenceBanana{},
		})
	})
}
