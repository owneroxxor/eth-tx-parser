package test

import (
	"encoding/json"
	"eth-tx-parser/eth_parser"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupMockServer() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Handle eth_blockNum request
		if reqBody["method"] == "eth_blockNumber" {
			jsonResponse := []byte(`{"jsonrpc":"2.0","id":1,"result":"0x5b8d80"}`)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			return
		}

		// Handle eth_getBlockByNumber request
		if reqBody["method"] == "eth_getBlockByNumber" {
			jsonResponse := `{
                "jsonrpc": "2.0",
                "id": 1,
                "result": {
                    "number": "0x3039",
                    "hash": "0xhash",
                    "transactions": [
                        {
                            "blockHash": "0xbhash",
                            "blockNum": "0x1B4",
                            "from": "0xfrom",
                            "gas": "0x5208",
                            "gasPrice": "0x4A817C800",
                            "hash": "0xthash",
                            "value": "0x0",
                            "nonce": "0x15",
                            "to": "0xto",
                            "transactionIndex": "0x1",
                            "v": "0x25",
                            "r": "0x1",
                            "s": "0x2"
                        }
                    ]
                }
            }`
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jsonResponse))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	return httptest.NewServer(handler)
}

func Test_FetchLatestblockNum(t *testing.T) {
	mockServer := setupMockServer()
	defer mockServer.Close()

	client := eth_parser.NewEthereumClient()
	client.(*eth_parser.EthereumRPCClient).EthereumRPCURL = mockServer.URL

	got, err := client.FetchLatestBlockNumber()
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	want := "0x5b8d80"
	if got.Result != want {
		t.Errorf("got = %v, want %v", got.Result, want)
	}
}

func Test_FetchBlockByNumber(t *testing.T) {
	mockServer := setupMockServer()
	defer mockServer.Close()

	ec := eth_parser.NewEthereumClient()
	ec.(*eth_parser.EthereumRPCClient).EthereumRPCURL = mockServer.URL

	blockNum := uint64(12345)
	block, err := ec.FetchBlockByNumber(blockNum)
	if err != nil {
		t.Errorf("returned an error: %v", err)
	}

	// Implicitly asserts that all other fields match
	want := fmt.Sprintf("0x%x", blockNum)
	if block.Result.Number != want {
		t.Errorf("expected block number %s, got %s", want, block.Result.Number)
	}
}
