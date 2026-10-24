package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "04/grpc"
)

func main() {
	conn, _ := grpc.NewClient(
		"localhost:5050",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := pb.NewITUServiceClient(conn)
	students, _ := client.GetStudents(context.Background(), &pb.Nothing{})

	fmt.Println(students)
}