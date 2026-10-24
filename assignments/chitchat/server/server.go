package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	chatpb "chitchat/grpc"
	"google.golang.org/grpc"
)

type Server struct {
	chatpb.UnimplementedChitChatServiceServer

	mu          sync.Mutex
	clients     map[string]*participant
	nextID      int
	logicalTime uint64
}

type participant struct {
	name   string
	events chan *chatpb.ChatEvent
}

func (s *Server) Join(request *chatpb.JoinRequest, stream grpc.ServerStreamingServer[chatpb.ChatEvent]) error {
	name := strings.TrimSpace(request.GetDisplayName())
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	s.mu.Lock()
	s.nextID++
	clientID := fmt.Sprintf("client-%d", s.nextID)
	clientEvents := make(chan *chatpb.ChatEvent, 10)
	s.clients[clientID] = &participant{name: name, events: clientEvents}

	s.logicalTime++
	event := &chatpb.ChatEvent{
		Content:     fmt.Sprintf("Participant %s joined Chit Chat at logical time %d", name, s.logicalTime),
		LogicalTime: s.logicalTime,
		ClientId:    clientID,
	}
	s.broadcastLocked(event)
	logicalTime := s.logicalTime
	s.mu.Unlock()

	log.Printf("component=Server event=join client_id=%s name=%q logical_time=%d", clientID, name, logicalTime)
	for event := range clientEvents {
		if err := stream.Send(event); err != nil {
			return err
		}
		log.Printf("component=Server event=delivery client_id=%s logical_time=%d", clientID, event.GetLogicalTime())
	}
	return nil
}

func (s *Server) Leave(_ context.Context, request *chatpb.LeaveRequest) (*chatpb.LeaveReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	participant, exists := s.clients[request.GetClientId()]
	if !exists {
		return nil, fmt.Errorf("unknown client ID %q", request.GetClientId())
	}
	delete(s.clients, request.GetClientId())
	close(participant.events)

	s.logicalTime++
	event := &chatpb.ChatEvent{
		Content:     fmt.Sprintf("Participant %s left Chit Chat at logical time %d", participant.name, s.logicalTime),
		LogicalTime: s.logicalTime,
		ClientId:    request.GetClientId(),
	}
	s.broadcastLocked(event)
	log.Printf("component=Server event=leave client_id=%s name=%q logical_time=%d", request.GetClientId(), participant.name, s.logicalTime)

	return &chatpb.LeaveReply{}, nil
}

func (s *Server) broadcastLocked(event *chatpb.ChatEvent) {
	for _, participant := range s.clients {
		participant.events <- event
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:5050")
	if err != nil {
		log.Fatal(err)
	}

	server := &Server{
		clients: make(map[string]*participant),
	}
	grpcServer := grpc.NewServer()
	chatpb.RegisterChitChatServiceServer(grpcServer, server)

	log.Println("component=Server event=startup address=localhost:5050")
	log.Fatal(grpcServer.Serve(listener))
}
