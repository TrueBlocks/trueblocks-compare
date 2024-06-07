// Setup provides utility functions needed to make the test run

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	sdk "github.com/TrueBlocks/trueblocks-core/sdk"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/config"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
)

var supportedProviders = []string{
	// "chifra",
	"key",
	"etherscan",
	"covalent",
	"alchemy",
}

type Comparison struct {
	Database         *Database
	DatabaseFileName string
	Providers        []string

	addressFilePath string
	minAppearances  int
	maxAppearances  int
}

func Setup(databaseFileName string) (c *Comparison) {
	var err error
	c = &Comparison{
		addressFilePath: os.Args[1],
		maxAppearances:  5000,
	}

	c.Database, err = NewDatabaseConnection(false, databaseFileName)
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

func (c *Comparison) DownloadAppearances() (err error) {
	addressChan := make(chan string, 100)
	filterByProvider := "key"
	if slices.Contains(c.Providers, "chifra") {
		filterByProvider = "chifra"
	}
	log.Println("Filtering by", filterByProvider)
	providers := make([]string, 0, len(c.Providers))
	for _, provider := range c.Providers {
		if provider != filterByProvider {
			providers = append(providers, provider)
		}
	}

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
		// toSave := make([]AppearanceData, 0, len(appearances))

		for i := 0; i < len(appearances); i++ {
			appData := AppearanceData{
				// Appearance: types.Appearance{
				// 	BlockNumber: appearances[i].BlockNumber,
				// 	TransactionIndex: appearances[i].TransactionIndex,
				// },
			}
			appData.Address = appearances[i].Address
			appData.BlockNumber = appearances[i].BlockNumber
			appData.TransactionIndex = appearances[i].TransactionIndex

			transactionsOpts := sdk.TransactionsOptions{
				TransactionIds: []string{
					fmt.Sprintf("%d.%d", appearances[i].BlockNumber, appearances[i].TransactionIndex),
				},
			}
			transactions, _, err := transactionsOpts.TransactionsUniq()
			if err != nil {
				return err
			}
			for _, uniqAppearance := range transactions {
				if uniqAppearance.Address != appearances[i].Address {
					continue
				}

				appData.Reason = uniqAppearance.Reason
			}

			// ("chifra state --no_header --parts balance %d-%d %s --changes", app.BlockNumber-2, app.BlockNumber+7, line)
			stateOpts := sdk.StateOptions{
				Addrs:   []string{appearances[i].Address.String()},
				Parts:   sdk.SPBalance,
				Changes: true,
				BlockIds: []string{
					strconv.FormatInt(int64(appearances[i].BlockNumber-2), 10),
					strconv.FormatInt(int64(appearances[i].BlockNumber+7), 10),
				},
			}
			state, _, err := stateOpts.State()
			if err != nil {
				return err
			}
			if len(state) > 1 {
				// there was a balance change
				appData.BalanceChange = true
			}
			// toSave = append(toSave, appData)
			if err = c.Database.SaveAppearance(filterByProvider, appData); err != nil {
				return err
			}
		}
		// if err = c.Database.SaveAppearances(filterByProvider, toSave); err != nil {
		// 	return err
		// }

		for _, provider := range providers {
			log.Println("Downloading from", provider, "address", address)
			providerTypes := typesByProvider(provider)
			for _, providerType := range providerTypes {
				opts := sdk.SlurpOptions{
					Source: stringToSlurpSource(provider),
					Addrs:  []string{address},
					Types:  providerType,
				}
				appearances, _, err := opts.SlurpAppearances()
				if err != nil {
					log.Fatalln("error downloading data from", provider, err)
				}
				// toSave := make([]AppearanceData, 0, len(appearances))

				// for i := 0; i < len(appearances); i++ {
				// 	appearances[i].Reason = providerType.String()
				// }
				for _, appearance := range appearances {
					appData := AppearanceData{}
					appData.Address = appearance.Address
					appData.BlockNumber = appearance.BlockNumber
					appData.TransactionIndex = appearance.TransactionIndex
					appData.Reason = providerType.String()
					// toSave = append(toSave, appData)
					if err = c.Database.SaveAppearance(provider, appData); err != nil {
						return err
					}
				}

				// if err = c.Database.SaveAppearances(provider, toSave); err != nil {
				// 	return err
				// }
				// if err = c.Database.MarkAsDownloaded(address, provider); err != nil {
				// 	return err
				// }
			}
		}
	}

	return
}

func typesByProvider(provider string) (slurpTypes []sdk.SlurpTypes) {
	switch provider {
	case "key", "covalent":
		slurpTypes = []sdk.SlurpTypes{sdk.STAll}
	case "etherscan":
		slurpTypes = []sdk.SlurpTypes{
			sdk.STExt, sdk.STInt, sdk.STToken, sdk.STNfts, sdk.ST1155, sdk.STMiner, sdk.STUncles, sdk.STWithdrawals,
		}
	case "alchemy":
		slurpTypes = []sdk.SlurpTypes{
			sdk.STExt, sdk.STInt, sdk.STToken, sdk.STNfts, sdk.ST1155,
		}
	}
	return
}

type Result struct {
	AddressCount      int
	AppearancesBy     map[string]int
	AppearancesOnlyBy map[string]int
	GroupedReasons    map[string][]GroupedReasons
	BalanceChanges    map[string]int
	AddressesBy       map[string]int
	AddressesOnlyBy   map[string]int
}

func (c *Comparison) Results() (r *Result, err error) {
	r = &Result{
		AppearancesBy:     make(map[string]int, len(c.Providers)),
		AddressesBy:       make(map[string]int, len(c.Providers)),
		AppearancesOnlyBy: make(map[string]int, len(c.Providers)),
		GroupedReasons:    make(map[string][]GroupedReasons),
		BalanceChanges:    make(map[string]int),
		AddressesOnlyBy:   make(map[string]int, len(c.Providers)),
	}
	r.AddressCount, err = c.Database.AddressCount()
	if err != nil {
		return
	}

	for _, provider := range c.Providers {
		var appearances []types.Appearance
		// var addressCount int
		appearances, err = c.Database.AppearancesHavingProvider(provider)
		if err != nil {
			return
		}
		r.AppearancesBy[provider] = len(appearances)

		r.AddressesBy[provider], err = c.Database.AddressCountHavingProvider(provider)
		if err != nil {
			return
		}
		// r.AddressesBy[provider] = addressCount

		appearances, err = c.Database.AppearancesByProviders([]string{provider})
		if err != nil {
			return
		}
		r.AppearancesOnlyBy[provider] = len(appearances)

		var groupedReasons []GroupedReasons
		groupedReasons, err = c.Database.UniqueAppearancesGroupedReasons(provider)
		if err != nil {
			return
		}
		r.GroupedReasons[provider] = groupedReasons

		r.BalanceChanges[provider], err = c.Database.AppearanceBalanceChangeCountOnlyByProvider(provider)
		if err != nil {
			return
		}

		r.AddressesOnlyBy[provider], err = c.Database.AddressCountByProviders([]string{provider})
		if err != nil {
			return
		}
		// r.AddressesOnlyBy[provider] = len(addressCount)
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
