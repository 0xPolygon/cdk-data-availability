package sequencer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/polygonvalidium"
	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/sequencer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_NewTracker(t *testing.T) {
	testErr := errors.New("test error")

	testTable := []struct {
		name     string
		initMock func(t *testing.T) *mocks.Etherman
		err      error
	}{
		{
			name: "successfully created tracker",
			initMock: func(t *testing.T) *mocks.Etherman {
				em := mocks.NewEtherman(t)

				em.On("TrustedSequencer").Return(common.Address{}, nil)
				em.On("TrustedSequencerURL").Return("127.0.0.1", nil)

				return em
			},
		},
		{
			name: "TrustedSequencer returns error",
			initMock: func(t *testing.T) *mocks.Etherman {
				em := mocks.NewEtherman(t)

				em.On("TrustedSequencer").Return(common.Address{}, testErr)

				return em
			},
			err: testErr,
		},
		{
			name: "TrustedSequencerURL returns error",
			initMock: func(t *testing.T) *mocks.Etherman {
				em := mocks.NewEtherman(t)

				em.On("TrustedSequencer").Return(common.Address{}, nil)
				em.On("TrustedSequencerURL").Return("", testErr)

				return em
			},
			err: testErr,
		},
	}

	for _, tt := range testTable {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			em := tt.initMock(t)
			defer em.AssertExpectations(t)

			_, err := sequencer.NewTracker(config.L1Config{
				Timeout:     types.NewDuration(time.Second * 10),
				RetryPeriod: types.NewDuration(time.Millisecond),
			}, em)
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTracker(t *testing.T) {
	var (
		addressesChan chan *polygonvalidium.PolygonvalidiumSetTrustedSequencer
		urlsChan      chan *polygonvalidium.PolygonvalidiumSetTrustedSequencerURL
	)

	ctx := context.Background()

	etherman := mocks.NewEtherman(t)
	defer etherman.AssertExpectations(t)

	etherman.On("TrustedSequencer").Return(common.Address{}, nil)
	etherman.On("TrustedSequencerURL").Return("127.0.0.1:8585", nil)

	addressesSubscription := mocks.NewSubscription(t)
	defer addressesSubscription.AssertExpectations(t)

	addressesSubscription.On("Err").Return(make(<-chan error))
	addressesSubscription.On("Unsubscribe").Return()

	etherman.On("WatchSetTrustedSequencer", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			var ok bool
			addressesChan, ok = args[1].(chan *polygonvalidium.PolygonvalidiumSetTrustedSequencer)
			require.True(t, ok)
		}).
		Return(addressesSubscription, nil)

	urlsSubscription := mocks.NewSubscription(t)
	defer urlsSubscription.AssertExpectations(t)

	urlsSubscription.On("Err").Return(make(<-chan error))
	urlsSubscription.On("Unsubscribe").Return()

	etherman.On("WatchSetTrustedSequencerURL", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			var ok bool
			urlsChan, ok = args[1].(chan *polygonvalidium.PolygonvalidiumSetTrustedSequencerURL)
			require.True(t, ok)
		}).
		Return(urlsSubscription, nil)

	tracker, err := sequencer.NewTracker(config.L1Config{
		Timeout:     types.NewDuration(time.Second * 10),
		RetryPeriod: types.NewDuration(time.Millisecond),
	}, etherman)
	require.NoError(t, err)

	tracker.Start(ctx)

	var (
		updatedAddress = common.BytesToAddress([]byte("updated"))
		updatedURL     = "127.0.0.1:9585"
	)

	eventually(t, 10, func() bool {
		return addressesChan != nil && urlsChan != nil
	})

	addressesChan <- &polygonvalidium.PolygonvalidiumSetTrustedSequencer{
		NewTrustedSequencer: updatedAddress,
	}

	urlsChan <- &polygonvalidium.PolygonvalidiumSetTrustedSequencerURL{
		NewTrustedSequencerURL: updatedURL,
	}

	tracker.Stop()

	// Wait for values to be updated
	eventually(t, 10, func() bool {
		return tracker.GetAddr() == updatedAddress && tracker.GetUrl() == updatedURL
	})
}

func eventually(t *testing.T, num int, f func() bool) {
	t.Helper()

	for i := 0; i < num; i++ {
		if f() {
			return
		}

		time.Sleep(time.Second)
	}

	t.Failed()
}
