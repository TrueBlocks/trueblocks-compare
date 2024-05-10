package main

func Download(db *Database, provider string, address string) {

}

// func download(line string, min, max int) {
// 	if strings.HasPrefix(line, "#") {
// 		return
// 	}

// 	cnt, _ := file.WordCount("store/list/"+line+".csv", true)
// 	if file.FileExists("store/list/"+line+".csv") && cnt > 0 {
// 		fmt.Println(colors.BrightYellow, "Skipping", line, "because it already exists", colors.Off)
// 		return
// 	}

// 	cmd := "chifra list --no_header --last_block 18517000 --fmt csv " + line + " | cut -d, -f 2-3 >store/list/" + line + ".csv"
// 	LogIt("Running: " + cmd)
// 	utils.System(cmd)
// 	LogIt("Finished...")

// 	cnt, _ = file.WordCount("store/list/"+line+".csv", true)
// 	if cnt <= min || cnt > max {
// 		fmt.Println(colors.Red, "Skipping", line, "because it has", cnt, "appearances", colors.Off)
// 		return
// 	}

// 	fmt.Println(colors.BrightYellow, "Processing", line, "because it has", cnt, "appearances", colors.Off)
// 	cmd = "chifra slurp --sleep 1 --no_header --types all --appearances 0-18517000  --fmt csv " + line + " | cut -d, -f 2-3 >store/etherscan/" + line + ".csv"
// 	LogIt("Running: " + cmd)
// 	utils.System(cmd)
// 	LogIt("Finished...")
// }
