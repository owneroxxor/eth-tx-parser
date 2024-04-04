package eth_parser

import (
	"sync"
)

type Storage interface {
	Subscribe(address string) bool
	IsSubscribed(address string) bool
	AddTransaction(address string, tx Transaction) bool
	GetTransactions(address string) []Transaction
	SetLastProcessedBlockNum(num uint64) bool
	GetLastProcessedBlockNum() uint64
}

type MemoryStorage struct {
	mu                    sync.Mutex
	subscribers           map[string]bool
	transactions          map[string][]Transaction
	lastProcessedBlockNum uint64
}

func NewMemoryStorage() Storage {
	return &MemoryStorage{
		subscribers:  make(map[string]bool),
		transactions: make(map[string][]Transaction),
	}
}

func (s *MemoryStorage) Subscribe(address string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.subscribers[address]; !exists {
		s.subscribers[address] = true
		return true
	}
	return false
}

func (s *MemoryStorage) IsSubscribed(address string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.subscribers[address]
	return exists
}

func (s *MemoryStorage) AddTransaction(address string, tx Transaction) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.subscribers[address]; exists {
		s.transactions[address] = append(s.transactions[address], tx)
		return true
	}
	return false
}

func (s *MemoryStorage) GetTransactions(address string) []Transaction {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.transactions[address]
}

func (s *MemoryStorage) SetLastProcessedBlockNum(num uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastProcessedBlockNum = num
	return true
}

func (s *MemoryStorage) GetLastProcessedBlockNum() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastProcessedBlockNum
}
