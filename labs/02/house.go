package main

import (
	"fmt"
	"time"
)

var Step = 1 * time.Second

func Person(name string, ch chan string) {
	for {
		fmt.Println(name, "waits for the hammer")
		<- ch
		fmt.Println(name, "got the hammer")
		time.Sleep(Step)
		fmt.Println(name, "let go of the hammer")
		ch <- "free"
		fmt.Println(name, "rests for a while")
		time.Sleep(Step)
	}
}

func main() {
	ch := make(chan string)
	go Person("A", ch)
	go Person("B", ch)
	ch <- "free"
	select {}
}