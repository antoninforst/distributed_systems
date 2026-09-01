package main

import "fmt"

func r1(source chan string, target chan string) {
	for {
		target <- "ping"
		fmt.Println(<-source)
	}
}
func r2(source chan string, target chan string) {
	for {
		fmt.Println(<-source)
		target <- "pong"
	}
}

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)
	go r1(ch1, ch2)
	r2(ch2, ch1)
}
