package test

import (
	"context"
	"reflect"
	"testing"
	"time"
	"tx_parser/eth_parser"
)

func Test_EthereumParser_GetCurrentBlock(t *testing.T) {
	ctx := context.Background()
	storage := eth_parser.NewMemoryStorage()
	storage.SetLastProcessedBlockNum(uint64(100))
	parser := eth_parser.NewEthereumParser(ctx, storage)

	got := parser.GetCurrentBlock()
	want := uint64(100)
	if got != want {
		t.Errorf("GetCurrentBlock() got = %v, want %v", got, want)
	}
}

func Test_EthereumParser_Subscribe(t *testing.T) {
	ctx := context.Background()
	storage := eth_parser.NewMemoryStorage()
	parser := eth_parser.NewEthereumParser(ctx, storage)

	t.Run("Subscribe valid address", func(t *testing.T) {
		address := "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5" // Valid ETH address
		got := parser.Subscribe(address)
		if !got {
			t.Errorf("Subscribe(%s) = %v, want %v", address, got, true)
		}
		if !storage.IsSubscribed(address) {
			t.Errorf("address %s was not correctly subscribed in storage", address)
		}
	})

	t.Run("Subscribe invalid address", func(t *testing.T) {
		address := "invalidAddress"
		got := parser.Subscribe(address)
		if got {
			t.Errorf("Subscribe(%s) = %v, want %v", address, got, false)
		}
	})
}

func Test_EthereumParser_GetTransactions(t *testing.T) {
	ctx := context.Background()
	storage := eth_parser.NewMemoryStorage()

	expectedTxs := []eth_parser.Transaction{{Hash: "tx1"}, {Hash: "tx2"}}
	storage.Subscribe("0x123")
	storage.AddTransaction("0x123", expectedTxs[0])
	storage.AddTransaction("0x123", expectedTxs[1])

	parser := eth_parser.NewEthereumParser(ctx, storage)
	txs := parser.GetTransactions("0x123")

	if !reflect.DeepEqual(txs, expectedTxs) {
		t.Errorf("returned %v, expected %v", txs, expectedTxs)
	}
}

func Test_EthereumParser_StartMonitor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := eth_parser.NewMemoryStorage()
	clientMock := NewClientMock()

	blockNum := uint64(2)
	clientMock.SetLatestBlockNumber(blockNum)

	blockWithTransactions := &eth_parser.Block{
		Result: eth_parser.BlockResult{
			Number: "0x1",
			Transactions: []eth_parser.Transaction{
				{
					Subscriber:       "0xabc123",
					BlockHash:        "0xbhash2",
					BlockNumber:      "0x1",
					From:             "0xabc123", // This address matches the subscription
					Gas:              "0x5208",
					GasPrice:         "0x4A817C800",
					Hash:             "0xtransactionhash1",
					Value:            "0x5af3107a4000",
					Nonce:            "0x15",
					To:               "0xdef456",
					TransactionIndex: "0x1",
					V:                "0x25",
					R:                "0x1",
					S:                "0x2",
				},
			},
		},
	}

	clientMock.SetBlockByNumber(2, blockWithTransactions)

	subscribedAddress := "0xabc123"
	storage.Subscribe(subscribedAddress)

	parser := eth_parser.NewEthereumParser(ctx, storage)
	parser.(*eth_parser.EthereumParser).Client = clientMock
	parser.(*eth_parser.EthereumParser).BlockPollingFreq = 1 * time.Millisecond

	// Allow some time for the monitor to fetch and process blocks
	time.Sleep(100 * time.Millisecond)

	cancel()

	// Verify that the transaction from block 1 is processed
	txs := storage.GetTransactions(subscribedAddress)
	if len(txs) == 0 {
		t.Fatalf("expected transactions for subscribed address %s, got none", subscribedAddress)
	}
}
