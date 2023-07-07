package l1

import (
	"context"

	"github.com/0xPolygon/supernets2-data-availability/config"
	"github.com/0xPolygon/supernets2-node/etherman"
	"github.com/0xPolygon/supernets2-node/etherman/smartcontracts/supernets2"
	"github.com/0xPolygon/supernets2-node/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// newEtherman constructs an etherman client that only needs the free API calls to ZkEVMAddr contract
func newEtherman(cfg config.L1Config) (*etherman.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.Duration)
	defer cancel()
	ethClient, err := ethclient.DialContext(ctx, cfg.WsURL)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", cfg.WsURL, err)
		return nil, err
	}
	supernets2, err := supernets2.NewSupernets2(common.HexToAddress(cfg.ZkEVMAddress), ethClient)
	if err != nil {
		return nil, err
	}
	return &etherman.Client{
		EthClient:  ethClient,
		Supernets2: supernets2,
	}, nil
}
