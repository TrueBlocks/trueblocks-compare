package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"
)

func compare(line string) {
	if strings.HasPrefix(line, "#") {
		return
	}

	LogIt("Postprocessing etherscan...")
	esSlice := getAppearanceMap("etherscan/" + line + ".csv")
	appsToFile(esSlice, "etherscan/"+line+".csv")

	LogIt("Postprocessing trueblocks...")
	tbSlice := getAppearanceMap("list/" + line + ".csv")
	appsToFile(tbSlice, "list/"+line+".csv")

	LogIt("Building diff maps...")
	diffMap := make(DiffMap)
	for _, app := range esSlice {
		diff := diffMap[app]
		diff.left = true
		diff.app = app
		diffMap[app] = diff
	}
	for _, app := range tbSlice {
		diff := diffMap[app]
		diff.right = true
		diff.app = app
		diffMap[app] = diff
	}

	LogIt("Sorting...")
	diffSlice := make([]Diff, 0, len(diffMap))
	for _, diff := range diffMap {
		diffSlice = append(diffSlice, diff)
	}
	sort.Slice(diffSlice, func(i, j int) bool {
		return diffSlice[i].app.BlockNumber < diffSlice[j].app.BlockNumber || (diffSlice[i].app.BlockNumber == diffSlice[j].app.BlockNumber && diffSlice[i].app.TransactionIndex < diffSlice[j].app.TransactionIndex)
	})

	LogIt("Comparing...")
	for _, diff := range diffSlice {
		app := diff.app
		if diff.left && !diff.right {
			out := fmt.Sprintf("%d,%d\n", app.BlockNumber, app.TransactionIndex)
			file.AppendToAsciiFile("es_only/"+line+".txt", out)
		}
		if !diff.left && diff.right {
			out := fmt.Sprintf("%d,%d\n", app.BlockNumber, app.TransactionIndex)
			file.AppendToAsciiFile("tb_only/"+line+".txt", out)
		}
		if diff.left && diff.right {
			out := fmt.Sprintf("%d,%d\n", app.BlockNumber, app.TransactionIndex)
			file.AppendToAsciiFile("both/"+line+".txt", out)
		}
	}
	clean(line)
	logger.Info("")
}

func getAppearanceMap(filename string) []types.SimpleAppearance {
	contents := file.AsciiFileToLines(filename)
	m := make(map[types.SimpleAppearance]bool)
	for _, line := range contents {
		parts := strings.Split(line, ",")
		app := types.SimpleAppearance{
			BlockNumber:      uint32(utils.MustParseUint(parts[0])),
			TransactionIndex: uint32(utils.MustParseUint(parts[1])),
		}
		m[app] = true
	}

	slice := make([]types.SimpleAppearance, 0, len(contents))
	for app, _ := range m {
		slice = append(slice, app)
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].BlockNumber < slice[j].BlockNumber || (slice[i].BlockNumber == slice[j].BlockNumber && slice[i].TransactionIndex < slice[j].TransactionIndex)
	})

	return slice
}

func appsToFile(slice []types.SimpleAppearance, filename string) {
	lines := make([]string, len(slice))
	for i, app := range slice {
		lines[i] = fmt.Sprintf("%d,%d", app.BlockNumber, app.TransactionIndex)
	}
	file.LinesToAsciiFile(filename, lines)
}

func clean(line string) {
	logger.Info("Cleaning:", line)
	contents := strings.Trim(file.AsciiFileToString("etherscan/"+line+".csv"), "\n")
	if len(contents) > 0 {
		file.StringToAsciiFile("etherscan/"+line+".csv", contents)
	} else {
		os.Remove("etherscan/" + line + ".csv")
	}
	contents = strings.Trim(file.AsciiFileToString("list/"+line+".csv"), "\n")
	if len(contents) > 0 {
		file.StringToAsciiFile("list/"+line+".csv", contents)
	} else {
		os.Remove("list/" + line + ".csv")
	}
}

func LogIt(msg string) {
	logger.Info(colors.Green+msg, colors.Off)
}
