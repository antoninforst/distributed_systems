package main

import (
	"fmt"
)

const PhilosophersCount = 5

func philosopher(id int, leftCh, rightCh chan string) {
	if id == 0 {
		rightCh <- "Hello"
		fmt.Printf("P0 Hello\n")
		<- leftCh
		leftCh <- "Olleh"
		fmt.Printf("P0 Olleh\n")
		<- rightCh
	} else {
		for i := 0; i < 2; i++ {
			select {
			case msg := <- leftCh:
				fmt.Printf("P%d: %s\n", id, msg)
				rightCh <- msg
			case msg := <- rightCh:
				fmt.Printf("P%d: %s\n", id, msg)
				leftCh <- msg
			}
		}
	}
}

func fork(id int, leftCh, rightCh chan string) {
	for i := 0; i < 2; i++ {
		select {
		case msg := <- leftCh:
			fmt.Printf("F%d: %s\n", id, msg)
			rightCh <- msg
		case msg := <- rightCh:
			fmt.Printf("F%d: %s\n", id, msg)
			leftCh <- msg
		}
	}
}

func main() {
	firstCh := make(chan string, 1)
	sharedCh := firstCh
	for i := 0; i < PhilosophersCount; i++ {
		middleCh := make(chan string, 1)
		var newCh chan string
		if i+1 == PhilosophersCount {
			newCh = firstCh
		} else {
			newCh = make(chan string, 1)
		}
		go philosopher(i, sharedCh, middleCh)
		go fork(i, middleCh, newCh)
		sharedCh = newCh
	}
	select {}
}