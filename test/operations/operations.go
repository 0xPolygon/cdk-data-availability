package operations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	cmdFolder = "test"
	// DefaultInterval is a time interval
	DefaultInterval = 2 * time.Second
	// DefaultDeadline is a time interval
	DefaultDeadline                          = 2 * time.Minute
	DefaultL1NetworkURL                      = "http://localhost:8545"
	DefaultL2NetworkURL                      = "http://localhost:8123"
	DefaultSequencerPrivateKey               = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	DefaultL2ChainID                  uint64 = 1001
	DefaultL1ChainID                  uint64 = 1337
	DefaultL1DataCommitteeContract           = "0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6"
	DefaultTimeoutTxToBeMined                = 1 * time.Minute
	DefaultL1CDKValidiumSmartContract        = "0x0DCd1Bf9A1b36cE34237eEaFef220932846BCD82"
)

var (
	// ErrTimeoutReached is thrown when the timeout is reached and
	// because the condition is not matched
	ErrTimeoutReached = fmt.Errorf("timeout has been reached")
)

// Poll retries the given condition with the given interval until it succeeds
// or the given deadline expires.
func Poll(interval, deadline time.Duration, condition ConditionFunc) error {
	timeout := time.After(deadline)
	tick := time.NewTicker(interval)

	for {
		select {
		case <-timeout:
			return ErrTimeoutReached
		case <-tick.C:
			ok, err := condition()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

// ConditionFunc is a generic function
type ConditionFunc func() (done bool, err error)

func nodeUpCondition() (done bool, err error) {
	return NodeUpCondition(DefaultL2NetworkURL)
}

func networkUpCondition() (bool, error) {
	return NodeUpCondition(DefaultL1NetworkURL)
}

// NodeUpCondition check if the container is up and running
func NodeUpCondition(target string) (bool, error) {
	var jsonStr = []byte(`{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}`)
	req, err := http.NewRequest(
		"POST", target,
		bytes.NewBuffer(jsonStr))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		// we allow connection errors to wait for the container up
		return false, nil
	}

	if res.Body != nil {
		defer func() {
			err = res.Body.Close()
		}()
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return false, err
	}

	r := struct {
		Result bool
	}{
		Result: true,
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return false, err
	}

	done := !r.Result

	return done, nil
}

// RunMakeTarget runs a Makefile target.
func RunMakeTarget(target string) error {
	cmd := exec.Command("make", target)
	return runCmd(cmd)
}

func runCmd(c *exec.Cmd) error {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current work directory: %v", err)
	}

	if strings.Contains(dir, cmdFolder) {
		// Making the change dir to work in any nesting directory level inside cmd folder
		base := filepath.Base(dir)
		for base != cmdFolder {
			dir = filepath.Dir(dir)
			base = filepath.Base(dir)
		}
	} else {
		dir = fmt.Sprintf("../../%s", cmdFolder)
	}
	c.Dir = dir

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// StartComponent starts a docker-compose component.
func StartComponent(component string, conditions ...ConditionFunc) error {
	cmdDown := fmt.Sprintf("stop-%s", component)
	if err := RunMakeTarget(cmdDown); err != nil {
		return err
	}
	cmdUp := fmt.Sprintf("run-%s", component)
	if err := RunMakeTarget(cmdUp); err != nil {
		return err
	}

	// Wait component to be ready
	for _, condition := range conditions {
		if err := Poll(DefaultInterval, DefaultDeadline, condition); err != nil {
			return err
		}
	}
	return nil
}

// StopComponent stops a docker-compose component.
func StopComponent(component string) error {
	cmdDown := fmt.Sprintf("stop-%s", component)
	return RunMakeTarget(cmdDown)
}

func stopNode() error {
	return StopComponent("node")
}

func stopNetwork() error {
	return StopComponent("network")
}

// Teardown stops all the components.
func Teardown() error {
	err := stopNode()
	if err != nil {
		return err
	}

	err = stopNetwork()
	if err != nil {
		return err
	}

	return nil
}

// Setup creates all the required components and initializes them according to
// the manager config.
func Setup() error {
	// Run network container
	err := StartNetwork()
	if err != nil {
		return err
	}

	// Run node container
	err = StartNode()
	if err != nil {
		return err
	}

	return nil
}

// StartNetwork starts the L1 network container
func StartNetwork() error {
	return StartComponent("network", networkUpCondition)
}

// StartNode starts the node container
func StartNode() error {
	return StartComponent("node", nodeUpCondition)
}

// GetAuth configures and returns an auth object.
func GetAuth(privateKeyStr string, chainID uint64) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyStr, "0x"))
	if err != nil {
		return nil, err
	}

	return bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(0).SetUint64(chainID))
}

// WaitTxToBeMined waits until a tx has been mined or the given timeout expires.
func WaitTxToBeMined(parentCtx context.Context, client ethClienter, tx *types.Transaction, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, client, tx)
	if errors.Is(err, context.DeadlineExceeded) {
		return err
	} else if err != nil {
		log.Errorf("error waiting tx %s to be mined: %w", tx.Hash(), err)
		return err
	}
	if receipt.Status == types.ReceiptStatusFailed {
		// Get revert reason
		reason, reasonErr := RevertReason(ctx, client, tx, receipt.BlockNumber)
		if reasonErr != nil {
			reason = reasonErr.Error()
		}
		return fmt.Errorf("transaction has failed, reason: %s, receipt: %+v. tx: %+v, gas: %v", reason, receipt, tx, tx.Gas())
	}
	log.Debug("Transaction successfully mined: ", tx.Hash())
	return nil
}

// RevertReason returns the revert reason for a tx that has a receipt with failed status
func RevertReason(ctx context.Context, c ethClienter, tx *types.Transaction, blockNumber *big.Int) (string, error) {
	if tx == nil {
		return "", nil
	}

	from, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err != nil {
		signer := types.LatestSignerForChainID(tx.ChainId())
		from, err = types.Sender(signer, tx)
		if err != nil {
			return "", err
		}
	}
	msg := ethereum.CallMsg{
		From: from,
		To:   tx.To(),
		Gas:  tx.Gas(),

		Value: tx.Value(),
		Data:  tx.Data(),
	}
	hex, err := c.CallContract(ctx, msg, blockNumber)
	if err != nil {
		return "", err
	}

	unpackedMsg, err := abi.UnpackRevert(hex)
	if err != nil {
		log.Warnf("failed to get the revert message for tx %v: %v", tx.Hash(), err)
		return "", errors.New("execution reverted")
	}

	return unpackedMsg, nil
}

// ApplyL2Txs sends the given L2 txs, waits for them to be consolidated and
// checks the final state.
func ApplyL2Txs(ctx context.Context, txs []*types.Transaction, auth *bind.TransactOpts, client *ethclient.Client, confirmationLevel ConfirmationLevel) ([]*big.Int, error) {
	var err error
	if auth == nil {
		auth, err = GetAuth(DefaultSequencerPrivateKey, DefaultL2ChainID)
		if err != nil {
			return nil, err
		}
	}

	if client == nil {
		client, err = ethclient.Dial(DefaultL2NetworkURL)
		if err != nil {
			return nil, err
		}
	}
	waitToBeMined := confirmationLevel != PoolConfirmationLevel
	var initialNonce uint64
	if waitToBeMined {
		initialNonce, err = client.NonceAt(ctx, auth.From, nil)
		if err != nil {
			return nil, err
		}
	}
	sentTxs, err := applyTxs(ctx, txs, auth, client, waitToBeMined)
	if err != nil {
		return nil, err
	}
	if confirmationLevel == PoolConfirmationLevel {
		return nil, nil
	}

	l2BlockNumbers := make([]*big.Int, 0, len(sentTxs))
	for i, tx := range sentTxs {
		// check transaction nonce against transaction reported L2 block number
		receipt, err := client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return nil, err
		}

		// get L2 block number
		l2BlockNumbers = append(l2BlockNumbers, receipt.BlockNumber)
		expectedNonce := initialNonce + uint64(i)
		if tx.Nonce() != expectedNonce {
			return nil, fmt.Errorf("mismatching nonce for tx %v: want %d, got %d\n", tx.Hash(), expectedNonce, tx.Nonce())
		}
		if confirmationLevel == TrustedConfirmationLevel {
			continue
		}

		// wait for l2 block to be virtualized
		log.Infof("waiting for the block number %v to be virtualized", receipt.BlockNumber.String())
		err = WaitL2BlockToBeVirtualized(receipt.BlockNumber, 1*time.Minute) //nolint:gomnd
		if err != nil {
			// tmp
			cmd := exec.Command(
				"docker", "logs", "zkevm-node",
			)
			out, _ := cmd.CombinedOutput()
			log.Debug("zkevm node: ", string(out))
			cmd = exec.Command(
				"docker", "logs", "--tail", "1000", "cdk-data-availability-0",
			)
			out, _ = cmd.CombinedOutput()
			log.Debug("DA0: ", string(out))
			cmd = exec.Command(
				"docker", "logs", "--tail", "1000", "cdk-data-availability-1",
			)
			out, _ = cmd.CombinedOutput()
			log.Debug("DA1: ", string(out))
			cmd = exec.Command(
				"docker", "logs", "--tail", "1000", "cdk-data-availability-2",
			)
			out, _ = cmd.CombinedOutput()
			log.Debug("DA2: ", string(out))
			cmd = exec.Command(
				"docker", "logs", "--tail", "1000", "cdk-data-availability-3",
			)
			out, _ = cmd.CombinedOutput()
			log.Debug("DA3: ", string(out))
			cmd = exec.Command(
				"docker", "logs", "--tail", "1000", "cdk-data-availability-4",
			)
			out, _ = cmd.CombinedOutput()
			log.Debug("DA4: ", string(out))
			return nil, err
		}
		if confirmationLevel == VirtualConfirmationLevel {
			continue
		}

		// wait for l2 block number to be consolidated
		log.Infof("waiting for the block number %v to be consolidated", receipt.BlockNumber.String())
		err = WaitL2BlockToBeConsolidated(receipt.BlockNumber, 4*time.Minute) //nolint:gomnd
		if err != nil {
			return nil, err
		}
	}

	return l2BlockNumbers, nil
}

// WaitL2BlockToBeVirtualized waits until a L2 Block has been virtualized or the given timeout expires.
func WaitL2BlockToBeVirtualized(l2Block *big.Int, timeout time.Duration) error {
	l2NetworkURL := "http://localhost:8123"
	return Poll(DefaultInterval, timeout, func() (bool, error) {
		return l2BlockVirtualizationCondition(l2Block, l2NetworkURL)
	})
}

// l2BlockConsolidationCondition
func l2BlockConsolidationCondition(l2Block *big.Int) (bool, error) {
	l2NetworkURL := "http://localhost:8123"
	response, err := rpc.JSONRPCCall(l2NetworkURL, "zkevm_isBlockConsolidated", rpc.HexEncodeBig(l2Block))
	if err != nil {
		return false, err
	}
	if response.Error != nil {
		return false, fmt.Errorf("%d - %s", response.Error.Code, response.Error.Message)
	}
	var result bool
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return false, err
	}
	return result, nil
}

// WaitL2BlockToBeConsolidated waits until a L2 Block has been consolidated or the given timeout expires.
func WaitL2BlockToBeConsolidated(l2Block *big.Int, timeout time.Duration) error {
	return Poll(DefaultInterval, timeout, func() (bool, error) {
		return l2BlockConsolidationCondition(l2Block)
	})
}

// ConfirmationLevel type used to describe the confirmation level of a transaction
type ConfirmationLevel int

// PoolConfirmationLevel indicates that transaction is added into the pool
const PoolConfirmationLevel ConfirmationLevel = 0

// TrustedConfirmationLevel indicates that transaction is  added into the trusted state
const TrustedConfirmationLevel ConfirmationLevel = 1

// VirtualConfirmationLevel indicates that transaction is  added into the virtual state
const VirtualConfirmationLevel ConfirmationLevel = 2

// VerifiedConfirmationLevel indicates that transaction is  added into the verified state
const VerifiedConfirmationLevel ConfirmationLevel = 3

// l2BlockVirtualizationCondition
func l2BlockVirtualizationCondition(l2Block *big.Int, l2NetworkURL string) (bool, error) {
	response, err := rpc.JSONRPCCall(l2NetworkURL, "zkevm_isBlockVirtualized", rpc.HexEncodeBig(l2Block))
	if err != nil {
		return false, err
	}
	if response.Error != nil {
		return false, fmt.Errorf("%d - %s", response.Error.Code, response.Error.Message)
	}
	var result bool
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return false, err
	}
	return result, nil
}

func applyTxs(ctx context.Context, txs []*types.Transaction, auth *bind.TransactOpts, client *ethclient.Client, waitToBeMined bool) ([]*types.Transaction, error) {
	var sentTxs []*types.Transaction

	for i := 0; i < len(txs); i++ {
		signedTx, err := auth.Signer(auth.From, txs[i])
		if err != nil {
			return nil, err
		}
		log.Infof("Sending Tx %v Nonce %v", signedTx.Hash(), signedTx.Nonce())
		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			return nil, err
		}

		sentTxs = append(sentTxs, signedTx)
	}
	if !waitToBeMined {
		return nil, nil
	}

	// wait for TX to be mined
	timeout := 180 * time.Second //nolint:gomnd
	for _, tx := range sentTxs {
		log.Infof("Waiting Tx %s to be mined", tx.Hash())
		err := WaitTxToBeMined(ctx, client, tx, timeout)
		if err != nil {
			return nil, err
		}
		log.Infof("Tx %s mined successfully", tx.Hash())
	}
	nTxs := len(txs)
	if nTxs > 1 {
		log.Infof("%d transactions added into the trusted state successfully.", nTxs)
	} else {
		log.Info("transaction added into the trusted state successfully.")
	}

	return sentTxs, nil
}

type ethClienter interface {
	ethereum.TransactionReader
	ethereum.ContractCaller
	bind.DeployBackend
}
