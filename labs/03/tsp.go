package main

import (
	"fmt"
	"strings"
)

type Message struct {
	SYN      bool
	ACK      bool
	SenderSeq    int
	ReceiverSeq int
	Data string
}

func (m Message) String() string {
	var parts []string

	if m.SYN {
		parts = append(parts, "SYN")
	} else {
		parts = append(parts, "   ")
	}

	if m.ACK {
		parts = append(parts, "ACK")
	} else {
		parts = append(parts, "   ")
	}

	if m.SenderSeq != -1 {
		parts = append(parts, fmt.Sprintf("%d", m.SenderSeq))
	}

	if m.ReceiverSeq != -1 {
		parts = append(parts, fmt.Sprintf("%d", m.ReceiverSeq))
	}

	if m.Data != "" {
		parts = append(parts, fmt.Sprintf("DATA: '%s'", m.Data))
	}	

	return strings.Join(parts, " ")
}

func log(isServer bool, isSending bool, message Message) {
	who := "SER"
	if !isServer {
		who = "CLI"
	}
	what := "Received:"
	if isSending {
		what = "Sent    :"
	}
	println("[", who, "]", what, message.String())
}

func err(isServer bool) {
	who := "SER"
	if !isServer {
		who = "CLI"
	}
	println("[", who, "]", "Ivalid response, closing connection.")
}

const SEQ_GAP = 200
const CLIENT_SEQ = 100
const INVALID_SEQ = -1

func server(serverChan chan<- Message, clientChan <-chan Message) {
	request := <-clientChan
	log(true, false, request)
	
	message := Message{
		SYN: true,
		ACK: true,
		SenderSeq: request.SenderSeq + SEQ_GAP,
		ReceiverSeq: request.SenderSeq + 1}
	serverChan <- message
	if message.ACK && message.ReceiverSeq == request.SenderSeq+1 {
		log(true, true, message)
	} else {
		// Never happens in this hand in
		err(false)
		return
	}
	ack := <-clientChan
	log(true, false, ack)
}

func client(serverChan <-chan Message, clientChan chan<- Message) {
	request := Message{
		SYN: true,
		SenderSeq: CLIENT_SEQ,
		ReceiverSeq: INVALID_SEQ}
	clientChan <- request
	log(false, true, request)

	response := <-serverChan
	log(false, false, response)

	if response.ACK && response.ReceiverSeq == CLIENT_SEQ+1 {
		message := Message{
			ACK: true,
			SenderSeq: response.ReceiverSeq,
			ReceiverSeq: response.SenderSeq + 1,
			Data: "GET /index.html",
		}
		clientChan <- message
		log(false, true, message)
	} else {
		// Never happens in this hand in
		err(false)
		return
	}
}

func main() {
	serverChan := make(chan Message)
	clientChan := make(chan Message)
	go client(serverChan, clientChan)
	server(serverChan, clientChan)
}