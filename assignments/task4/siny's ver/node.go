package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "ricart-agrawala/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	pb.UnimplementedMutualExclusionServer

	ID      string
	Address string

	// Addresses of the other nodes
	Peers map[string]string

	// Lamport logical clock
	Clock int64

	// Ricart-Agrawala state
	Requesting        bool
	InCriticalSection bool
	RequestTimestamp  int64

	// Permissions received from other nodes
	RepliesReceived map[string]bool

	// Requests whose replies must be postponed
	DeferredReplies map[string]bool

	// Channels used to release deferred rpc requests
	DeferredReplyChans map[string]chan struct{}

	// grpc clients used to communicate with other nodes
	Clients map[string]pb.MutualExclusionClient

	// grpc connections
	Connections map[string]*grpc.ClientConn

	// grpc server
	Server *grpc.Server

	// Protects shared node state from concurrent access
	Mu sync.Mutex
}

// incrementClock updates the logical clock for a local event.
func (n *Node) incrementClock() int64 {
	n.Mu.Lock()
	defer n.Mu.Unlock()

	n.Clock++
	return n.Clock
}

// sendTimestamp updates the clock before sending a message.
func (n *Node) sendTimestamp() int64 {
	n.Mu.Lock()
	defer n.Mu.Unlock()

	n.Clock++
	return n.Clock
}

// receiveTimestamp updates the clock after receiving a message.
func (n *Node) receiveTimestamp(receivedTimestamp int64) int64 {
	n.Mu.Lock()
	defer n.Mu.Unlock()

	if receivedTimestamp > n.Clock {
		n.Clock = receivedTimestamp
	}

	n.Clock++
	return n.Clock
}

// hasPriority checks whether this node's request has priority over an incoming request.
func (n *Node) hasPriority(timestamp int64, nodeID string) bool {
	if n.RequestTimestamp < timestamp {
		return true
	}

	if n.RequestTimestamp == timestamp && n.ID < nodeID {
		return true
	}

	return false
}

// RequestPermission handles incoming permission requests.
func (n *Node) RequestPermission(
	ctx context.Context,
	req *pb.Request,
) (*pb.Reply, error) {

	// Update Lamport clock after receiving the request.
	n.receiveTimestamp(req.Timestamp)

	n.Mu.Lock()

	// Decide whether the reply must be deferred.
	deferReply := false

	if n.InCriticalSection {
		deferReply = true
	} else if n.Requesting && n.hasPriority(
		req.Timestamp,
		req.NodeId,
	) {
		deferReply = true
	}

	if deferReply {
		if n.DeferredReplies == nil {
			n.DeferredReplies = make(map[string]bool)
		}

		if n.DeferredReplyChans == nil {
			n.DeferredReplyChans = make(map[string]chan struct{})
		}

		// Create a channel to wait for the deferred reply.
		replyChan := make(chan struct{})

		n.DeferredReplies[req.NodeId] = true
		n.DeferredReplyChans[req.NodeId] = replyChan

		log.Printf(
			"Node %s defers REPLY to Node %s",
			n.ID,
			req.NodeId,
		)

		n.Mu.Unlock()

		// Wait until permission is granted or the request is cancelled.
		select {
		case <-replyChan:
			log.Printf(
				"Node %s sends deferred REPLY to Node %s",
				n.ID,
				req.NodeId,
			)

			// Increment the clock and attach the timestamp to the reply.
			timestamp := n.sendTimestamp()

			return &pb.Reply{
				NodeId:    n.ID,
				Timestamp: timestamp,
			}, nil

		case <-ctx.Done():

			// Remove the cancelled request if it is still pending.
			n.Mu.Lock()

			if n.DeferredReplyChans[req.NodeId] == replyChan {
				delete(n.DeferredReplyChans, req.NodeId)
				delete(n.DeferredReplies, req.NodeId)
			}

			n.Mu.Unlock()

			return nil, ctx.Err()
		}
	}

	n.Mu.Unlock()

	// Grant permission immediately.
	log.Printf(
		"Node %s grants permission to Node %s",
		n.ID,
		req.NodeId,
	)

	// Increment the clock and attach the timestamp to the reply.
	timestamp := n.sendTimestamp()

	return &pb.Reply{
		NodeId:    n.ID,
		Timestamp: timestamp,
	}, nil
}

// StartServer starts the grpc server for this node.
func (n *Node) StartServer() error {
	listener, err := net.Listen("tcp", n.Address)
	if err != nil {
		return err
	}

	n.Server = grpc.NewServer()

	pb.RegisterMutualExclusionServer(n.Server, n)

	go func() {
		if err := n.Server.Serve(listener); err != nil {
			log.Printf(
				"Node %s server error: %v",
				n.ID,
				err,
			)
		}
	}()

	log.Printf(
		"Node %s listening on %s",
		n.ID,
		n.Address,
	)

	return nil
}

// ConnectToPeers creates grpc clients to all other nodes.
func (n *Node) ConnectToPeers() error {
	n.Clients = make(map[string]pb.MutualExclusionClient)
	n.Connections = make(map[string]*grpc.ClientConn)

	for peerID, address := range n.Peers {
		if peerID == n.ID {
			continue
		}

		conn, err := grpc.NewClient(
			address,
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)

		if err != nil {
			n.CloseConnections()

			return fmt.Errorf(
				"node %s could not connect to node %s: %w",
				n.ID,
				peerID,
				err,
			)
		}

		n.Connections[peerID] = conn
		n.Clients[peerID] = pb.NewMutualExclusionClient(conn)
	}

	return nil
}

// CloseConnections closes all outgoing grpc connections.
func (n *Node) CloseConnections() {
	for peerID, conn := range n.Connections {
		if err := conn.Close(); err != nil {
			log.Printf(
				"Node %s could not close connection to Node %s: %v",
				n.ID,
				peerID,
				err,
			)
		}

		delete(n.Connections, peerID)
		delete(n.Clients, peerID)
	}
}

// RequestCriticalSection requests permission from all other nodes.
func (n *Node) RequestCriticalSection(ctx context.Context) error {
	n.Mu.Lock()

	if n.Requesting || n.InCriticalSection {
		n.Mu.Unlock()

		return fmt.Errorf(
			"node %s already has a pending request or is in the critical section",
			n.ID,
		)
	}

	// Generate a new Lamport timestamp and mark the node as requesting.
	n.Clock++
	n.RequestTimestamp = n.Clock
	n.Requesting = true
	n.RepliesReceived = make(map[string]bool)

	timestamp := n.RequestTimestamp

	n.Mu.Unlock()

	log.Printf(
		"Node %s requests the critical section (timestamp: %d)",
		n.ID,
		timestamp,
	)

	// Create a cancellable context for all outgoing requests.
	requestCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Send request messages to all peers concurrently.
	var wg sync.WaitGroup
	errChan := make(chan error, len(n.Clients))

	for peerID, client := range n.Clients {
		wg.Add(1)

		go func(peerID string, client pb.MutualExclusionClient) {
			defer wg.Done()

			reply, err := client.RequestPermission(
				requestCtx,
				&pb.Request{
					NodeId:    n.ID,
					Timestamp: timestamp,
				},
			)

			if err != nil {
				errChan <- fmt.Errorf(
					"node %s could not get permission from node %s: %w",
					n.ID,
					peerID,
					err,
				)

				// Cancel the other pending requests.
				cancel()
				return
			}

			// Update Lamport clock after receiving the reply.
			n.receiveTimestamp(reply.Timestamp)

			n.Mu.Lock()
			n.RepliesReceived[reply.NodeId] = true
			n.Mu.Unlock()

			log.Printf(
				"Node %s received REPLY from Node %s",
				n.ID,
				reply.NodeId,
			)

		}(peerID, client)
	}

	wg.Wait()
	close(errChan)

	// Check whether any request failed.
	for err := range errChan {
		n.Mu.Lock()
		n.Requesting = false
		n.RepliesReceived = make(map[string]bool)
		n.Mu.Unlock()

		return err
	}

	// Enter the critical section after receiving all permissions.
	n.Mu.Lock()
	defer n.Mu.Unlock()

	if len(n.RepliesReceived) != len(n.Clients) {
		n.Requesting = false
		n.RepliesReceived = make(map[string]bool)

		return fmt.Errorf(
			"node %s did not receive all permissions",
			n.ID,
		)
	}

	n.Requesting = false
	n.InCriticalSection = true

	log.Printf(
		"Node %s ENTERS the critical section",
		n.ID,
	)

	return nil
}

// ReleaseCriticalSection releases the critical section and sends all deferred replies.
func (n *Node) ReleaseCriticalSection() error {
	n.Mu.Lock()

	if !n.InCriticalSection {
		n.Mu.Unlock()

		return fmt.Errorf(
			"node %s is not in the critical section",
			n.ID,
		)
	}

	// Mark the node as no longer using the critical section.
	n.InCriticalSection = false
	n.Clock++

	log.Printf(
		"Node %s EXITS the critical section",
		n.ID,
	)

	// Collect all deferred reply channels.
	var channels []chan struct{}

	for peerID, replyChan := range n.DeferredReplyChans {
		channels = append(channels, replyChan)

		delete(n.DeferredReplyChans, peerID)
		delete(n.DeferredReplies, peerID)
	}

	n.Mu.Unlock()

	// Release deferred rpc handlers.
	for _, replyChan := range channels {
		close(replyChan)
	}

	return nil
}
