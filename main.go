package main

import (
	"os"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"
)

func main() {
	removeOpt, downloadOpt := false, false
	for _, arg := range os.Args[1:] {
		if arg == "--remove" {
			removeOpt = true
		} else if arg == "--download" {
			downloadOpt = true
		}
	}

	if removeOpt {
		remove()
	}

	min, max := 35000, 50000
	contents := file.AsciiFileToLines("addresses.txt")
	for _, line := range contents {
		if downloadOpt {
			download(line, min, max)
		}
		compare(line, min, max)
	}
}

func remove() {
	utils.System("rm -fR tb_only es_only both ; mkdir tb_only es_only both")
}
