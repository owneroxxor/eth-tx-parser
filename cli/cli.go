package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	parser "tx_parser/eth_parser"
)

type CLI struct {
	ctx     context.Context
	parser  parser.Parser
	scanner *bufio.Scanner
	output  io.Writer // Adds flexibility to read the CLI output for testing purpouses
}

func NewCLI(ctx context.Context, p parser.Parser) *CLI {
	return &CLI{
		ctx:     ctx,
		parser:  p,
		scanner: bufio.NewScanner(os.Stdin),
		output:  os.Stdout,
	}
}

func (cli *CLI) Run() {
	fmt.Fprintln(cli.output, "\nEthereum Transaction Monitor CLI")
	fmt.Fprintln(cli.output, "\nCommands:")
	fmt.Fprintln(cli.output, "- subscribe [eth_address]: monitor transactions for a given Ethereum address.")
	fmt.Fprintln(cli.output, "- get_txs [eth_address]: get all transactions stored for a given Ethereum address.")
	fmt.Fprintln(cli.output, "- live [*|eth_address]: show live transactions for all or a specific subscribed Ethereum address.")
	fmt.Fprintln(cli.output, "\nPress ENTER (without typing a command) at any time to exit.")

	for {
		if err := cli.ctx.Err(); err != nil {
			break // Will break out of the loop if the context was cancelled
		}

		fmt.Print("\nEnter command: ")
		if !cli.scanner.Scan() {
			break // Exit on scan failure
		}

		line := cli.scanner.Text()
		if line == "" {
			break // Exit on empty line
		}

		cli.handleCommand(strings.Fields(line))
	}

	if err := cli.scanner.Err(); err != nil {
		fmt.Fprintln(cli.output, "error reading from cli input:", err)
	}

	cli.parser.Stop()
}

func (cli *CLI) handleCommand(parts []string) {
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "subscribe":
		cli.HandleSubscribe(parts[1:])
	case "get_txs":
		cli.HandleGetTxs(parts[1:])
	case "live":
		cli.HandleLive(parts[1:])
	default:
		fmt.Fprintln(cli.output, "Unknown command:", parts[0])
	}
}

func (cli *CLI) HandleSubscribe(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(cli.output, "Usage: subscribe [eth_address]")
		return
	}
	address := args[0]
	if cli.parser.Subscribe(address) {
		fmt.Fprintf(cli.output, "Subscribed to %s.\n", address)
	} else {
		fmt.Fprintln(cli.output, "Invalid address format or already subscribed to address.")
	}
}

func (cli *CLI) HandleGetTxs(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(cli.output, "Usage: get_txs [eth_address]")
		return
	}
	address := args[0]
	transactions := cli.parser.GetTransactions(address)
	if len(transactions) == 0 {
		fmt.Fprintf(cli.output, "There are still no transactions for %s or you are not subscribed to it.\n", address)
	} else {
		fmt.Fprintf(cli.output, "Transactions for %s:\n", address)
		for _, tx := range transactions {
			cli.printTx(tx)
		}
	}
}

func (cli *CLI) HandleLive(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(cli.output, "Usage: live [*|eth_address]")
		return
	}
	filter := args[0]

	fmt.Fprintln(cli.output, "Starting live transaction monitoring... Press ENTER to leave this mode.")
	ctx, cancel := context.WithCancel(cli.ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case tx, ok := <-cli.parser.Listen():
				if !ok {
					return
				}
				if filter == "*" || filter == tx.From || filter == tx.To {
					cli.printTx(tx)
				}
			}
		}
	}()

	cli.scanner.Scan()
	cancel()
	fmt.Fprintln(cli.output, "Stopped live transaction monitoring.")
}

func (cli *CLI) printTx(tx parser.Transaction) {
	fmt.Fprintf(cli.output, "=> Transaction for address [%s]:\n", tx.Subscriber)
	fmt.Fprintf(cli.output, "   Hash: %s\n", tx.Hash)
	fmt.Fprintf(cli.output, "   From: %s\n", tx.From)
	fmt.Fprintf(cli.output, "   To: %s\n", tx.To)
	fmt.Fprintf(cli.output, "   Amount: %s ETH\n\n", tx.ETHAmount())
}
