package main

import (
	pb "04/grpc"
	"context"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedITUServiceServer
	students []*pb.Student
}

func (s *Server) GetStudents(ctx context.Context, req *pb.Nothing) (*pb.Students, error) {
	students := pb.Students {
		Students: s.students,
	}

	return &students, nil
}

func main() {
	server := Server{
		students: []*pb.Student{&pb.Student{Name: "bob", Age: 19}},
	}
	listener, _ := net.Listen("tcp", "localhost:5050")
	grpcServer := grpc.NewServer()

	pb.RegisterITUServiceServer(grpcServer, &server)

	grpcServer.Serve(listener)

}


