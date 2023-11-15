package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	dataavailability "github.com/0xPolygon/cdk-data-availability"
	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/sequencer"
	"github.com/0xPolygon/cdk-data-availability/services/datacom"
	"github.com/0xPolygon/cdk-data-availability/services/sync"
	"github.com/0xPolygon/cdk-data-availability/synchronizer"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

const appName = "cdk-data-availability"

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
	setupLog(c.Log)

	// Prepare DB
	pg, err := db.NewSQLDB(c.DB)
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
	// Load EtherMan
	etherman, err := etherman.New(c.L1)
	if err != nil {
		log.Fatal(err)
	}

	// derive address
	selfAddr := crypto.PubkeyToAddress(pk.PublicKey)

	// ensure synchro/reorg start block is set
	err = synchronizer.InitStartBlock(storage, c.L1)
	if err != nil {
		log.Fatal(err)
	}

	var cancelFuncs []context.CancelFunc

	sequencerTracker, err := sequencer.NewSequencerTracker(c.L1, etherman)
	if err != nil {
		log.Fatal(err)
	}
	go sequencerTracker.Start()
	cancelFuncs = append(cancelFuncs, sequencerTracker.Stop)

	detector, err := synchronizer.NewReorgDetector(c.L1.RpcURL, 1*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	err = detector.Start()
	if err != nil {
		log.Fatal(err)
	}

	cancelFuncs = append(cancelFuncs, detector.Stop)

	batchSynchronizer, err := synchronizer.NewBatchSynchronizer(c.L1, selfAddr, storage, detector.Subscribe(), etherman, sequencerTracker)
	if err != nil {
		log.Fatal(err)
	}
	go batchSynchronizer.Start()
	cancelFuncs = append(cancelFuncs, batchSynchronizer.Stop)

	// Register services
	server := rpc.NewServer(
		c.RPC,
		[]rpc.Service{
			{
				Name:    sync.APISYNC,
				Service: sync.NewSyncEndpoints(storage),
			},
			{
				Name: datacom.APIDATACOM,
				Service: datacom.NewDataComEndpoints(
					storage,
					pk,
					sequencerTracker,
				),
			},
		},
	)

	// Run!
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}

	waitSignal(cancelFuncs)
	return nil
}

func setupLog(c log.Config) {
	log.Init(c)
}

func waitSignal(cancelFuncs []context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for sig := range signals {
		switch sig {
		case os.Interrupt, os.Kill:
			log.Info("terminating application gracefully...")

			exitStatus := 0
			for _, cancel := range cancelFuncs {
				cancel()
			}
			os.Exit(exitStatus)
		}
	}
}
