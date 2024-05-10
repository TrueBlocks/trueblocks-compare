package main

// func tb_only() {
// 	contents := file.AsciiFileToLines("store/addresses.txt")
// 	for _, line := range contents {
// 		if !file.FileExists("store/tb_only/" + line + ".txt") {
// 			logger.Info("Skipping address", line)
// 			continue
// 		}
// 		logger.Info("Processing address", line)
// 		apps := getAppearanceMap("store/tb_only/" + line + ".txt")
// 		for i, app := range apps {
// 			cmd := fmt.Sprintf("chifra transactions --uniq --no_header %d.%d | grep %s >>store/found/%s.txt", app.BlockNumber, app.TransactionIndex, line, line)
// 			utils.System(cmd)
// 			logger.Info(i, len(apps), line, app.BlockNumber, app.TransactionIndex)
// 		}
// 	}
// }
