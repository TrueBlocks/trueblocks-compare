package main

import (
	"os"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
)

type Diff struct {
	app   types.SimpleAppearance
	left  bool
	right bool
}

type DiffMap map[types.SimpleAppearance]Diff

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

	contents := file.AsciiFileToLines("addresses.txt")
	for _, line := range contents {
		if downloadOpt {
			download(line)
		}
		compare(line)
	}
}
