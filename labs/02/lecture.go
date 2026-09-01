package main

import (
	"fmt"
	"os"
	"strconv"
)

func printSlice(slice []int, printContent bool) {
	if printContent {
		fmt.Printf("len=%d cap=%d %v\n", len(slice), cap(slice), slice)
	} else {
		fmt.Printf("len=%d cap=%d\n", len(slice), cap(slice))
	}
}

func main() {
	input := os.Args
	if len(input) < 2 {
		fmt.Println("Please provide an integer argument.")
		return
	}
	n, _ := strconv.Atoi(input[1])
	slice := make([]int, n)
	printContent := n < 40
	printSlice(slice, printContent)
	slice = append(slice, 1)
	printSlice(slice, printContent)
}