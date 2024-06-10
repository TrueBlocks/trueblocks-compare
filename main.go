package main

import (
	"flag"
	"log"
	"os"
)

var min, max = 0, 5000

var addressFilePath string
var reuseDatabaseFile string
var dbFileName string
var dataDir string

func init() {
	flag.StringVar(&reuseDatabaseFile, "reuse", "", "reuse database")
	flag.StringVar(&dataDir, "datadir", "", "directory to save the database to")
	flag.Parse()

	addressFilePath = flag.Arg(0)
}

func main() {
	// Setup
	// DetectProviders
	// DownloadAppearances
	// Results
	// Report

	comparison := Setup(addressFilePath, dataDir, dbFileName)
	if reuseDatabaseFile == "" {
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
