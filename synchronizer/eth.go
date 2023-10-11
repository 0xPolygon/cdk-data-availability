package synchronizer

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkdatacommittee"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkvalidium"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func newRPCEtherman(cfg config.L1Config) (*etherman.Client, error) {
	return newEtherman(cfg, cfg.RpcURL)
}

func newWSEtherman(cfg config.L1Config) (*etherman.Client, error) {
	return newEtherman(cfg, cfg.WsURL)
}

// newEtherman constructs an etherman client that only needs the free contract calls
func newEtherman(cfg config.L1Config, url string) (*etherman.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.Duration)
	defer cancel()
	ethClient, err := ethclient.DialContext(ctx, url)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", url, err)
		return nil, err
	}
	cdkValidium, err := cdkvalidium.NewCdkvalidium(common.HexToAddress(cfg.CDKValidiumAddress), ethClient)
	if err != nil {
		return nil, err
	}
	dataCommittee, err :=
		cdkdatacommittee.NewCdkdatacommittee(common.HexToAddress(cfg.DataCommitteeAddress), ethClient)
	if err != nil {
		return nil, err
	}
	return &etherman.Client{
		EthClient:     ethClient,
		CDKValidium:   cdkValidium,
		DataCommittee: dataCommittee,
	}, nil
}

// ParseEvent unpacks the keys in a SequenceBatches event
func ParseEvent(event *cdkvalidium.CdkvalidiumSequenceBatches, txData []byte) (uint64, []common.Hash, error) {
	a, err := abi.JSON(strings.NewReader(cdkvalidium.CdkvalidiumABI))
	if err != nil {
		return 0, nil, err
	}
	method, err := a.MethodById(txData[:4])
	if err != nil {
		return 0, nil, err
	}
	data, err := method.Inputs.Unpack(txData[4:])
	if err != nil {
		return 0, nil, err
	}
	var batches []cdkvalidium.CDKValidiumBatchData
	bytes, err := json.Marshal(data[0])
	if err != nil {
		return 0, nil, err
	}
	err = json.Unmarshal(bytes, &batches)
	if err != nil {
		return 0, nil, err
	}

	var keys []common.Hash
	for _, batch := range batches {
		keys = append(keys, batch.TransactionsHash)
	}
	return event.Raw.BlockNumber, keys, nil
}
