package test

import (
	"fmt"
	"tx_parser/eth_parser"
)

type ParserMock struct {
	ReturnGetCurrentBlock uint64
	ReturnSubscribe       bool
	ReturnGetTransactions []eth_parser.Transaction
	ReturnListen          chan eth_parser.Transaction
}

func (m *ParserMock) GetCurrentBlock() uint64 {
	return m.ReturnGetCurrentBlock
}

func (m *ParserMock) Subscribe(address string) bool {
	return m.ReturnSubscribe
}

func (m *ParserMock) GetTransactions(address string) []eth_parser.Transaction {
	return m.ReturnGetTransactions
}

func (m *ParserMock) Listen() <-chan eth_parser.Transaction {
	return m.ReturnListen
}

func (m *ParserMock) Stop() {}

type ClientMock struct {
	LatestBlockNumber *eth_parser.BlockNumber
	BlockByNumber     map[uint64]*eth_parser.Block
	Err               error
}

func NewClientMock() *ClientMock {
	return &ClientMock{
		BlockByNumber: make(map[uint64]*eth_parser.Block),
	}
}

func (m *ClientMock) FetchLatestBlockNumber() (*eth_parser.BlockNumber, error) {
	return m.LatestBlockNumber, m.Err
}

func (m *ClientMock) FetchBlockByNumber(blockNumber uint64) (*eth_parser.Block, error) {
	if block, exists := m.BlockByNumber[blockNumber]; exists {
		return block, m.Err
	}
	return nil, fmt.Errorf("block %d not found", blockNumber)
}

func (m *ClientMock) SetLatestBlockNumber(blockNum uint64) {
	m.LatestBlockNumber = &eth_parser.BlockNumber{
		JsonRPC: "2.0",
		Result:  fmt.Sprintf("0x%x", blockNum),
	}
}

func (m *ClientMock) SetBlockByNumber(blockNum uint64, block *eth_parser.Block) {
	m.BlockByNumber[blockNum] = block
}
