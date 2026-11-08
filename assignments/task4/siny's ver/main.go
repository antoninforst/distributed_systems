package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "ricart-agrawala/proto"
)

func main() {
	// Read the node ID from the command line.
	nodeID := flag.String("id", "", "Node ID (A, B or C)")
	flag.Parse()

	// Available nodes and their addresses.
	allNodes := map[string]string{
		"A": "localhost:50051",
		"B": "localhost:50052",
		"C": "localhost:50053",
	}

	// Check that the provided ID is valid.
	address, exists := allNodes[*nodeID]
	if !exists {
		log.Fatal("Invalid node ID. Use -id=A, -id=B or -id=C")
	}

	// Build the list of peers.
	peers := make(map[string]string)

	for peerID, peerAddress := range allNodes {
		if peerID != *nodeID {
			peers[peerID] = peerAddress
		}
	}

	// Create this node.
	node := &Node{
		ID:                 *nodeID,
		Address:            address,
		Peers:              peers,
		Clock:              0,
		RepliesReceived:    make(map[string]bool),
		DeferredReplies:    make(map[string]bool),
		DeferredReplyChans: make(map[string]chan struct{}),
		Clients:            make(map[string]pb.MutualExclusionClient),
	}

	// Start the grpc server.
	if err := node.StartServer(); err != nil {
		log.Fatalf("Could not start node %s: %v", *nodeID, err)
	}

	defer node.CloseConnections()

	// Create grpc clients for the other nodes.
	if err := node.ConnectToPeers(); err != nil {
		log.Fatalf("Could not connect node %s to its peers: %v", *nodeID, err)
	}

	log.Printf("Node %s is ready", *nodeID)

	// Give the other nodes time to start.
	log.Printf("Node %s will request the critical section in 5 seconds", *nodeID)
	startupDelay := map[string]time.Duration{
		"A": 5 * time.Second,
		"B": 5 * time.Second,
		"C": 5 * time.Second,
	}

	time.Sleep(startupDelay[*nodeID])

	// Request permission from all peers.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := node.RequestCriticalSection(ctx); err != nil {
		log.Fatalf("Node %s could not enter the critical section: %v", *nodeID, err)
	}

	// Simulate work inside the critical section.
	log.Printf("Node %s is executing the critical section", *nodeID)
	time.Sleep(2 * time.Second)

	// Release the critical section.
	if err := node.ReleaseCriticalSection(); err != nil {
		log.Fatalf("Node %s could not release the critical section: %v", *nodeID, err)
	}

	log.Printf("Node %s has finished successfully", *nodeID)

	// Keep the gRPC server running to handle incoming requests.
	log.Printf("Node %s is still running and waiting for incoming requests", *nodeID)

	select {}
}
