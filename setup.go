// Setup provides utility functions needed to make the test run

package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"

	sdk "github.com/TrueBlocks/trueblocks-core/sdk"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/config"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
)

var supportedProviders = []string{
	"chifra",
	"key",
	"etherscan",
	// "covalent",
	// "alchemy",
}

type Comparison struct {
	Database  *Database
	Providers []string

	addressFilePath string
	minAppearances  int
	maxAppearances  int
}

func Setup() (c *Comparison) {
	var err error
	c = &Comparison{
		addressFilePath: os.Args[1],
		maxAppearances:  5000,
	}

	c.Database, err = NewDatabaseConnection(false)
	if err != nil {
		log.Fatalln("opening sqlite database:", err)
	}

	c.detectProviders("mainnet")
	if len(c.Providers) == 0 {
		log.Fatalln("no providers configured. Please see README")
	}

	return
}

// detectEnabledProviders goes over the list of providers and checks
// if we have API key for each one. If we do, then the provider is enabled.
func (c *Comparison) detectProviders(chain string) {
	detected := make([]string, 0, len(supportedProviders))

	var hasChifra bool
	var hasKey bool

	for _, providerName := range supportedProviders {
		switch providerName {
		case "chifra":
			// statusOpts := &sdk.StatusOptions{}
			// _, _, err := statusOpts.StatusDiagnose()
			_, err := exec.LookPath("chifra")
			if err == nil {
				detected = append(detected, "chifra")
				hasChifra = true
			} else {
				log.Println("you don't have chifra:", err)
			}
			continue
		case "key":
			if config.GetChain(chain).KeyEndpoint != "" {
				detected = append(detected, "key")
				hasKey = true
			}
			continue
		default:
			if config.GetKey(providerName).ApiKey != "" {
				detected = append(detected, providerName)
			}
		}
		if !hasChifra && !hasKey {
			log.Fatalln("either chifra or Key is required")
		}
	}

	c.Providers = detected
}

func stringToSlurpSource(name string) (source sdk.SlurpSource) {
	switch name {
	case "etherscan":
		source = sdk.SSEtherscan
	case "alchemy":
		source = sdk.SSAlchemy
	case "covalent":
		source = sdk.SSCovalent
	case "key":
		source = sdk.SSKey
	default:
		log.Fatalln("unsupported provider", name)
	}

	return source
}

// type AppearancesByProvider struct {
// 	Provider    string
// 	Appearances []types.Appearance
// }

func (c *Comparison) DownloadAppearances() (err error) {
	addressChan := make(chan string, 100)
	filterByProvider := "key"
	if slices.Contains(c.Providers, "chifra") {
		filterByProvider = "chifra"
	}
	providers := slices.DeleteFunc(c.Providers, func(element string) bool {
		return element == filterByProvider
	})

	err = loadAddressesFromFile(c.addressFilePath, addressChan)
	if err != nil {
		log.Fatalln(err)
	}

	for address := range addressChan {
		appearances, ok, err := c.checkAddress(filterByProvider, address)
		if err != nil {
			return err
		}
		if !ok {
			log.Println("Address", address, "is incompatible")
			c.Database.SaveIncompatibleAddress(address, appearances)
			continue
		}
		if err = c.Database.SaveAppearances(filterByProvider, appearances); err != nil {
			return err
		}

		for _, provider := range providers {
			log.Println("Downloading from", provider, "address", address)
			opts := sdk.SlurpOptions{
				Source: stringToSlurpSource(provider),
				Addrs:  []string{address},
				Types:  sdk.STAll,
			}
			appearances, _, err := opts.SlurpAppearances()
			if err != nil {
				log.Fatalln("error downloading data from", provider, err)
			}
			if err = c.Database.SaveAppearances(provider, appearances); err != nil {
				return err
			}
		}
	}

	return
}

func (c *Comparison) checkAddress(provider string, address string) (appearances []types.Appearance, ok bool, err error) {
	if provider == "chifra" {
		listOpts := &sdk.ListOptions{
			Addrs: []string{address},
		}
		appearances, _, err = listOpts.List()

	} else {
		opts := sdk.SlurpOptions{
			Source: stringToSlurpSource(provider),
			Addrs:  []string{address},
			Types:  sdk.STAll,
		}
		appearances, _, err = opts.SlurpAppearances()
	}

	ok = len(appearances) >= c.minAppearances && len(appearances) <= c.maxAppearances
	return
}

func loadAddressesFromFile(filePath string, addressChan chan string) (err error) {
	var addressFile *os.File
	if filePath == "" {
		addressFile = os.Stdin
	} else {
		addressFile, err = os.Open(filePath)
		if err != nil {
			log.Fatalln("opening addresses file:", err)
		}
	}
	defer close(addressChan)

	reader := bufio.NewReader(addressFile)
	for {
		address, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// We are done
				break
			}
			log.Fatalln("while processing addresses file:", err)
		}
		// Drop \n
		sanitized := strings.TrimSuffix(address, "\n")
		if len(sanitized) == 0 {
			continue
		}
		addressChan <- sanitized
	}

	return
}
