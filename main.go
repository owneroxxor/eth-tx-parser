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
	parser := eth_parser.NewEthereumParser(ctx, nil) // Passing nil so it uses the default storage
	cli := cli.NewCLI(ctx, parser)

	// Setup signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
		fmt.Println("\nreceived termination signal, exiting...")
		os.Exit(0)
	}()

	cli.Run()
}
