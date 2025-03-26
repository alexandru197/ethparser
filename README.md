# Ethereum Blockchain Parser

## Overview
A lightweight, memory-based Ethereum blockchain transaction parser with support for address subscription and transaction tracking.

## Features
- Subscribe to Ethereum addresses
- Retrieve transactions for subscribed addresses
- In-memory transaction storage
- Minimal external dependencies

## Requirements
- Go 1.20+
- Access to Ethereum JSON-RPC endpoint

## Installation
```bash
git clone https://github.com/alexandru197/ethparser.git
cd ethparser
go mod tidy
```


## Usage
### HTTP API Mode
```bash
go run main.go
```

## API Endpoints
- `/block` - returns the current block number
- `/subscribe` - subscribes an address passed as a query parameter
- `/transactions` - returns transactions for a given address

## Project Structure:
The code is split into a parser package (containing the interface and EthParser implementation), a server package (which registers HTTP endpoints to interact with the parser), and a main.go file that wires everything together.

## Unit Tests:
The tests in parser_test.go use an httptest server to simulate Ethereum JSON‑RPC responses so that you can validate block fetching, transaction processing, and address subscriptions without relying on a live node. Similarly, the tests in server_test.go create a dummy parser to verify the HTTP endpoints.

## Limitations
- In-memory storage (not persistent)
- Basic transaction filtering
- Requires continuous running for real-time tracking

## Future Improvements
- Persistent storage support
- More advanced transaction filtering
- Performance optimizations
- Advanced notification mechanisms
- API documentation enhanced with tools such as Swagger