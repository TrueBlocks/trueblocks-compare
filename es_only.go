package main

// func es_only() {
// 	cnt := 0
// 	contents := file.AsciiFileToLines("store/addresses.txt")
// 	for _, line := range contents {
// 		if !file.FileExists("store/es_only/" + line + ".txt") {
// 			logger.Info("Skipping address", line)
// 			continue
// 		}
// 		logger.Info("Processing address", line)
// 		apps := getAppearanceMap("store/es_only/" + line + ".txt")
// 		for _, app := range apps {
// 			cmd := fmt.Sprintf("chifra state --no_header --parts balance %d-%d %s --changes", app.BlockNumber-2, app.BlockNumber+7, line)
// 			fmt.Printf("-------- %d.%d ----------------\n", app.BlockNumber, app.TransactionIndex)
// 			// fmt.Printf("%s\n", cmd)
// 			// fmt.Printf("---------------------------------\n")
// 			utils.System(cmd)
// 			logger.Info(cnt, line, app.BlockNumber, app.TransactionIndex)
// 			fmt.Println()
// 			cnt++
// 		}
// 	}
// }
