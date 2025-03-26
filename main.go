package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alexandru197/ethparser/parser"
	"github.com/alexandru197/ethparser/server"
)

func main() {
	// Initialize the EthParser with the Ethereum JSON RPC endpoint.
	ethParser := parser.NewEthParser("https://ethereum-rpc.publicnode.com", 10*time.Second)

	// Setup HTTP endpoints using the server package.
	srv := server.NewServer(ethParser)
	srv.SetupRoutes()

	port := 8080
	log.Printf("Starting HTTP server on port %d...", port)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
