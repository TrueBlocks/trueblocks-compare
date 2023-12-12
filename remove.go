package main

import "github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"

func remove() {
	utils.System("rm -fR tb_only es_only both ; mkdir tb_only es_only both")
}
