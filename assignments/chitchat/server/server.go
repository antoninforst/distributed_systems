package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	chatpb "chitchat/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	chatpb.UnimplementedChitChatServiceServer

	mu          sync.Mutex
	clients     map[string]*participant
	nextID      int
	logicalTime uint64
}

type participant struct {
	name string

	mu     sync.Mutex
	queue  []*chatpb.ChatEvent
	notify chan struct{}
	done   chan struct{}
	closed bool
}

func newParticipant(name string) *participant {
	return &participant{
		name:   name,
		queue:  make([]*chatpb.ChatEvent, 0),
		notify: make(chan struct{}, 1),
		done:   make(chan struct{}),
	}
}

// enqueue adds an event to the participant's queue without
// waiting for the client to receive it.
func (p *participant) enqueue(event *chatpb.ChatEvent) bool {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return false
	}

	p.queue = append(p.queue, event)
	p.mu.Unlock()

	select {
	case p.notify <- struct{}{}:
	default:
		// A notification is already pending.
	}

	return true
}

// nextEvent returns the next queued event.
// It returns false when the participant has closed and its queue is empty.
func (p *participant) nextEvent() (*chatpb.ChatEvent, bool) {
	for {
		p.mu.Lock()

		if len(p.queue) > 0 {
			event := p.queue[0]
			p.queue[0] = nil
			p.queue = p.queue[1:]
			p.mu.Unlock()
			return event, true
		}

		if p.closed {
			p.mu.Unlock()
			return nil, false
		}

		p.mu.Unlock()

		select {
		case <-p.notify:
		case <-p.done:
		}
	}
}

// stop closes the participant's notification channel exactly once.
func (p *participant) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.closed {
		p.closed = true
		close(p.done)
	}
}

func (s *Server) Join(
	request *chatpb.JoinRequest,
	stream grpc.ServerStreamingServer[chatpb.ChatEvent],
) error {
	name := strings.TrimSpace(request.GetDisplayName())
	if name == "" {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}

	s.mu.Lock()

	s.nextID++
	clientID := fmt.Sprintf("client-%d", s.nextID)
	p := newParticipant(name)
	s.clients[clientID] = p

	s.logicalTime++
	event := &chatpb.ChatEvent{
		Content: fmt.Sprintf(
			"Participant %s joined Chit Chat at logical time %d",
			name,
			s.logicalTime,
		),
		LogicalTime: s.logicalTime,
		ClientId:    clientID,
	}

	s.broadcastLocked(event)
	logicalTime := s.logicalTime

	s.mu.Unlock()

	log.Printf(
		"component=Server event=connection client_id=%s name=%q",
		clientID,
		name,
	)
	log.Printf(
		"component=Server event=join client_id=%s name=%q logical_time=%d",
		clientID,
		name,
		logicalTime,
	)

	// Ensure that unexpected disconnections clean up the participant.
	defer s.disconnect(clientID)

	for {
		event, ok := p.nextEvent()
		if !ok {
			return nil
		}

		if err := stream.Send(event); err != nil {
			log.Printf(
				"component=Server event=delivery_error client_id=%s logical_time=%d error=%q",
				clientID,
				event.GetLogicalTime(),
				err,
			)
			return err
		}

		log.Printf(
			"component=Server event=delivery client_id=%s logical_time=%d",
			clientID,
			event.GetLogicalTime(),
		)
	}
}

func (s *Server) Publish(
	_ context.Context,
	request *chatpb.PublishRequest,
) (*chatpb.PublishReply, error) {
	content := strings.TrimSpace(request.GetContent())

	if content == "" {
		return nil, status.Error(codes.InvalidArgument, "message cannot be empty")
	}

	if len([]rune(content)) > 128 {
		return nil, status.Error(codes.InvalidArgument, "message cannot exceed 128 characters")
	}

	s.mu.Lock()

	p, exists := s.clients[request.GetClientId()]
	if !exists {
		s.mu.Unlock()
		return nil, status.Errorf(
			codes.NotFound,
			"unknown client ID %q",
			request.GetClientId(),
		)
	}

	s.logicalTime++

	event := &chatpb.ChatEvent{
		Content:     fmt.Sprintf("%s: %s", p.name, content),
		LogicalTime: s.logicalTime,
		ClientId:    request.GetClientId(),
	}

	s.broadcastLocked(event)
	logicalTime := s.logicalTime

	s.mu.Unlock()

	log.Printf(
		"component=Server event=publish client_id=%s logical_time=%d",
		request.GetClientId(),
		logicalTime,
	)

	return &chatpb.PublishReply{}, nil
}

func (s *Server) Leave(
	_ context.Context,
	request *chatpb.LeaveRequest,
) (*chatpb.LeaveReply, error) {
	if !s.disconnect(request.GetClientId()) {
		return nil, status.Errorf(
			codes.NotFound,
			"unknown client ID %q",
			request.GetClientId(),
		)
	}

	return &chatpb.LeaveReply{}, nil
}

// disconnect removes a participant and broadcasts the leave event
// to all participants who remain connected.
func (s *Server) disconnect(clientID string) bool {
	s.mu.Lock()

	p, exists := s.clients[clientID]
	if !exists {
		s.mu.Unlock()
		return false
	}

	delete(s.clients, clientID)

	s.logicalTime++
	event := &chatpb.ChatEvent{
		Content: fmt.Sprintf(
			"Participant %s left Chit Chat at logical time %d",
			p.name,
			s.logicalTime,
		),
		LogicalTime: s.logicalTime,
		ClientId:    clientID,
	}

	s.broadcastLocked(event)
	logicalTime := s.logicalTime

	p.stop()

	s.mu.Unlock()

	log.Printf(
		"component=Server event=disconnection client_id=%s name=%q",
		clientID,
		p.name,
	)
	log.Printf(
		"component=Server event=leave client_id=%s name=%q logical_time=%d",
		clientID,
		p.name,
		logicalTime,
	)

	return true
}

// broadcastLocked queues an event for every currently connected participant.
// The caller must hold s.mu.
func (s *Server) broadcastLocked(event *chatpb.ChatEvent) {
	for clientID, p := range s.clients {
		if !p.enqueue(event) {
			log.Printf(
				"component=Server event=queue_closed client_id=%s logical_time=%d",
				clientID,
				event.GetLogicalTime(),
			)
		}
	}
}

func main() {
	address := "localhost:5050"

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf(
			"component=Server event=startup_error address=%s error=%q",
			address,
			err,
		)
	}

	server := &Server{
		clients: make(map[string]*participant),
	}

	grpcServer := grpc.NewServer()
	chatpb.RegisterChitChatServiceServer(grpcServer, server)

	log.Printf("component=Server event=startup address=%s", address)

	// Listen for Ctrl+C and shut down gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	go func() {
		<-stop
		log.Printf("component=Server event=shutdown")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.Printf("component=Server event=serve_error error=%q", err)
	}
}
