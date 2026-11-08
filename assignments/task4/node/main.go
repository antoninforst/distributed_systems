package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "task4/grpc"

	"google.golang.org/grpc"
)

const PEERCOUNT = 3
const BASEPORT = 5000

func main() {
	var id int
	_, err := fmt.Scanln(&id)

	if err != nil || id < 1 || id > PEERCOUNT {
		log.Fatalf("Invalid input. Please enter a number between 1 and %d.", PEERCOUNT)
	}

	address := BASEPORT + id

	//generate peers
	peers := []int{}
	for i := 1; i <= PEERCOUNT; i++ {
		if i != id {
			peers = append(peers, BASEPORT+i)
		}
	}

	node := NewNode(id, address, peers)

	// server setup
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", address))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", address, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMutualExclusionServiceServer(grpcServer, node)

	go func() {
		if serveErr := grpcServer.Serve(listener); serveErr != nil {
			log.Fatalf("gRPC server error: %v", serveErr)
		}
	}()

	go func() {
		time.Sleep(time.Duration(id) * time.Second)
		node.RequestCriticalSection()
	}()

	// Keep the node alive
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Printf("Node %d shutting down", id)
	grpcServer.GracefulStop()
}
