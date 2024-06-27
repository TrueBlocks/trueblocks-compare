package main

import (
	"flag"
	"log"
	"os"
)

var minAppearances, maxAppearances = 0, 5000

var addressFilePath string
var reuseDatabaseFile string
var dataDir string
var format string

func init() {
	// Parse command line options
	flag.StringVar(&reuseDatabaseFile, "reuse", "", "reuse database")
	flag.StringVar(&dataDir, "datadir", "", "directory to save the database to")
	flag.StringVar(&format, "format", "txt", "output format: csv (machine readable) or txt (human readable)")
	flag.Parse()

	addressFilePath = flag.Arg(0)
}

func main() {
	// First we set everything up: we connect to the database and detect available providers
	// (based on API keys stored in trueBlocks.toml file)
	comparison := Setup(addressFilePath, dataDir, reuseDatabaseFile)

	// If we aren't using an existing database, we will download data from the providers
	if reuseDatabaseFile == "" {
		if err := comparison.DownloadAppearances(); err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Println("Using existing database, will not download data")
	}

	// Load the results
	results, err := comparison.Results()
	if err != nil {
		log.Fatalln(err)
	}

	// Print the report
	if format == "csv" {
		_ = results.Csv(os.Stdout)
	} else {
		results.Text()
	}
}
