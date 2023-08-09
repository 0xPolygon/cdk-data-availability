package synchronizer

import (
	"context"
	"time"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/blocktracker"
	"github.com/umbracle/ethgo/jsonrpc"
)

// ReorgDetector watches for block reorganizations on chain, and sends messages to subscribing components when a reorg
// is detected.
type ReorgDetector struct {
	rpcUrl        string
	pollingPeriod time.Duration
	subscribers   []chan ReorgBlock
	cancel        context.CancelFunc
}

// ReorgBlock is emitted to subscribers when a reorg is detected. Number is the block to which the chain rewound.
type ReorgBlock struct {
	Number uint64
}

// NewReorgDetector creates a new ReorgDetector
func NewReorgDetector(rpcUrl string, pollingPeriod time.Duration) (*ReorgDetector, error) {
	return &ReorgDetector{
		rpcUrl:        rpcUrl,
		pollingPeriod: pollingPeriod,
	}, nil
}

// Subscribe returns a channel on which the caller can receive reorg messages
func (rd *ReorgDetector) Subscribe() <-chan ReorgBlock {
	ch := make(chan ReorgBlock)
	rd.subscribers = append(rd.subscribers, ch)
	return ch
}

// Start starts the ReorgDetector tracking for reorg events
func (rd *ReorgDetector) Start() error {

	ctx, cancel := context.WithCancel(context.Background())
	rd.cancel = cancel

	blocks := make(chan *ethgo.Block)
	err := rd.trackBlocks(ctx, blocks)
	if err != nil {
		return err
	}

	go func() {
		var lastBlock *ethgo.Block
		for {
			select {
			case block := <-blocks:
				if lastBlock != nil {
					if lastBlock.Number+1 >= block.Number {
						lca := ReorgBlock{Number: block.Number}
						for _, ch := range rd.subscribers {
							ch <- lca
						}
					}
				}
				lastBlock = block
			case <-ctx.Done():
				close(blocks)
				return
			}
		}
	}()

	return nil
}

func (rd *ReorgDetector) Stop() {
	if rd.cancel == nil {
		return
	}
	rd.cancel()
}

func (rd *ReorgDetector) trackBlocks(ctx context.Context, ch chan *ethgo.Block) error {
	client, err := jsonrpc.NewClient(rd.rpcUrl)
	if err != nil {
		return err
	}
	tracker := blocktracker.NewJSONBlockTracker(client.Eth())
	tracker.PollInterval = rd.pollingPeriod
	go func() {
		_ = tracker.Track(ctx, func(block *ethgo.Block) error {
			ch <- block
			return nil
		})
	}()
	return nil
}
