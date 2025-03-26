package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexandru197/ethparser/parser"
)

// dummyParser is a simple implementation of the Parser interface for testing.
type dummyParser struct {
	currentBlock int
	txs          map[string][]parser.Transaction
	subs         map[string]bool
}

func (d *dummyParser) GetCurrentBlock() int {
	return d.currentBlock
}
func (d *dummyParser) Subscribe(address string) bool {
	if d.subs == nil {
		d.subs = make(map[string]bool)
	}
	addr := strings.ToLower(address)
	if d.subs[addr] {
		return false
	}
	d.subs[addr] = true
	return true
}
func (d *dummyParser) GetTransactions(address string) []parser.Transaction {
	addr := strings.ToLower(address)
	return d.txs[addr]
}

func TestBlockHandler(t *testing.T) {
	dp := &dummyParser{currentBlock: 100}
	s := NewServer(dp)
	req := httptest.NewRequest("GET", "/block", nil)
	w := httptest.NewRecorder()
	s.blockHandler(w, req)

	var resp map[string]int
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["current_block"] != 100 {
		t.Errorf("Expected current block 100, got %d", resp["current_block"])
	}
}

func TestSubscribeHandler(t *testing.T) {
	dp := &dummyParser{}
	s := NewServer(dp)
	req := httptest.NewRequest("GET", "/subscribe?address=0xabc", nil)
	w := httptest.NewRecorder()
	s.subscribeHandler(w, req)
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["subscribed"] != true {
		t.Errorf("Expected subscription to be true, got %v", resp["subscribed"])
	}
	// Test duplicate subscription.
	req = httptest.NewRequest("GET", "/subscribe?address=0xabc", nil)
	w = httptest.NewRecorder()
	s.subscribeHandler(w, req)
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["subscribed"] != false {
		t.Errorf("Expected duplicate subscription to be false, got %v", resp["subscribed"])
	}
}

func TestTransactionsHandler(t *testing.T) {
	dp := &dummyParser{
		txs: map[string][]parser.Transaction{
			"0xabc": {
				{Hash: "0x123", From: "0xabc", To: "0xdef", Value: "0x1"},
			},
		},
	}
	s := NewServer(dp)
	req := httptest.NewRequest("GET", "/transactions?address=0xabc", nil)
	w := httptest.NewRecorder()
	s.transactionsHandler(w, req)
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	txs, ok := resp["transactions"].([]interface{})
	if !ok || len(txs) != 1 {
		t.Errorf("Expected one transaction, got %v", resp["transactions"])
	}
}
