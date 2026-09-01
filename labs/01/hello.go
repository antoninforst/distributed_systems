package main

import (
	"fmt"
)

func printWithSeparator(word string, n int, separator string, suffix string) {
	if n > 0 {
		fmt.Print(word)
		for i := 1; i < n; i++ {
			fmt.Print(separator + word)
		}
	}

	fmt.Print(suffix)
}

func main() {
	var word string

	fmt.Scan(&word)

	printWithSeparator(word, 3, " ", "")
}
