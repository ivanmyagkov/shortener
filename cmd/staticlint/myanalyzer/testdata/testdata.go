package main

import (
	"fmt"
	"os"
)

func test() {
	fmt.Println("exit in goroutine.")
	os.Exit(0)
}

func main() {
	// should not raise a flag in my analyzer
	go test()
	// raise a flag in my analyzer
	fmt.Println("exit from main().")
	os.Exit(0) // want "os.Exit exists in main body"
}
