package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	chatpb "chitchat/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Your name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	name = strings.TrimSpace(name)

	connection, err := grpc.NewClient(
		"localhost:5050",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	client := chatpb.NewChitChatServiceClient(connection)
	stream, err := client.Join(context.Background(), &chatpb.JoinRequest{
		DisplayName: name,
	})
	if err != nil {
		log.Fatal(err)
	}

	firstEvent, err := stream.Recv()
	if err != nil {
		log.Fatal(err)
	}
	clientID := firstEvent.GetClientId()
	printEvent(firstEvent)

	go func() {
		for {
			event, err := stream.Recv()
			if err != nil {
				return
			}
			printEvent(event)
		}
	}()

	fmt.Println("You are in Chit Chat. Press Enter to leave.")
	reader.ReadString('\n')

	_, err = client.Leave(context.Background(), &chatpb.LeaveRequest{
		ClientId: clientID,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You left Chit Chat.")
}

func printEvent(event *chatpb.ChatEvent) {
	fmt.Printf("[%d] %s\n", event.GetLogicalTime(), event.GetContent())
	log.Printf("component=Client event=received logical_time=%d content=%q", event.GetLogicalTime(), event.GetContent())
}
