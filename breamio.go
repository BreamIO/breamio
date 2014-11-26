package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	licence "github.com/maxnordlund/breamio/licence"
)

var (
	// These variables will be injected during build process
	fingerprint string
	gitSHA      string
	// Flag define here
	versionFlag bool
)

const (
	Company     = "Bream IO"
	Product     = "EyeStream"
	Version     = "v1.3-alpha"
	evalEndDate = "1 Apr 2015"
	evalLayout  = "2 Jan 2006"
)

func main() {

	flag.Parse()
	if versionFlag {
		printVersionInfo()
		return // We do not want to run the rest of the program if version flag is set
	}

	defer func() {
		fmt.Println("Thank you for using our product.")
	}()

	// We want to be able to print version information before quitting because of ended eval period
	// Start a routine that will check that we do not pass
	// evaluation date during runtime.
	go func() {
		licence.RepeatedlyCheckEvalTime(evalLayout, evalEndDate)
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	fmt.Println("Welcome to", Company, Product, Version)
	logic := bl.New(briee.New)

	go func() {
		<-done
		signal.Stop(done)
		logic.Close()
		logic.Logger().Println("Closed")
	}()

	logic.ListenAndServe()
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.BoolVar(&versionFlag, "version", false, "Enable this flag to print version information")
	flag.BoolVar(&versionFlag, "v", false, "Enable this flag to print version information")
}

func printVersionInfo() {
	fmt.Printf("Product: %s\nVersion number: %s\nCompany: %s\n", Product, Version, Company)
	if fingerprint != "" {
		fmt.Printf("Fingerprint: %s\n", fingerprint)
	}
	if gitSHA != "" {
		fmt.Printf("gitSHA: %s\n", gitSHA)
	}
}
