package main

import (
	"fmt"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"
)

func download(line string, min, max int) {
	if strings.HasPrefix(line, "#") {
		return
	}

	cmd := "chifra list --no_header --last_block 18517000 --fmt csv " + line + " | cut -d, -f 2-3 >list/" + line + ".csv"
	LogIt("Running: " + cmd)
	utils.System(cmd)
	LogIt("Finished...")
	cnt, _ := file.WordCount("list/"+line+".csv", true)
	if cnt <= min || cnt > max {
		fmt.Println(colors.Red, "Skipping", line, "because it has", cnt, "appearances", colors.Off)
		return
	}
	fmt.Println(colors.BrightYellow, "Processing", line, "because it has", cnt, "appearances", colors.Off)
	cmd = "chifra slurp --sleep 1 --no_header --types all --appearances 0-18517000  --fmt csv " + line + " | cut -d, -f 2-3 >etherscan/" + line + ".csv"
	LogIt("Running: " + cmd)
	utils.System(cmd)
	LogIt("Finished...")
}
