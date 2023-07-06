package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/0xPolygon/supernets2.0-data-availability/config"
	"github.com/0xPolygon/supernets2.0-data-availability/db"
	"github.com/0xPolygon/supernets2.0-data-availability/dummyinterfaces"
	"github.com/0xPolygon/supernets2.0-data-availability/services/datacom"
	"github.com/0xPolygon/supernets2.0-data-availability/services/sync"
	"github.com/0xPolygonHermez/zkevm-node"
	dbConf "github.com/0xPolygonHermez/zkevm-node/db"
	"github.com/0xPolygonHermez/zkevm-node/jsonrpc"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

const appName = "supernets2.0-data-availability"

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
	app.Version = zkevm.Version
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
	sequencerAddr := common.HexToAddress(c.SequencerAddr)

	// Register services
	server := jsonrpc.NewServer(
		c.RPC,
		0,
		&dummyinterfaces.DummyPool{},
		&dummyinterfaces.DummyState{},
		&dummyinterfaces.DummyStorage{},
		[]jsonrpc.Service{
			{
				Name:    sync.APISYNC,
				Service: sync.NewSyncEndpoints(storage),
			},
			{
				Name: datacom.APIDATACOM,
				Service: datacom.NewDataComEndpoints(
					storage,
					sequencerAddr,
					pk,
				),
			},
		},
	)

	// Run!
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	var cancelFuncs []context.CancelFunc
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
