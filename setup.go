// Setup provides utility functions needed to make the test run

package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	sdk "github.com/TrueBlocks/trueblocks-core/sdk"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/config"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
)

var supportedProviders = []string{
	"key",
	"etherscan",
	"covalent",
	"alchemy",
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

	for _, providerName := range supportedProviders {
		if providerName == "key" {
			if config.GetChain(chain).KeyEndpoint != "" {
				detected = append(detected, "key")
			}
			continue
		}
		if config.GetKey(providerName).ApiKey != "" {
			detected = append(detected, providerName)
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

type AppearancesByProvider struct {
	Provider    string
	Appearances []types.Appearance
}

func (c *Comparison) DownloadAppearances() (err error) {
	var wg sync.WaitGroup
	addressChan := make(chan string, 100)
	appearanceChan := make(chan AppearancesByProvider, 100)
	go func() {
		err = loadAddressesFromFile(c.addressFilePath, addressChan)
		if err != nil {
			// return
			log.Fatalln(err)
		}
	}()
	var providerWg sync.WaitGroup
	channelByProvider := make(map[string]chan string, len(supportedProviders))
	for _, provider := range supportedProviders {
		channelByProvider[provider] = make(chan string, 100)
		// defer close(channelByProvider[provider])
		providerWg.Add(1)
		go func(provider string) {
			for address := range channelByProvider[provider] {
				opts := sdk.SlurpOptions{
					Source: stringToSlurpSource(provider),
					Addrs:  []string{address},
					Types:  sdk.STAll,
				}
				apps, _, err := opts.SlurpAppearances()
				if err != nil {
					log.Fatalln("error downloading data from", provider, err)
				}
				appearanceChan <- AppearancesByProvider{
					Provider:    provider,
					Appearances: apps,
				}
			}
			providerWg.Done()
		}(provider)
	}

	go func() {
		for apps := range appearanceChan {
			log.Println("Saving", len(apps.Appearances), "appearances")
			if err := c.Database.SaveAppearances(apps.Provider, apps.Appearances); err != nil {
				log.Fatalln("error while saving appearances:", err)
			}
			wg.Done()
		}
	}()

	for address := range addressChan {
		downloadedFor, _, err := c.Database.Downloaded(address)
		if err != nil {
			return err
		}

		for _, provider := range supportedProviders {
			if slices.Contains(downloadedFor, provider) {
				continue
			}
			wg.Add(1)
			channelByProvider[provider] <- address
		}
	}
	for _, channel := range channelByProvider {
		close(channel)
	}

	providerWg.Wait()
	close(appearanceChan)

	wg.Wait()

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
