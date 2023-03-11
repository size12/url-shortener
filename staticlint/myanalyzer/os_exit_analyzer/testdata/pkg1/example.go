package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, world!")
	os.Exit(0) // want "you shouldn't use os.Exit in function main"
}
