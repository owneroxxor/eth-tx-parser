package main

import (
	"context"
	"eth-tx-parser/cli"
	"eth-tx-parser/eth_parser"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	parser := eth_parser.NewEthereumParser(ctx, nil) // Passing nil so it uses the default storage
	cli := cli.NewCLI(ctx, parser)

	setupSignalHandling(cancel)

	cli.Run()
}

func setupSignalHandling(cancelFunc context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nReceived termination signal, exiting...")
		cancelFunc()
		os.Exit(0)
	}()
}
