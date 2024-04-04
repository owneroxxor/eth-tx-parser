package eth_parser

import (
	"fmt"
	"math/big"
	"strconv"
)

type Request struct {
	JsonRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type BlockNumber struct {
	JsonRPC string `json:"jsonrpc"`
	Result  string `json:"result"`
	ID      int    `json:"id"`
}

func (bn *BlockNumber) ToUint64() (uint64, error) {
	return strconv.ParseUint(bn.Result[2:], 16, 64)
}

type Block struct {
	JsonRPC string      `json:"jsonrpc"`
	BlockID int         `json:"id"`
	Result  BlockResult `json:"result"`
}

type BlockResult struct {
	Difficulty       string        `json:"difficulty"`
	ExtraData        string        `json:"extraData"`
	GasLimit         string        `json:"gasLimit"`
	GasUsed          string        `json:"gasUsed"`
	Hash             string        `json:"hash"`
	LogsBloom        string        `json:"logsBloom"`
	Miner            string        `json:"miner"`
	MixHash          string        `json:"mixHash"`
	Nonce            string        `json:"nonce"`
	Number           string        `json:"number"`
	ParentHash       string        `json:"parentHash"`
	ReceiptsRoot     string        `json:"receiptsRoot"`
	SHA3Uncles       string        `json:"sha3Uncles"`
	Size             string        `json:"size"`
	StateRoot        string        `json:"stateRoot"`
	Timestamp        string        `json:"timestamp"`
	TotalDifficulty  string        `json:"totalDifficulty"`
	Transactions     []Transaction `json:"transactions"`
	TransactionsRoot string        `json:"transactionsRoot"`
	Uncles           []interface{} `json:"uncles"`
}

type Transaction struct {
	Subscriber       string // Additional field added to identify the subscriber party of the tx
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Value            string `json:"value"`
	V                string `json:"v"`
	R                string `json:"r"`
	S                string `json:"s"`
}

func (t *Transaction) ETHAmount() string {
	valueInWei := new(big.Int)
	valueInWei.SetString(t.Value[2:], 16)

	weiInETH := new(big.Float).SetInt(new(big.Int).SetInt64(1e18)) // 1 ETH = 1e18 wei
	valueInETH := new(big.Float).Quo(new(big.Float).SetInt(valueInWei), weiInETH)

	return fmt.Sprintf("%.8f", valueInETH)
}
