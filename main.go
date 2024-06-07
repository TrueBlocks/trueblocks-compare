package main

import (
	"flag"
	"fmt"
	"log"
)

var min, max = 0, 5000

var reuseDatabase bool
var dbFileName string

func init() {
	flag.BoolVar(&reuseDatabase, "reuse", false, "reuse database")
	flag.Parse()

	if reuseDatabase {
		if flag.Arg(0) == "" {
			log.Fatalln("--reuse requires database file name")
		}
		dbFileName = flag.Arg(0)
	}
}

func main() {
	// Setup
	// DetectProviders
	// DownloadAppearances
	// Results
	// Report

	comparison := Setup(dbFileName)
	if !reuseDatabase {
		if err := comparison.DownloadAppearances(); err != nil {
			log.Fatalln(err)
		}
	}

	results, err := comparison.Results()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Addresses (count < %d):\t%d \n", comparison.maxAppearances, results.AddressCount)
	for _, provider := range comparison.Providers {
		fmt.Printf("Found by %s:\t%d addresses,\t%d appearances\n", provider, results.AddressesBy[provider], results.AppearancesBy[provider])
	}
	for _, provider := range comparison.Providers {
		fmt.Printf("Found only by %s:\t%d addresses,\t %d appearances\n", provider, results.AddressesOnlyBy[provider], results.AppearancesOnlyBy[provider])
		for _, groupedReasons := range results.GroupedReasons[provider] {
			fmt.Printf("\t%s: %d\n", groupedReasons.Reason, groupedReasons.Count)
		}
		fmt.Println("\tBalance changes:", results.BalanceChanges[provider])
	}
}
