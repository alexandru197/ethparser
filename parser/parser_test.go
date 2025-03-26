package parser

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// testRPCHandler simulates Ethereum JSON-RPC responses.
func testRPCHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req map[string]interface{}
	json.Unmarshal(body, &req)
	method, ok := req["method"].(string)
	if !ok {
		http.Error(w, "Invalid method", http.StatusBadRequest)
		return
	}
	switch method {
	case "eth_blockNumber":
		// Return fixed block number: 16 (0x10)
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result":  "0x10",
		}
		json.NewEncoder(w).Encode(resp)
	case "eth_getBlockByNumber":
		// Return a block with one transaction.
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result": map[string]interface{}{
				"transactions": []map[string]interface{}{
					{
						"hash":  "0x123",
						"from":  "0xabc",
						"to":    "0xdef",
						"value": "0x1",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	default:
		http.Error(w, "Unknown method", http.StatusBadRequest)
	}
}

func TestFetchCurrentBlockNumber(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(testRPCHandler))
	defer ts.Close()

	parser := NewEthParser(ts.URL, 100*time.Millisecond)
	// Allow some time for polling to occur.
	time.Sleep(200 * time.Millisecond)
	current := parser.GetCurrentBlock()
	if current != 16 {
		t.Errorf("Expected block number 16, got %d", current)
	}
	parser.Stop()
}

func TestSubscribeAndGetTransactions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(testRPCHandler))
	defer ts.Close()

	parser := NewEthParser(ts.URL, 100*time.Millisecond)
	defer parser.Stop()

	// Subscribe an address.
	subscribed := parser.Subscribe("0xabc")
	if !subscribed {
		t.Errorf("Expected subscription to succeed")
	}
	// Wait for polling and processing block.
	time.Sleep(200 * time.Millisecond)

	txs := parser.GetTransactions("0xabc")
	if len(txs) == 0 {
		t.Errorf("Expected at least one transaction for subscribed address")
	}
	// Ensure duplicate subscription returns false.
	if parser.Subscribe("0xabc") {
		t.Errorf("Expected duplicate subscription to return false")
	}
}
