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

	log.Printf(
		"component=Client event=connection client_id=%s name=%q",
		clientID,
		name,
	)

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

	fmt.Println("You are in Chit Chat.")
	fmt.Println("Type a message and press Enter to send it.")
	fmt.Println("Type /quit to leave the chat.")

	for {
		fmt.Print("> ")
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			break
		}

		message = strings.TrimSpace(message)

		if message == "/quit" {
			break
		}

		if message == "" {
			continue
		}

		if len([]rune(message)) > 128 {
			fmt.Println("Message cannot exceed 128 characters.")
			continue
		}

		_, err = client.Publish(context.Background(), &chatpb.PublishRequest{
			ClientId: clientID,
			Content:  message,
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
			continue
		}
	}

	_, err = client.Leave(context.Background(), &chatpb.LeaveRequest{
		ClientId: clientID,
	})
	if err != nil {
		log.Printf("Error leaving chat: %v", err)
	}

	log.Printf(
		"component=Client event=disconnection client_id=%s name=%q",
		clientID,
		name,
	)

	fmt.Println("You left Chit Chat.")
}

func printEvent(event *chatpb.ChatEvent) {
	fmt.Printf("[%d] %s\n", event.GetLogicalTime(), event.GetContent())

	log.Printf(
		"component=Client event=received client_id=%s logical_time=%d content=%q",
		event.GetClientId(),
		event.GetLogicalTime(),
		event.GetContent(),
	)
}
