package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alexandru197/ethparser/parser"
)

// Server wraps a Parser instance to expose HTTP endpoints.
type Server struct {
	Parser parser.Parser
}

// NewServer creates a new Server instance.
func NewServer(p parser.Parser) *Server {
	return &Server{Parser: p}
}

// blockHandler returns the current block number.
func (s *Server) blockHandler(w http.ResponseWriter, r *http.Request) {
	current := s.Parser.GetCurrentBlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"current_block": current})
}

// subscribeHandler subscribes an address passed as a query parameter.
func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter missing", http.StatusBadRequest)
		return
	}
	subscribed := s.Parser.Subscribe(address)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address":    address,
		"subscribed": subscribed,
	})
}

// transactionsHandler returns transactions for a given address.
func (s *Server) transactionsHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter missing", http.StatusBadRequest)
		return
	}
	// Normalize address.
	address = strings.ToLower(address)
	txs := s.Parser.GetTransactions(address)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address":      address,
		"transactions": txs,
	})
}

// SetupRoutes registers the HTTP endpoints.
func (s *Server) SetupRoutes() {
	http.HandleFunc("/block", s.blockHandler)
	http.HandleFunc("/subscribe", s.subscribeHandler)
	http.HandleFunc("/transactions", s.transactionsHandler)
}
