package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Transaction represents an Ethereum transaction (simplified).
type Transaction struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
}

// Parser defines the operations of our blockchain parser.
type Parser interface {
	// GetCurrentBlock returns the last parsed block number.
	GetCurrentBlock() int
	// Subscribe adds an address to be observed.
	Subscribe(address string) bool
	// GetTransactions returns inbound/outbound transactions for an address.
	GetTransactions(address string) []Transaction
}

// EthParser implements the Parser interface.
type EthParser struct {
	mu           sync.Mutex
	currentBlock int
	subscribers  map[string]bool          // Subscribed addresses.
	txStore      map[string][]Transaction // Mapping from address to transactions.
	rpcEndpoint  string
	pollInterval time.Duration
	stopCh       chan struct{}
}

// NewEthParser returns an initialized EthParser.
func NewEthParser(rpcEndpoint string, pollInterval time.Duration) *EthParser {
	ep := &EthParser{
		subscribers:  make(map[string]bool),
		txStore:      make(map[string][]Transaction),
		rpcEndpoint:  rpcEndpoint,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
	}
	go ep.startPolling()
	return ep
}

// Stop terminates the background polling.
func (p *EthParser) Stop() {
	close(p.stopCh)
}

// GetCurrentBlock returns the last parsed block number.
func (p *EthParser) GetCurrentBlock() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currentBlock
}

// Subscribe adds an address to the observer list.
func (p *EthParser) Subscribe(address string) bool {
	address = strings.ToLower(address)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.subscribers[address] {
		return false
	}
	p.subscribers[address] = true
	return true
}

// GetTransactions returns stored transactions for an address.
func (p *EthParser) GetTransactions(address string) []Transaction {
	address = strings.ToLower(address)
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.txStore[address]
}

// startPolling periodically checks for new blocks and processes them.
func (p *EthParser) startPolling() {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			blockNum, err := p.fetchCurrentBlockNumber()
			if err != nil {
				log.Println("Error fetching block number:", err)
				continue
			}
			p.mu.Lock()
			if blockNum > p.currentBlock {
				// Process new blocks sequentially.
				for b := p.currentBlock + 1; b <= blockNum; b++ {
					if err := p.processBlock(b); err != nil {
						log.Printf("Error processing block %d: %v", b, err)
					}
				}
				p.currentBlock = blockNum
			}
			p.mu.Unlock()
		case <-p.stopCh:
			return
		}
	}
}

// fetchCurrentBlockNumber makes a JSON-RPC call to get the latest block number.
func (p *EthParser) fetchCurrentBlockNumber() (int, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	resp, err := http.Post(p.rpcEndpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, err
	}
	cleaned := strings.TrimPrefix(result.Result, "0x")
	blockNum, err := strconv.ParseInt(cleaned, 16, 64)
	if err != nil {
		return 0, err
	}
	return int(blockNum), nil
}

// processBlock fetches block details (including transactions) for a given block number.
func (p *EthParser) processBlock(blockNum int) error {
	hexBlock := fmt.Sprintf("0x%x", blockNum)
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{hexBlock, true},
		"id":      1,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(p.rpcEndpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Result struct {
			Transactions []Transaction `json:"transactions"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return err
	}
	// Filter transactions for subscribed addresses.
	for _, tx := range result.Result.Transactions {
		from := strings.ToLower(tx.From)
		to := strings.ToLower(tx.To)
		p.mu.Lock()
		if p.subscribers[from] {
			p.txStore[from] = append(p.txStore[from], tx)
		}
		if p.subscribers[to] {
			p.txStore[to] = append(p.txStore[to], tx)
		}
		p.mu.Unlock()
	}
	return nil
}
