package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	globalTime "github.com/maxnordlund/breamio/globalTime"
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
	Version     = "v0.9"
	evalEndDate = "5 Jan 2015"
	evalLayout  = "2 Jan 2006"
)

func main() {
	err := checkEvalPeriod()
	if err != nil {
		log.Println("Failed to check evaluation date, please verify your internet connection")
		log.Println("Now exiting program")
		os.Exit(0)
	}
	go func() {
		for _ = range time.Tick(24 * time.Hour) {
			err := checkEvalPeriod()
			if err != nil { //Try 5 more times or exit the program
				it := 0
				for _ = range time.Tick(4 * time.Minute) {
					it++
					err = checkEvalPeriod()
					if err == nil {
						break
					}
					if it > 4 {
						log.Println("Failed to check evaluation date, please verify your internet connection")
						log.Println("Now exiting program")
						os.Exit(0)
					}
				}
			}
		}
	}()
	flag.Parse()
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

	checkEvalPeriod() // We want to be able to print version information before quitting because of ended eval period

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
	flag.BoolVar(&versionFlag, "version", false, "Enable this flag to print version information")
	flag.BoolVar(&versionFlag, "v", false, "Enable this flag to print version information")
}

func checkEvalPeriod() error {
	// Evaluation period check
	googleTime, err := globalTime.GetGoogleTime()
	if err != nil {
		return err
	}
	endDate, err := time.Parse(evalLayout, evalEndDate)
	if err != nil {
		return err
	}
	if googleTime.After(endDate) {
		log.Println("Evaluation period is over. It ended", evalEndDate)
		os.Exit(0)
	}
	return nil
}
