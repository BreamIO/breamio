package main

import (
	"fmt"
)

var (
	Version string
	GitSHA  string
)

func main() {
	fmt.Println("Hello World", Version, GitSHA)
}
