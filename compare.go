package main

// type Diff struct {
// 	app   types.SimpleAppearance
// 	left  bool
// 	right bool
// }

// type DiffMap map[types.SimpleAppearance]Diff

// func compare(line string, min, max int) {
// 	if strings.HasPrefix(line, "#") {
// 		return
// 	}

// 	cnt, _ := file.WordCount("store/list/"+line+".csv", true)
// 	if cnt <= min || cnt > max {
// 		fmt.Println(colors.Red, "Skipping", line, "because it has", cnt, "appearances", colors.Off)
// 		return
// 	}

// 	esDownload := "store/etherscan/" + line + ".csv"
// 	tbDownload := "store/list/" + line + ".csv"

// 	LogIt("Postprocessing etherscan...")
// 	esSlice := getAppearanceMap(esDownload)
// 	appsToFile(esSlice, esDownload)

// 	LogIt("Postprocessing trueblocks...")
// 	tbSlice := getAppearanceMap(tbDownload)
// 	appsToFile(tbSlice, tbDownload)

// 	LogIt("Building diff maps...")
// 	diffMap := make(DiffMap)
// 	for _, app := range esSlice {
// 		diff := diffMap[app]
// 		diff.left = true
// 		diff.app = app
// 		diffMap[app] = diff
// 	}
// 	for _, app := range tbSlice {
// 		diff := diffMap[app]
// 		diff.right = true
// 		diff.app = app
// 		diffMap[app] = diff
// 	}

// 	LogIt("Sorting...")
// 	diffSlice := make([]Diff, 0, len(diffMap))
// 	for _, diff := range diffMap {
// 		diffSlice = append(diffSlice, diff)
// 	}
// 	sort.Slice(diffSlice, func(i, j int) bool {
// 		return diffSlice[i].app.BlockNumber < diffSlice[j].app.BlockNumber || (diffSlice[i].app.BlockNumber == diffSlice[j].app.BlockNumber && diffSlice[i].app.TransactionIndex < diffSlice[j].app.TransactionIndex)
// 	})

// 	LogIt("Comparing...")
// 	es_only := make([]string, 0, len(diffSlice))
// 	tb_only := make([]string, 0, len(diffSlice))
// 	both := make([]string, 0, len(diffSlice))
// 	for _, diff := range diffSlice {
// 		app := diff.app
// 		if diff.left && !diff.right {
// 			es_only = append(es_only, fmt.Sprintf("%d,%d\n", app.BlockNumber, app.TransactionIndex))
// 		}
// 		if !diff.left && diff.right {
// 			tb_only = append(tb_only, fmt.Sprintf("%d,%d\n", app.BlockNumber, app.TransactionIndex))
// 		}
// 		if diff.left && diff.right {
// 			both = append(both, fmt.Sprintf("%d,%d\n", app.BlockNumber, app.TransactionIndex))
// 		}
// 	}

// 	file.LinesToAsciiFile("store/es_only/"+line+".txt", es_only)
// 	file.LinesToAsciiFile("store/tb_only/"+line+".txt", tb_only)
// 	file.LinesToAsciiFile("store/both/"+line+".txt", both)

// 	clean(line)

// 	logger.Info("")
// }

// func getAppearanceMap(filename string) []types.SimpleAppearance {
// 	contents := file.AsciiFileToLines(filename)
// 	m := make(map[types.SimpleAppearance]bool)
// 	for _, line := range contents {
// 		parts := strings.Split(line, ",")
// 		if len(parts) > 1 {
// 			app := types.SimpleAppearance{
// 				BlockNumber:      uint32(utils.MustParseUint(parts[0])),
// 				TransactionIndex: uint32(utils.MustParseUint(parts[1])),
// 			}
// 			m[app] = true
// 		}
// 	}

// 	slice := make([]types.SimpleAppearance, 0, len(contents))
// 	for app := range m {
// 		slice = append(slice, app)
// 	}
// 	sort.Slice(slice, func(i, j int) bool {
// 		return slice[i].BlockNumber < slice[j].BlockNumber || (slice[i].BlockNumber == slice[j].BlockNumber && slice[i].TransactionIndex < slice[j].TransactionIndex)
// 	})

// 	return slice
// }

// func appsToFile(slice []types.SimpleAppearance, filename string) {
// 	lines := make([]string, len(slice))
// 	for i, app := range slice {
// 		lines[i] = fmt.Sprintf("%d,%d", app.BlockNumber, app.TransactionIndex)
// 	}
// 	file.LinesToAsciiFile(filename, lines)
// }

// func cleanOne(fn string) {
// 	apps := getAppearanceMap(fn)
// 	if len(apps) == 0 {
// 		os.Remove(fn)
// 	} else {
// 		contents := make([]string, 0, len(apps))
// 		for _, app := range apps {
// 			contents = append(contents, fmt.Sprintf("%d,%d", app.BlockNumber, app.TransactionIndex))
// 		}
// 		file.LinesToAsciiFile(fn, contents)
// 	}
// }

// func clean(line string) {
// 	logger.Info("Cleaning:", line)
// 	cleanOne(fmt.Sprintf("store/etherscan/%s.csv", line))
// 	cleanOne(fmt.Sprintf("store/list/%s.csv", line))
// 	cleanOne(fmt.Sprintf("store/both/%s.txt", line))
// 	cleanOne(fmt.Sprintf("store/es_only/%s.txt", line))
// 	cleanOne(fmt.Sprintf("store/tb_only/%s.txt", line))
// }

// func LogIt(msg string) {
// 	logger.Info(colors.Green+msg, colors.Off)
// }
