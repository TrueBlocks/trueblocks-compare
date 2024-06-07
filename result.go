package main

import (
	"encoding/csv"
	"fmt"
	"io"
)

type Result struct {
	AddressCount      int
	AppearancesBy     map[string]int
	AppearancesOnlyBy map[string]int
	GroupedReasons    map[string][]GroupedReasons
	BalanceChanges    map[string]int
	AddressesBy       map[string]int
	AddressesOnlyBy   map[string]int
	providers         []string
}

func (r *Result) Csv(outputWriter io.Writer) (err error) {
	w := csv.NewWriter(outputWriter)
	err = w.Write([]string{
		"label", "addresses", "appearances", "unique addresses", "unique appearances", "unique appearance source", "balance changes",
	})
	if err != nil {
		return
	}
	if err = w.Write([]string{"addresses compared", fmt.Sprint(r.AddressCount)}); err != nil {
		return
	}

	for _, provider := range r.providers {
		var groupedReasonsString string
		for _, gr := range r.GroupedReasons[provider] {
			groupedReasonsString += fmt.Sprintf("%s: %d;", gr.Reason, gr.Count)
		}

		err = w.Write([]string{
			provider,
			fmt.Sprint(r.AddressesBy[provider]),
			fmt.Sprint(r.AppearancesBy[provider]),
			fmt.Sprint(r.AddressesOnlyBy[provider]),
			fmt.Sprint(r.AppearancesOnlyBy[provider]),
			groupedReasonsString,
			fmt.Sprint(r.BalanceChanges[provider]),
		})
		if err != nil {
			return
		}
	}

	w.Flush()
	return
}

func (r *Result) Text() {
	fmt.Printf("Addresses compared: %d \n", r.AddressCount)
	for _, provider := range r.providers {
		fmt.Printf("Provider: %s\n", provider)
		fmt.Printf("\t%d addresses and %d appearances in total\n", r.AddressesBy[provider], r.AppearancesBy[provider])
		fmt.Printf("\t%d unique addresses\n", r.AddressesOnlyBy[provider])
		fmt.Printf("\t%d unique appearances\n", r.AppearancesOnlyBy[provider])

		if r.AppearancesOnlyBy[provider] > 0 {
			fmt.Printf("\n\tSources of unique appearances:\n")
			for _, groupedReasons := range r.GroupedReasons[provider] {
				fmt.Printf("\t\t%s: %d\n", groupedReasons.Reason, groupedReasons.Count)
			}
		}
		if provider == "chifra" || provider == "key" {
			fmt.Println("\tBalance changes:", r.BalanceChanges[provider])
		}
	}
}
