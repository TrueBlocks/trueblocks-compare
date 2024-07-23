package main

import (
	"encoding/csv"
	"fmt"
	"io"
)

type Result struct {
	AddressCountTotal   int
	AppearancesBy       map[string]int
	UniqueAppearancesBy map[string]int
	GroupedReasonsBy    map[string][]GroupedReasons
	BalanceChangesBy    map[string]int
	AddressesBy         map[string]int
	UniqueAddressesBy   map[string]int
	providers           []string
}

func (r *Result) Csv(outputWriter io.Writer) (err error) {
	w := csv.NewWriter(outputWriter)
	err = w.Write([]string{
		"label", "addresses", "appearances", "unique addresses", "unique appearances", "unique appearance source", "balance changes",
	})
	if err != nil {
		return
	}
	if err = w.Write([]string{"addresses compared", fmt.Sprint(r.AddressCountTotal)}); err != nil {
		return
	}

	for _, provider := range r.providers {
		var groupedReasonsString string
		for _, gr := range r.GroupedReasonsBy[provider] {
			groupedReasonsString += fmt.Sprintf("%s: %d;", gr.Reason, gr.Count)
		}

		err = w.Write([]string{
			provider,
			fmt.Sprint(r.AddressesBy[provider]),
			fmt.Sprint(r.AppearancesBy[provider]),
			fmt.Sprint(r.UniqueAddressesBy[provider]),
			fmt.Sprint(r.UniqueAppearancesBy[provider]),
			groupedReasonsString,
			fmt.Sprint(r.BalanceChangesBy[provider]),
		})
		if err != nil {
			return
		}
	}

	w.Flush()
	return
}

func (r *Result) Text() {
	fmt.Printf("Addresses compared: %d \n", r.AddressCountTotal)
	for _, provider := range r.providers {
		fmt.Printf("Provider: %s\n", provider)
		fmt.Printf("\t%d addresses and %d appearances in total\n", r.AddressesBy[provider], r.AppearancesBy[provider])
		fmt.Printf("\t%d unique addresses\n", r.UniqueAddressesBy[provider])
		fmt.Printf("\t%d unique appearances\n", r.UniqueAppearancesBy[provider])

		if r.UniqueAppearancesBy[provider] > 0 {
			fmt.Printf("\n\tSources of unique appearances:\n")
			for _, groupedReasons := range r.GroupedReasonsBy[provider] {
				fmt.Printf("\t\t%s: %d\n", groupedReasons.Reason, groupedReasons.Count)
			}
		}
		if provider == "chifra" || provider == "key" {
			fmt.Println("\tBalance changes:", r.BalanceChangesBy[provider])
		}
	}
}
