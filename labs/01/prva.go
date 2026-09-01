package main

import "fmt"

type Crossword struct {
	Width int
	Height int
	Data []string
}

func NewCrosswordFromIo() Crossword {
	var c Crossword
	fmt.Scan(&c.Width, &c.Height)
	c.Data = make([]string, c.Height)
	fmt.Println(c.Height)
	for i := 0; i < c.Height; i++ {
		fmt.Scan(&c.Data[i])
	}
	return c
}

func isSmaller(a *string, b *string) bool {
	

func (c *Crossword) Find() string {

	return c.Data[0]
}

func main() {
	crossword := NewCrosswordFromIo()

	fmt.Printf("Found: %s\n", crossword.Find())
}