package etherman

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkdatacommittee"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkvalidium"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
)

// Etherman defines functions that should be implemented by Etherman
type Etherman interface {
	GetCurrentDataCommittee() (*DataCommittee, error)
	GetCurrentDataCommitteeMembers() ([]DataCommitteeMember, error)
	GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)
	TrustedSequencer() (common.Address, error)
	WatchSetTrustedSequencer(
		ctx context.Context,
		events chan *cdkvalidium.CdkvalidiumSetTrustedSequencer,
	) (event.Subscription, error)
	TrustedSequencerURL() (string, error)
	WatchSetTrustedSequencerURL(
		ctx context.Context,
		events chan *cdkvalidium.CdkvalidiumSetTrustedSequencerURL,
	) (event.Subscription, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	FilterSequenceBatches(
		opts *bind.FilterOpts,
		numBatch []uint64,
	) (*cdkvalidium.CdkvalidiumSequenceBatchesIterator, error)
}

var _ Etherman = (*EthermanImpl)(nil)

// EthermanImpl is the implementation of EtherMan.
type EthermanImpl struct {
	EthClient     *ethclient.Client
	CDKValidium   *cdkvalidium.Cdkvalidium
	DataCommittee *cdkdatacommittee.Cdkdatacommittee
}

// New creates a new etherman
func New(cfg config.L1Config) (*EthermanImpl, error) {
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

	return &EthermanImpl{
		EthClient:     ethClient,
		CDKValidium:   cdkValidium,
		DataCommittee: dataCommittee,
	}, nil
}

// GetTx function get ethereum tx
func (e *EthermanImpl) GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
	return e.EthClient.TransactionByHash(ctx, txHash)
}

// TrustedSequencer gets trusted sequencer address
func (e *EthermanImpl) TrustedSequencer() (common.Address, error) {
	return e.CDKValidium.TrustedSequencer(&bind.CallOpts{Pending: false})
}

// WatchSetTrustedSequencer watches trusted sequencer address
func (e *EthermanImpl) WatchSetTrustedSequencer(
	ctx context.Context,
	events chan *cdkvalidium.CdkvalidiumSetTrustedSequencer,
) (event.Subscription, error) {
	return e.CDKValidium.WatchSetTrustedSequencer(&bind.WatchOpts{Context: ctx}, events)
}

// TrustedSequencerURL gets trusted sequencer's RPC url
func (e *EthermanImpl) TrustedSequencerURL() (string, error) {
	return e.CDKValidium.TrustedSequencerURL(&bind.CallOpts{Pending: false})
}

// WatchSetTrustedSequencerURL watches trusted sequencer's RPC url
func (e *EthermanImpl) WatchSetTrustedSequencerURL(
	ctx context.Context,
	events chan *cdkvalidium.CdkvalidiumSetTrustedSequencerURL,
) (event.Subscription, error) {
	return e.CDKValidium.WatchSetTrustedSequencerURL(&bind.WatchOpts{Context: ctx}, events)
}

// HeaderByNumber returns header by number from the eth client
func (e *EthermanImpl) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return e.EthClient.HeaderByNumber(ctx, number)
}

// FilterSequenceBatches retrieves filtered batches on CDK validium
func (e *EthermanImpl) FilterSequenceBatches(opts *bind.FilterOpts,
	numBatch []uint64) (*cdkvalidium.CdkvalidiumSequenceBatchesIterator, error) {
	return e.CDKValidium.FilterSequenceBatches(opts, numBatch)
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
func (e *EthermanImpl) GetCurrentDataCommittee() (*DataCommittee, error) {
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
		AddressesHash:      addrsHash,
		RequiredSignatures: reqSign.Uint64(),
		Members:            members,
	}, nil
}

// GetCurrentDataCommitteeMembers return the currently registered data committee members
func (e *EthermanImpl) GetCurrentDataCommitteeMembers() ([]DataCommitteeMember, error) {
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
