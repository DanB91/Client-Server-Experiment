package main

import (
	"fmt"
	"os"
)

func main() {
	println("Hello!!!")
}

func println(f string, args ...any) {
	fmt.Printf(f+"\n", args...)
}

func fatal(f string, args ...any) {
	println(f, args...)
	os.Exit(1)
}
