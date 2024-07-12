package etherman

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygondatacommittee"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
)

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

// Etherman defines functions that should be implemented by Etherman
type Etherman interface {
	GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)

	GetCurrentDataCommittee() (*DataCommittee, error)
	GetCurrentDataCommitteeMembers() ([]DataCommitteeMember, error)
	TrustedSequencer(ctx context.Context) (common.Address, error)
	WatchSetTrustedSequencer(
		ctx context.Context,
		events chan *polygonvalidiumetrog.PolygonvalidiumetrogSetTrustedSequencer,
	) (event.Subscription, error)
	TrustedSequencerURL(ctx context.Context) (string, error)
	WatchSetTrustedSequencerURL(
		ctx context.Context,
		events chan *polygonvalidiumetrog.PolygonvalidiumetrogSetTrustedSequencerURL,
	) (event.Subscription, error)
	FilterSequenceBatches(
		opts *bind.FilterOpts,
		numBatch []uint64,
	) (*polygonvalidiumetrog.PolygonvalidiumetrogSequenceBatchesIterator, error)
}

// etherman is the implementation of EtherMan.
type etherman struct {
	EthClient     *ethclient.Client
	CDKValidium   *polygonvalidiumetrog.Polygonvalidiumetrog
	DataCommittee *polygondatacommittee.Polygondatacommittee
}

// New creates a new etherman
func New(ctx context.Context, cfg config.L1Config) (Etherman, error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout.Duration)
	defer cancel()

	ethClient, err := ethclient.DialContext(ctx, cfg.RpcURL)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", cfg.RpcURL, err)
		return nil, err
	}

	cdkValidium, err := polygonvalidiumetrog.NewPolygonvalidiumetrog(
		common.HexToAddress(cfg.PolygonValidiumAddress),
		ethClient,
	)
	if err != nil {
		return nil, err
	}

	dataCommittee, err := polygondatacommittee.NewPolygondatacommittee(
		common.HexToAddress(cfg.DataCommitteeAddress),
		ethClient,
	)
	if err != nil {
		return nil, err
	}

	return &etherman{
		EthClient:     ethClient,
		CDKValidium:   cdkValidium,
		DataCommittee: dataCommittee,
	}, nil
}

// GetTx function get ethereum tx
func (e *etherman) GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
	return e.EthClient.TransactionByHash(ctx, txHash)
}

// HeaderByNumber returns header by number from the eth client
func (e *etherman) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return e.EthClient.HeaderByNumber(ctx, number)
}

// BlockByNumber returns a block by the given number
func (e *etherman) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return e.EthClient.BlockByNumber(ctx, number)
}

// CodeAt returns the contract code of the given account.
func (e *etherman) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	return e.EthClient.CodeAt(ctx, account, blockNumber)
}

// TrustedSequencer gets trusted sequencer address
func (e *etherman) TrustedSequencer(ctx context.Context) (common.Address, error) {
	return e.CDKValidium.TrustedSequencer(&bind.CallOpts{
		Context: ctx,
		Pending: false,
	})
}

// WatchSetTrustedSequencer watches trusted sequencer address
func (e *etherman) WatchSetTrustedSequencer(
	ctx context.Context,
	events chan *polygonvalidiumetrog.PolygonvalidiumetrogSetTrustedSequencer,
) (event.Subscription, error) {
	return e.CDKValidium.WatchSetTrustedSequencer(&bind.WatchOpts{Context: ctx}, events)
}

// TrustedSequencerURL gets trusted sequencer's RPC url
func (e *etherman) TrustedSequencerURL(ctx context.Context) (string, error) {
	return e.CDKValidium.TrustedSequencerURL(&bind.CallOpts{
		Context: ctx,
		Pending: false,
	})
}

// WatchSetTrustedSequencerURL watches trusted sequencer's RPC url
func (e *etherman) WatchSetTrustedSequencerURL(
	ctx context.Context,
	events chan *polygonvalidiumetrog.PolygonvalidiumetrogSetTrustedSequencerURL,
) (event.Subscription, error) {
	return e.CDKValidium.WatchSetTrustedSequencerURL(&bind.WatchOpts{Context: ctx}, events)
}

// FilterSequenceBatches retrieves filtered batches on CDK validium
func (e *etherman) FilterSequenceBatches(opts *bind.FilterOpts,
	numBatch []uint64) (*polygonvalidiumetrog.PolygonvalidiumetrogSequenceBatchesIterator, error) {
	return e.CDKValidium.FilterSequenceBatches(opts, numBatch)
}

// GetCurrentDataCommittee return the currently registered data committee
func (e *etherman) GetCurrentDataCommittee() (*DataCommittee, error) {
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
func (e *etherman) GetCurrentDataCommitteeMembers() ([]DataCommitteeMember, error) {
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
