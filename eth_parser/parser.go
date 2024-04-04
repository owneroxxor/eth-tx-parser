package eth_parser

import (
	"context"
	"log"
	"regexp"
	"sync"
	"time"
)

type Parser interface {
	GetCurrentBlock() uint64
	Subscribe(address string) bool
	GetTransactions(address string) []Transaction
	Listen() <-chan Transaction // Provides event-driven architecture capability
	Stop()                      // Stops the monitor
}

type EthereumParser struct {
	ctx              context.Context
	Client           EthereumClient
	storage          Storage // Easily attachable storage interface
	tx_chan          chan Transaction
	BlockPollingFreq time.Duration
	monitorStarted   bool
	closeOnce        sync.Once
}

func NewEthereumParser(ctx context.Context, storage Storage) Parser {
	if storage == nil {
		storage = NewMemoryStorage() // Use default storage
	}

	ep := &EthereumParser{
		ctx:              ctx,
		Client:           NewEthereumClient(),
		storage:          storage,
		tx_chan:          make(chan Transaction, 10),
		BlockPollingFreq: 5 * time.Second,
	}

	go ep.startMonitor()

	return ep
}

func (ep *EthereumParser) GetCurrentBlock() uint64 {
	return ep.storage.GetLastProcessedBlockNum()
}

func (ep *EthereumParser) Subscribe(address string) bool {
	validAddress := regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)
	if validAddress.MatchString(address) {
		return ep.storage.Subscribe(address)
	}
	return false
}

func (ep *EthereumParser) GetTransactions(address string) []Transaction {
	return ep.storage.GetTransactions(address)
}

func (ep *EthereumParser) Listen() <-chan Transaction {
	return ep.tx_chan
}

func (ep *EthereumParser) startMonitor() {
	if ep.monitorStarted {
		return
	}
	ep.monitorStarted = true

	ticker := time.NewTicker(ep.BlockPollingFreq)

	defer ticker.Stop()
	defer ep.Stop()

	for {
		select {
		case <-ep.ctx.Done():
			return

		case <-ticker.C:
			latestBlockNumInstance, err := ep.Client.FetchLatestBlockNumber()
			if err != nil {
				log.Println("failed to fetch latest block:", err)
				continue
			}
			latestBlockNum, err := latestBlockNumInstance.ToUint64()
			if err != nil {
				log.Println("impossible to use latest block number:", err)
				return
			}

			if ep.storage.GetLastProcessedBlockNum() == 0 { // Only executed in the first run
				ep.storage.SetLastProcessedBlockNum(latestBlockNum - 1)
			}

			if latestBlockNum > ep.storage.GetLastProcessedBlockNum() {
				for blockNum := ep.storage.GetLastProcessedBlockNum() + 1; blockNum <= latestBlockNum; blockNum++ {
					block, err := ep.Client.FetchBlockByNumber(blockNum)
					if err != nil {
						log.Println("impossible to retrieve block information:", err)
						return
					}
					for _, tx := range block.Result.Transactions {
						for _, addr := range []string{tx.From, tx.To} {
							if ep.storage.IsSubscribed(addr) {
								tx.Subscriber = addr
								if ok := ep.storage.AddTransaction(tx.Subscriber, tx); !ok {
									log.Println("failed to store transaction, bad storage. exiting now")
									return
								}
								// Send to the live feed
								select {
								case ep.tx_chan <- tx:
								default: // Skip if the channel is full
								}
								break
							}
						}
					}
				}
				if ok := ep.storage.SetLastProcessedBlockNum(latestBlockNum); !ok {
					log.Println("failed to set the last processed block number, bad storage. exiting now")
					return
				}
			}
		}
	}
}

func (ep *EthereumParser) Stop() {
	ep.closeOnce.Do(func() {
		close(ep.tx_chan)
	})
	ep.monitorStarted = false
}
