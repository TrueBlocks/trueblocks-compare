package main

import (
	"flag"
	"log"
	"os"
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

	_ = results.Csv(os.Stdout)

	results.Text()
}
