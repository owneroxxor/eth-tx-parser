package eth_parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type EthereumClient interface {
	FetchLatestBlockNumber() (*BlockNumber, error)
	FetchBlockByNumber(blockNumber uint64) (*Block, error)
}

type EthereumRPCClient struct {
	HTTPClient          *http.Client
	EthereumRPCURL      string
	RPCVersion          string
	ReqEncoding         string
	EthBlockNumber      string
	EthGetBlockByNumber string
	seq                 uint64

	// Exponential backoff settings
	maxAttempts  int // Maximum retriable attempts
	backoffScale int // Starting backoff scale in seconds
}

func NewEthereumClient() EthereumClient {
	return &EthereumRPCClient{
		HTTPClient:          &http.Client{},
		EthereumRPCURL:      "https://cloudflare-eth.com",
		RPCVersion:          "2.0",
		ReqEncoding:         "application/json",
		EthBlockNumber:      "eth_blockNumber",
		EthGetBlockByNumber: "eth_getBlockByNumber",
		maxAttempts:         5,
		backoffScale:        1,
		seq:                 0,
	}
}

func (ec *EthereumRPCClient) request(method string, params []interface{}) ([]byte, error) {
	reqBody := map[string]interface{}{
		"jsonrpc": ec.RPCVersion,
		"method":  method,
		"params":  params,
		"id":      ec.seq,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize body")
	}

	req, err := http.NewRequest("POST", ec.EthereumRPCURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new request")
	}
	req.Header.Set("Content-Type", ec.ReqEncoding)

	return ec.expBackoff(req)
}

func (ec *EthereumRPCClient) expBackoff(req *http.Request) ([]byte, error) {
	var resp *http.Response
	var err error
	var body []byte

	scale := ec.backoffScale

	for attempt := 0; attempt < ec.maxAttempts; attempt++ {
		ec.seq++
		resp, err = ec.HTTPClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				body, err = io.ReadAll(resp.Body)
				if err == nil {
					return body, errors.Wrap(err, "failed to read from body")
				}
			} else if !shouldRetry(resp.StatusCode) {
				return nil, errors.Wrapf(err, "received non-retryable status code %d", resp.StatusCode)
			}
		}

		// Calculate and wait for the next attempt's backoff duration
		wait := time.Duration(rand.Intn(int(scale))) * time.Second
		time.Sleep(wait)
		scale *= 2 // Double the scale for the next attempt
	}

	return nil, errors.Wrapf(err, "after %d attempts, last error", ec.maxAttempts)
}

func shouldRetry(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout, // 408
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

func (ec *EthereumRPCClient) FetchLatestBlockNumber() (*BlockNumber, error) {
	body, err := ec.request(ec.EthBlockNumber, []interface{}{})
	if err != nil {
		return nil, errors.Wrap(err, "request for latest block number failed")
	}

	blockNum := &BlockNumber{}
	if err := json.Unmarshal(body, blockNum); err != nil {
		return nil, errors.Wrap(err, "failed on the deserialization of block number")
	}

	return blockNum, nil
}

func (ec *EthereumRPCClient) FetchBlockByNumber(blockNumber uint64) (*Block, error) {
	blockNumberHex := fmt.Sprintf("0x%x", blockNumber)
	body, err := ec.request(ec.EthGetBlockByNumber, []interface{}{blockNumberHex, true})
	if err != nil {
		return nil, errors.Wrap(err, "request to fetch block by number failed")
	}

	block := &Block{}
	err = json.Unmarshal(body, block)
	if err != nil {
		return nil, errors.Wrap(err, "failed on the deserialization of block")
	}

	return block, err
}
