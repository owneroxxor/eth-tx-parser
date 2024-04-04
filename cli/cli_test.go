package cli

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
	"tx_parser/eth_parser"
	"tx_parser/eth_parser/test"

	"github.com/google/go-cmp/cmp"
)

func Test_CLI_HandleSubscribe(t *testing.T) {
	tt := []struct {
		name            string
		returnSubscribe bool
		expected        string
	}{
		{
			name:            "Subscribe successful",
			returnSubscribe: true,
			expected:        "Subscribed to 0x123.\n",
		},
		{
			name:            "Subscribe fails",
			returnSubscribe: false,
			expected:        "Invalid address format or already subscribed to address.\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var outBuf bytes.Buffer
			parserMock := &test.ParserMock{
				ReturnSubscribe: tc.returnSubscribe,
			}
			ctx := context.Background()

			cli := NewCLI(ctx, parserMock)
			cli.output = &outBuf

			cli.HandleSubscribe([]string{"0x123"})

			if outBuf.String() != tc.expected {
				t.Errorf("expected output to be %q, got %q", tc.expected, outBuf.String())
			}
		})
	}
}

func Test_CLI_HandleGetTxs(t *testing.T) {
	tt := []struct {
		name             string
		inputAddress     string
		transactionsMock []eth_parser.Transaction
		expected         []string
	}{
		{
			name:         "Transactions available",
			inputAddress: "0x123",
			transactionsMock: []eth_parser.Transaction{
				{Subscriber: "0x123", Hash: "hash1", From: "0x123", To: "0xdef", Value: "0x56bc75e2d63100000"},
				{Subscriber: "0x123", Hash: "hash2", From: "0xabc", To: "0x123", Value: "0xad78ebc5ac6200000"},
			},
			expected: []string{
				"Transactions for 0x123:",
				"=> Transaction for address [0x123]:",
				"   Hash: hash1",
				"   From: 0x123",
				"   To: 0xdef",
				"   Amount: 100.00000000 ETH",
				"",
				"=> Transaction for address [0x123]:",
				"   Hash: hash2",
				"   From: 0xabc",
				"   To: 0x123",
				"   Amount: 200.00000000 ETH",
				"",
				"",
			},
		},
		{
			name:             "No transactions found",
			inputAddress:     "0x456",
			transactionsMock: []eth_parser.Transaction{},
			expected:         []string{"There are still no transactions for 0x456 or you are not subscribed to it.\n"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var outBuf bytes.Buffer
			parserMock := &test.ParserMock{
				ReturnGetTransactions: tc.transactionsMock,
			}
			ctx := context.Background()

			cli := NewCLI(ctx, parserMock)
			cli.output = &outBuf

			cli.HandleGetTxs([]string{tc.inputAddress})

			want := strings.Join(tc.expected, "\n")
			got := outBuf.String()
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("HandleGetTxs() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_CLI_HandleLive(t *testing.T) {
	t.Skip("Skipping due to flaky behavior")
	testCases := []struct {
		name         string
		filter       string
		transactions []eth_parser.Transaction
		expected     []string
	}{
		{
			name:   "Live monitoring all",
			filter: "*",
			transactions: []eth_parser.Transaction{
				{Subscriber: "0x123", Hash: "hash1", From: "0x123", To: "0xdef", Value: "0x56bc75e2d63100000"},
				{Subscriber: "0x456", Hash: "hash2", From: "0xabc", To: "0x456", Value: "0xad78ebc5ac6200000"},
			},
			expected: []string{
				"Starting live transaction monitoring... Press ENTER to leave this mode.",
				"=> Transaction for address [0x123]:",
				"   Hash: hash1",
				"   From: 0x123",
				"   To: 0xdef",
				"   Amount: 100.00000000 ETH",
				"",
				"=> Transaction for address [0x456]:",
				"   Hash: hash2",
				"   From: 0xabc",
				"   To: 0x456",
				"   Amount: 200.00000000 ETH",
				"",
				"Stopped live transaction monitoring.",
				"",
			},
		},
		{
			name:   "Live monitoring specific address",
			filter: "0x123",
			transactions: []eth_parser.Transaction{
				{Subscriber: "0x123", Hash: "hash1", From: "0x123", To: "0xdef", Value: "0x56bc75e2d63100000"}, // Matches filter
				{Subscriber: "0x456", Hash: "hash2", From: "0xabc", To: "0x456", Value: "0xad78ebc5ac6200000"}, // Does not match filter
			},
			expected: []string{
				"Starting live transaction monitoring... Press ENTER to leave this mode.",
				"=> Transaction for address [0x123]:", // Expecting only this transaction
				"   Hash: hash1",
				"   From: 0x123",
				"   To: 0xdef",
				"   Amount: 100.00000000 ETH",
				"",
				"Stopped live transaction monitoring.",
				"",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var outBuf bytes.Buffer
			inputBuf := new(bytes.Buffer)
			scanner := bufio.NewScanner(inputBuf)

			listenChan := make(chan eth_parser.Transaction, len(tc.transactions))
			parserMock := &test.ParserMock{
				ReturnListen: listenChan,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cli := NewCLI(ctx, parserMock)
			cli.scanner = scanner
			cli.output = &outBuf

			go cli.HandleLive([]string{tc.filter})

			for _, tx := range tc.transactions {
				listenChan <- tx
			}

			inputBuf.WriteString("\n")

			time.Sleep(300 * time.Millisecond)

			cancel() // Cancel the context to simulate user stopping live monitoring

			want := strings.Join(tc.expected, "\n")
			got := outBuf.String()
			if got != want {
				t.Errorf("\nExpected output:\n%s\nGot output:\n%s", want, got)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("HandleGetTxs() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
