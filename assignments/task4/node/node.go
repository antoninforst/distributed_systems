package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	pb "task4/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const RequestTimeout = 10 * time.Second
const RetryDelay = 500 * time.Millisecond

type Request struct {
	Timestamp int
	NodeID    int
}

type DeferredRequest struct {
	Request Request
	WaitCh  chan struct{}
}

type Node struct {
	ID      int
	Address int
	Peers   []int

	mu          sync.Mutex
	clock       int
	myRequest   Request
	requesting  bool
	backlog     []DeferredRequest
}

func NewNode(id, address int, peers []int) *Node {
	return &Node{
		ID:         id,
		Address:    address,
		Peers:      peers,
		clock:      0,
		requesting: false,
	}
}

func NewRequest(timestamp int, nodeID int) Request {
	return Request{Timestamp: timestamp, NodeID: nodeID}
}

func (n *Node) logEvent(event string) {
	fmt.Printf("[%d at %d] %s\n", n.ID, n.clock, event)
}

func (n *Node) RequestCriticalSection() {
	n.mu.Lock()

	if n.requesting {
		n.mu.Unlock()
		return
	}

	n.clock++
	n.myRequest = NewRequest(n.clock, n.ID)
	n.requesting = true
	requestTimestamp := n.myRequest.Timestamp
	n.mu.Unlock()

	n.logEvent(fmt.Sprintf("Request (t=%d)", requestTimestamp))
	n.SendRequestToPeers(n.myRequest)
}

func (n *Node) SendRequestToPeers(request Request) {
	type peerResult struct {
		peer int
		err  error
	}
	peers := append([]int(nil), n.Peers...)

	if len(peers) == 0 {
		n.EnterCriticalSection()
		return
	}

	pending := make(map[int]struct{}, len(peers))
	for _, peer := range peers {
		pending[peer] = struct{}{}
	}

	for len(pending) > 0 {
		var wg sync.WaitGroup
		results := make(chan peerResult, len(pending))

		for peerAddress := range pending {
			peerAddress := peerAddress
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := n.sendRequestToPeer(request, peerAddress)
				results <- peerResult{peer: peerAddress, err: err}
			}()
		}

		wg.Wait()
		close(results)

		nextPending := make(map[int]struct{})
		for result := range results {
			if result.err != nil {
				n.logEvent(fmt.Sprintf("Peer %d unavailable, retrying", result.peer))
				nextPending[result.peer] = struct{}{}
			}
		}

		pending = nextPending
		if len(pending) > 0 {
			time.Sleep(RetryDelay)
		}
	}

	n.EnterCriticalSection()
}

func (n *Node) EnterCriticalSection() {
	n.mu.Lock()
	n.clock++
	entryTime := n.clock
	n.mu.Unlock()

	n.logEvent(fmt.Sprintf("ENTER critical section (request t=%d)", n.myRequest.Timestamp))

	time.Sleep(500 * time.Millisecond)

	n.mu.Lock()
	n.clock++
	exitTime := n.clock

	n.requesting = false
	deferred := append([]DeferredRequest(nil), n.backlog...)
	n.backlog = nil
	n.mu.Unlock()

	n.logEvent(fmt.Sprintf("EXIT critical section (entered at %d, exited at %d)", entryTime, exitTime))

	for _, deferredRequest := range deferred {
		close(deferredRequest.WaitCh)
		n.logEvent(fmt.Sprintf(
			"Backlog release: grant deferred request from node %d (t=%d)",
			deferredRequest.Request.NodeID,
			deferredRequest.Request.Timestamp,
		))
	}
}

func (n *Node) RequestPermission(ctx context.Context, incoming *pb.Request) (*pb.Reply, error) {
	incomingNodeID, err := strconv.Atoi(incoming.NodeId)
	if err != nil {
		return nil, fmt.Errorf("invalid requester node id %q: %w", incoming.NodeId, err)
	}

	incomingRequest := Request{
		Timestamp: int(incoming.Timestamp),
		NodeID:    incomingNodeID,
	}

	n.mu.Lock()
	if incomingRequest.Timestamp > n.clock {
		n.clock = incomingRequest.Timestamp
	}
	n.clock++

	deferReply := false
	var waitCh chan struct{}

	if n.requesting && n.hasOwnRequestPriorityOver(incomingRequest) {
		deferReply = true
		waitCh = make(chan struct{})
		n.backlog = append(n.backlog, DeferredRequest{Request: incomingRequest, WaitCh: waitCh})
		n.logEvent(fmt.Sprintf(
			"DEFER reply to node %d (incoming t=%d, own t=%d)",
			incomingRequest.NodeID,
			incomingRequest.Timestamp,
			n.myRequest.Timestamp,
		))
	} else {
		n.logEvent(fmt.Sprintf(
			"GRANT to node %d (incoming t=%d)",
			incomingRequest.NodeID,
			incomingRequest.Timestamp,
		))
	}
	n.mu.Unlock()

	if deferReply {
		select {
		case <-waitCh:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	n.mu.Lock()
	n.clock++
	currentClock := n.clock
	n.mu.Unlock()

	return &pb.Reply{
		NodeId:    strconv.Itoa(n.ID),
		Timestamp: uint64(currentClock),
	}, nil
}

func (n *Node) hasOwnRequestPriorityOver(incoming Request) bool {
	if n.myRequest.Timestamp < incoming.Timestamp {
		return true
	}
	if n.myRequest.Timestamp > incoming.Timestamp {
		return false
	}

	return n.myRequest.NodeID < incoming.NodeID
}

func (n *Node) sendRequestToPeer(request Request, peerAddress int) error {
	target := fmt.Sprintf("localhost:%d", peerAddress)

	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("dial %s: %w", target, err)
	}
	defer conn.Close()

	client := pb.NewMutualExclusionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	grpcRequest := &pb.Request{
		NodeId:    strconv.Itoa(request.NodeID),
		Timestamp: uint64(request.Timestamp),
	}

	grpcReply, err := client.RequestPermission(ctx, grpcRequest)
	if err != nil {
		return fmt.Errorf("request permission from %s: %w", target, err)
	}

	replyTimestamp := int(grpcReply.Timestamp)

	n.mu.Lock()
	if replyTimestamp > n.clock {
		n.clock = replyTimestamp
	}
	n.clock++
	n.mu.Unlock()

	return nil
}
