package etherman

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkdatacommittee"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkvalidium"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ethereumClient interface {
	ethereum.ChainReader
	ethereum.ChainStateReader
	ethereum.ContractCaller
	ethereum.GasEstimator
	ethereum.GasPricer
	ethereum.LogFilterer
	ethereum.TransactionReader
	ethereum.TransactionSender

	bind.DeployBackend
}

// Etherman is the implementation of EtherMan.
type Etherman struct {
	EthClient     ethereumClient
	CDKValidium   *cdkvalidium.Cdkvalidium
	DataCommittee *cdkdatacommittee.Cdkdatacommittee
}

// New creaters a enw etherman
func New(cfg config.L1Config) (*Etherman, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.Duration)
	defer cancel()
	ethClient, err := ethclient.DialContext(ctx, cfg.WsURL)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", cfg.WsURL, err)
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
	return &Etherman{
		EthClient:     ethClient,
		CDKValidium:   cdkValidium,
		DataCommittee: dataCommittee,
	}, nil
}

// GetTx function get ethereum tx
func (e *Etherman) GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
	return e.EthClient.TransactionByHash(ctx, txHash)
}

// TrustedSequencer gets trusted sequencer address
func (e *Etherman) TrustedSequencer() (common.Address, error) {
	return e.CDKValidium.TrustedSequencer(&bind.CallOpts{Pending: false})
}

// TrustedSequencerURL gets trusted sequencer's RPC url
func (e *Etherman) TrustedSequencerURL() (string, error) {
	return e.CDKValidium.TrustedSequencerURL(&bind.CallOpts{Pending: false})
}

// DataCommitteeMember represents a member of the Data Committee
type DataCommitteeMember struct {
	Addr common.Address
	URL  string
}

// DataCommittee represents a specific committee
type DataCommittee struct {
	AddressesHash      common.Hash
	Members            []DataCommitteeMember
	RequiredSignatures uint64
}

// GetCurrentDataCommittee return the currently registered data committee
func (e *Etherman) GetCurrentDataCommittee() (*DataCommittee, error) {
	addrsHash, err := e.DataCommittee.CommitteeHash(&bind.CallOpts{Pending: false})
	if err != nil {
		return nil, fmt.Errorf("error getting CommitteeHash from L1 SC: %w", err)
	}
	reqSign, err := e.DataCommittee.RequiredAmountOfSignatures(&bind.CallOpts{Pending: false})
	if err != nil {
		return nil, fmt.Errorf("error getting RequiredAmountOfSignatures from L1 SC: %w", err)
	}
	members, err := e.GetCurrentDataCommitteeMembers()
	if err != nil {
		return nil, err
	}

	return &DataCommittee{
		AddressesHash:      common.Hash(addrsHash),
		RequiredSignatures: reqSign.Uint64(),
		Members:            members,
	}, nil
}

// GetCurrentDataCommitteeMembers return the currently registered data committee members
func (e *Etherman) GetCurrentDataCommitteeMembers() ([]DataCommitteeMember, error) {
	members := []DataCommitteeMember{}
	nMembers, err := e.DataCommittee.GetAmountOfMembers(&bind.CallOpts{Pending: false})
	if err != nil {
		return nil, fmt.Errorf("error getting GetAmountOfMembers from L1 SC: %w", err)
	}
	for i := int64(0); i < nMembers.Int64(); i++ {
		member, err := e.DataCommittee.Members(&bind.CallOpts{Pending: false}, big.NewInt(i))
		if err != nil {
			return nil, fmt.Errorf("error getting Members %d from L1 SC: %w", i, err)
		}
		members = append(members, DataCommitteeMember{
			Addr: member.Addr,
			URL:  member.Url,
		})
	}
	return members, nil
}
