package main

import (
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/uniq"
)

func tb_only() {
	cnt := 0
	contents := file.AsciiFileToLines("addresses.txt")
	for _, line := range contents {
		if !file.FileExists("store/tb_only/" + line + ".txt") {
			logger.Info("Skipping address", line)
			continue
		}
		logger.Info("Processing address", line)
		apps := getAppearanceMap("store/tb_only/" + line + ".txt")
		for _, app := range apps {
			uniq.GetUniqAddressesInTransaction(app.BlockNumber, app.TransactionIndex)
			// cmd := fmt.Sprintf("chifra transactions --uniq --no_header %d.%d | grep %s", app.BlockNumber, app.TransactionIndex, line)
			// // fmt.Printf("-------- %d.%d ----------------\n", app.BlockNumber, app.TransactionIndex)
			// // fmt.Printf("%s\n", cmd)
			// // fmt.Printf("---------------------------------\n")
			// time.Sleep(3 * time.Second)
			// logger.Info("Sleeping 3	seconds")
			// utils.System(cmd)
			// // fmt.Println(cmd)
			logger.Info(cnt, line, app.BlockNumber, app.TransactionIndex)
			// time.Sleep(2 * time.Second)
			// logger.Info("Sleeping 2	seconds")
			// fmt.Println()
			cnt++
		}
	}
}
