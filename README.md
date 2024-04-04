# Ethereum Transactions Monitor CLI
The ETH Transactions Parser CLI is a tool that enables users to monitor, store, and stream live Ethereum blockchain transactions for specified addresses.

## Quick Start
Ensure you have Go installed on your machine. Then, clone the project and run it using:
```bash
go mod tidy
go run main.go
```

## Features
- **Subscription**: Monitor transactions for any Ethereum address.
- **Transaction Retrieval**: Access stored transactions for monitored addresses.
- **Live Monitoring**: Receive real-time updates for transactions involving subscribed addresses.
  
## Usage
### Subscribing to an Address
Monitor transactions for 0x...:
```
subscribe 0x...
```
### Retrieving Transactions
Fetch all transactions for 0x...:
```
get_txs 0x...
```
### Live Transaction Monitoring
For a specific address:
```
live 0x...
```
For all subscribed addresses:
```
live *
```

## Architecture
The application comprises two main components:

- **CLI**: Manages user interaction.
- **ETH Parser**: Interfaces with the Ethereum blockchain.

### Parser Interface
Implements core functionalities, accessible through:
```go
type Parser interface {
	GetCurrentBlock() uint64
	Subscribe(address string) bool
	GetTransactions(address string) []Transaction
	Listen() <-chan Transaction // Live transaction feed
	Stop()                      // Halts monitoring
}
```
The `Parser` interface provides both polling and push methods - `GetTransactions()` and `Listen()` respectively - to keep track of the subscribed addresses transactions. This interface can be hooked to a notifications service for example, where it would notify for any incoming/outgoing transaction for a given monitored ETH address.
### Storage Interface
Flexible storage management, with a thread-safe in-memory default:
```go
type Storage interface {
	// Subscription and transaction management
}
```
By default, a thread-safe memory storage is used, but one may use a different storage mechanism by directly passing it to the instantiation function:
```go
parser := eth_parser.NewEthereumParser(ctx, customStorage)
```
