package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	dataavailability "github.com/0xPolygon/supernets2-data-availability"
	"github.com/0xPolygon/supernets2-data-availability/config"
	"github.com/0xPolygon/supernets2-data-availability/db"
	"github.com/0xPolygon/supernets2-data-availability/jsonrpc"
	"github.com/0xPolygon/supernets2-data-availability/services/datacom"
	"github.com/0xPolygon/supernets2-data-availability/services/sync"
	"github.com/0xPolygon/supernets2-data-availability/synchronizer"
	dbConf "github.com/0xPolygon/supernets2-node/db"
	"github.com/0xPolygon/supernets2-node/log"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

const appName = "supernets2-data-availability"

var (
	configFileFlag = cli.StringFlag{
		Name:     config.FlagCfg,
		Aliases:  []string{"c"},
		Usage:    "Configuration `FILE`",
		Required: false,
	}
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = dataavailability.Version
	app.Commands = []*cli.Command{
		{
			Name:    "run",
			Aliases: []string{},
			Usage:   fmt.Sprintf("Run the %v", appName),
			Action:  start,
			Flags:   []cli.Flag{&configFileFlag},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func start(cliCtx *cli.Context) error {
	// Load config
	c, err := config.Load(cliCtx)
	if err != nil {
		panic(err)
	}

	// Configure logging
	log.Init(c.Log)

	// Prepare DB
	pg, err := dbConf.NewSQLDB(c.DB)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.RunMigrationsUp(pg); err != nil {
		log.Fatal(err)
	}
	storage := db.New(pg)

	// Load private key
	pk, err := config.NewKeyFromKeystore(c.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	// Derive own address
	selfAddr := crypto.PubkeyToAddress(pk.PublicKey)

	var cancel cancelManager

	// Get and track the sequencer address
	sequencerTracker, err := synchronizer.NewSequencerTracker(c.L1)
	if err != nil {
		log.Fatal(err)
	}
	go sequencerTracker.Start()
	cancel.add(sequencerTracker.Stop)

	// Start a chain reorganization detector for components that need to reset when this happens
	detector, err := synchronizer.NewReorgDetector(c.L1.RpcURL, 1*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	err = detector.Start()
	if err != nil {
		log.Fatal(err)
	}
	cancel.add(detector.Stop)

	// Start the batch synchronizer for back-filling missed batches
	batchSynchronizer, err := synchronizer.NewBatchSynchronizer(c.L1, selfAddr, storage, detector.Subscribe())
	if err != nil {
		log.Fatal(err)
	}
	go batchSynchronizer.Start()
	cancel.add(batchSynchronizer.Stop)

	services := []jsonrpc.Service{
		{
			Name:    sync.APISYNC,
			Service: sync.NewEndpoints(storage),
		},
		{
			Name: datacom.APIDATACOM,
			Service: datacom.NewEndpoints(
				storage,
				pk,
				sequencerTracker,
			),
		},
	}

	server := jsonrpc.NewServer(c.RPC, services)

	if err = server.Start(); err != nil {
		log.Fatal(err)
	}

	cancel.waitSignal()

	return nil
}

type cancelManager struct {
	cancels []context.CancelFunc
}

func (cm *cancelManager) add(cancel context.CancelFunc) {
	cm.cancels = append(cm.cancels, cancel)
}

func (cm *cancelManager) waitSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for sig := range signals {
		switch sig {
		case os.Interrupt, os.Kill:
			log.Info("terminating application gracefully...")

			exitStatus := 0
			for _, cancel := range cm.cancels {
				cancel()
			}
			os.Exit(exitStatus)
		}
	}
}
