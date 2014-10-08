package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	pflag "github.com/ogier/pflag"
)

var (
	// These variables will be injected during build process
	fingerprint string
	gitSHA      string
	// Flag define here
	versionFlag bool
)

const (
	Company = "Bream IO"
	Product = "EyeStream"
	Version = "v0.9"
)

func main() {
	pflag.Parse()
	if versionFlag {

		fmt.Printf("Product: %s\nVersion number: %s\nCompany: %s\n", Product, Version, Company)
		if fingerprint != "" {
			fmt.Printf("Fingerprint: %s\n", fingerprint)
		}
		if gitSHA != "" {
			fmt.Printf("gitSHA: %s\n", gitSHA)
		}
		return //if version flag is true then we don't want to run the program
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	fmt.Println("Welcome to", Company, Product, Version)
	logic := bl.New(briee.New)

	go func() {
		<-done
		logic.Close()
	}()

	logic.ListenAndServe()
	fmt.Println("Thank you for using our product.")
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	pflag.BoolVarP(&versionFlag, "version", "v", false, "Enable this flag to print version information")
}
