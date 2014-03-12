package main

import (
	"fmt"
	
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/briee"
	bl "github.com/maxnordlund/breamio/beenleigh"
)

const (
	Company = "Bream IO"
	Product = "Eriver"
	Version = "v2.0"
)

func main() {
	fmt.Println("Welcome to", Company, Product, Version)
	logic := bl.New(briee.New, aioli.New())
	logic.ListenAndServe()
	fmt.Println("Thank you for using our product.")
}