package main

import "fmt"

func first() {
	fmt.Println("I am first")
}

func second() {
	fmt.Println("I am second")
}

func main()  {
	go first()
	go second()
}
