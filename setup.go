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

	sdk "github.com/TrueBlocks/trueblocks-core/sdk/v3"
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

// Setup initializes database connection and detects providers
func Setup(addressFile string, dataDir string, databaseFileName string) (c *Comparison) {
	var err error
	c = &Comparison{
		addressFilePath: addressFile,
		minAppearances:  minAppearances,
		maxAppearances:  maxAppearances,
	}

	c.Database, err = NewDatabaseConnection(false, dataDir, databaseFileName)
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

// stringToSlurpSource translates provider name to TrueBlocks SDK provider ID
// (--source value when using chifra slurp on the command line)
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

// DownloadAppearances downloads data from all providers and stores appearances
// in the database
func (c *Comparison) DownloadAppearances() (err error) {
	// First we need to filter addresses with too few and too many appearances out

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

	go func() {
		err = loadAddressesFromFile(c.addressFilePath, addressChan)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	for address := range addressChan {
		log.Println(address)
		appearances, ok, err := c.checkAddress(filterByProvider, address)
		if err != nil {
			return err
		}
		if !ok {
			// Incompatible addresses are saved
			log.Println("Address", address, "is incompatible")
			_ = c.Database.SaveIncompatibleAddress(address, appearances)
			continue
		}

		// Since we had to fetch appearances from the filtering provider
		// (either chifra or Key), we will save them
		for _, appearance := range appearances {
			appData := AppearanceData{}
			appData.Address = appearance.Address
			appData.BlockNumber = appearance.BlockNumber
			appData.TransactionIndex = appearance.TransactionIndex

			// Get reason (where has chifra found the appearance)
			appData.Reason, err = getChifraReason(&appearance)
			if err != nil {
				return err
			}

			// Check if there was a balance change
			appData.BalanceChange, err = getChifraBalanceChange(&appearance)
			if err != nil {
				return err
			}

			if err = c.Database.SaveAppearance(filterByProvider, appData); err != nil {
				return err
			}
		}

		// Now we download the data from the other providers
		for _, provider := range providers {
			log.Println("Downloading from", provider, "address", address)
			providerTypes := typesByProvider(provider)
			for _, providerType := range providerTypes {
				opts := sdk.SlurpOptions{
					Source: stringToSlurpSource(provider),
					Addrs:  []string{address},
					Parts:  providerType,
				}
				appearances, _, err := opts.SlurpAppearances()
				if err != nil {
					log.Fatalln("error downloading data from", provider, err)
				}
				for _, appearance := range appearances {
					appData := AppearanceData{}
					appData.Address = appearance.Address
					appData.BlockNumber = appearance.BlockNumber
					appData.TransactionIndex = appearance.TransactionIndex
					appData.Reason = providerType.String()

					if err = c.Database.SaveAppearance(provider, appData); err != nil {
						return err
					}
				}
			}
		}
	}

	return
}

// typesByProvider returns slice of sdk.SlurpParts that are supported by
// the given provider
func typesByProvider(provider string) (slurpTypes []sdk.SlurpParts) {
	switch provider {
	case "key", "covalent":
		slurpTypes = []sdk.SlurpParts{sdk.SLPAll}
	case "etherscan":
		slurpTypes = []sdk.SlurpParts{
			sdk.SPExt, sdk.SPInt, sdk.SPToken, sdk.SPNfts, sdk.SP1155, sdk.SPMiner, sdk.SPUncles, sdk.SPWithdrawals,
		}
	case "alchemy":
		slurpTypes = []sdk.SlurpParts{
			sdk.SPExt, sdk.SPInt, sdk.SPToken, sdk.SPNfts, sdk.SP1155,
		}
	}
	return
}

// getChifraReason tells us where chifra has found the appearance (e.g. logs, from, etc.)
func getChifraReason(appearance *types.Appearance) (reason string, err error) {
	transactionsOpts := sdk.TransactionsOptions{
		TransactionIds: []string{
			fmt.Sprintf("%d.%d", appearance.BlockNumber, appearance.TransactionIndex),
		},
	}
	// Get reason
	uniqTransactions, _, err := transactionsOpts.TransactionsUniq()
	if err != nil {
		return
	}
	for _, transaction := range uniqTransactions {
		if transaction.Address != appearance.Address {
			continue
		}

		reason = transaction.Reason
	}

	return
}

// getChifraBalanceChange returns true if the balance has changed
func getChifraBalanceChange(appearance *types.Appearance) (balanceChange bool, err error) {
	stateOpts := sdk.StateOptions{
		Addrs:   []string{appearance.Address.String()},
		Parts:   sdk.SPBalance,
		Changes: true,
		BlockIds: []string{
			// Range big enough to capture balance change caused by mining rewards
			strconv.FormatInt(int64(appearance.BlockNumber-2), 10),
			strconv.FormatInt(int64(appearance.BlockNumber+7), 10),
		},
	}
	state, _, err := stateOpts.State()
	if err != nil {
		return
	}
	// there was a balance change
	balanceChange = len(state) > 1
	return
}

// checkAddress checks if the address appearance count doesn't exceed the limits
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
			Parts:  sdk.SLPAll,
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

func (c *Comparison) Results() (r *Result, err error) {
	r = &Result{
		// Number of addresses found by the provider
		AddressesBy: make(map[string]int, len(c.Providers)),
		// Number of appearances found by the provider
		AppearancesBy: make(map[string]int, len(c.Providers)),
		// Number of addresses found ONLY by the provider
		UniqueAddressesBy: make(map[string]int, len(c.Providers)),
		// Number of appearances found ONLY by the provider
		UniqueAppearancesBy: make(map[string]int, len(c.Providers)),
		// Where did it find them?
		GroupedReasonsBy: make(map[string][]GroupedReasons),
		// Number of appearances involving balance change
		BalanceChangesBy: make(map[string]int),

		providers: c.Providers,
	}
	r.AddressCountTotal, err = c.Database.AddressCountTotal()
	if err != nil {
		err = fmt.Errorf("getting total appearance count: %w", err)
		return
	}

	for _, provider := range c.Providers {
		r.AppearancesBy[provider], err = c.Database.AppearanceCount(provider)
		if err != nil {
			err = fmt.Errorf("getting appearance count: %w", err)
			return
		}

		r.AddressesBy[provider], err = c.Database.AddressCount(provider)
		if err != nil {
			err = fmt.Errorf("getting address count: %w", err)
			return
		}

		r.UniqueAppearancesBy[provider], err = c.Database.UniqueAppearanceCount(provider)
		if err != nil {
			err = fmt.Errorf("getting unique appearance count: %w", err)
			return
		}

		r.GroupedReasonsBy[provider], err = c.Database.UniqueAppearancesGroupedReasons(provider)
		if err != nil {
			err = fmt.Errorf("getting grouped reasons: %w", err)
			return
		}

		r.BalanceChangesBy[provider], err = c.Database.BalanceChangeCount(provider)
		if err != nil {
			err = fmt.Errorf("getting balance change: %w", err)
			return
		}

		r.UniqueAddressesBy[provider], err = c.Database.UniqueAddressCount(provider)
		if err != nil {
			err = fmt.Errorf("getting unique address count: %w", err)
			return
		}
	}
	return
}
