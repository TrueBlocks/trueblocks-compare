package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"
)

func main() {
	cleanOpt, combineOpt, removeOpt, downloadOpt := false, false, false, false
	for _, arg := range os.Args[1:] {
		if arg == "--remove" {
			removeOpt = true
		} else if arg == "--download" {
			downloadOpt = true
		} else if arg == "--combine" {
			combineOpt = true
		} else if arg == "--clean" {
			cleanOpt = true
		}
	}

	if cleanOpt {
		cleanAll()
		return
	}
	if combineOpt {
		combine()
		return
	}

	if removeOpt {
		remove()
		return
	}

	min, max := 0, 10000
	contents := file.AsciiFileToLines("addresses.txt")
	for _, line := range contents {
		line = strings.ToLower(line)
		if downloadOpt {
			download(line, min, max)
		}
		compare(line, min, max)
	}
}

func remove() {
	// utils.System("rm -fR tb_only es_only both ; mkdir tb_only es_only both")
	contents := file.AsciiFileToLines("addresses.txt")
	for _, line := range contents {
		line = strings.ToLower(line)
		fn := fmt.Sprintf("store/list/%s.csv", line)
		cnt, _ := file.WordCount(fn, true)
		logger.Info(fn, cnt)
		if cnt > 10000 {
			cmd := fmt.Sprintf("rm -f store/list/%s.csv store/etherscan/%s.csv store/tb_only/%s.csv store/es_only/%s.csv store/both/%s.csv", line, line, line, line, line)
			utils.System(cmd)
			logger.Info("Removed", line)
		}
	}
}

func count(fn string) int {
	contents := file.AsciiFileToLines(fn)
	return len(contents)
}

func combine() {
	fmt.Print("address,list,etherscan,both,es_only,tb_only\n")
	contents := file.AsciiFileToLines("addresses.txt")
	for _, line := range contents {
		line = strings.ToLower(line)
		tb := count(fmt.Sprintf("store/list/%s.csv", line))
		es := count(fmt.Sprintf("store/etherscan/%s.csv", line))
		both := count(fmt.Sprintf("store/both/%s.txt", line))
		es_only := count(fmt.Sprintf("store/es_only/%s.txt", line))
		tb_only := count(fmt.Sprintf("store/tb_only/%s.txt", line))
		if tb > 0 && tb < 10001 {
			out := fmt.Sprintf("%s,%d,%d,%d,%d,%d\n", line, tb, es, both, es_only, tb_only)
			fmt.Print(out)
		}
	}
}

func cleanAll() {
	contents := file.AsciiFileToLines("addresses.txt")
	for _, line := range contents {
		line = strings.ToLower(line)
		clean(line)
	}
}
