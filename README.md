# trueblocks-compare

A repository used to compare EtherScan against TrueBlocks.

## Folder Structure

```[shell]
.                  # The root of the repo. Where the code is stored.
├── bin            # The location of the built file.
└── store          # The location of all the data.
    ├── etherscan  # A folder containing all data downloaded from EtherScan.
    ├── list       # A folder containing all the data produced by chifra list.
    ├── both       # A folder containing appearances found in both data sources.
    ├── es_only    # A folder containing appearances found only in EtherScan.
    └── tb_only    # A folder containing appearances found only in TrueBlocks.
```

## Code Structure

The code is written in GoLang and is located in the root of the repo. It is split into 3 files:

```[shell]
├── main.go        # The main file. It is used to run the code
├── compare.go     # The file containing the code to compare the data
└── download.go    # The file containing the code to download the data
```

## The Addresses.txt File

Also at the root of the repo is a file called `addresses.txt.`. This is the list of addresses we compared. It is used to download the data from EtherScan and `chifra`. Feel free to replace this file with your own list of addresses.

## The Code

### Downloading the data

The code to run the comparison is located in `main.go`. Read this very simple file. It's clear how it works.

The `download.go` file contains the code used to download the data from each source. It reads the `addresses.txt` file and processes each line using `chifra` and `os.System`. The code first creates a list of all appearances using `chifra list`. It stores this list into the `store/list` folder. The data has the form:

```[shell]
blockNumber,transactionIndex
```

Next, we count how many records are found by `chifra list.` If there's not too many (EtherScan doesn't download more than 10,000 records, so we ignore addresses with more than 10,000 records), we procede to download from EtherScan. It stores the EtherScan data into the `store/etherscan` folder.

The command it uses for `chifra list` is:

```[shell]
chifra list --no_header --last_block 18517000 --fmt csv <address> | cut -d, -f 1,2 >store/list/<address>.csv
```

If there's less than 10,000 records, it downloads from EtherScan using the command:

```[shell]
chifra slurp --types all 0-18517000 --fmt csv <address> | cut -d, -f 1,2 >store/etherscan/<address>.csv
```

Note that the `chifra slurp` command has a `--types` option which takes a value of `all`. This means it hits all eight of EtherScan's API's data types: `ext | int | token | nfts | 1155 | miner | uncles | withdrawals`. This is the only way to get all the data from EtherScan. This, when combined with EtherScan's rate limiting, means that this process takes a long time to run. `chifra list` is WAY faster.

At the end of this process, we have one file in each of the two folders (`store/list` and `store/etherscan`) for each address in the `addresses.txt` file with less than 10,000 appearances. The with the name of the file is `<address>.csv`. This allows us to compare the results easily.

The process only download runs the `download` process if you provide the `--download` flag. Otherwise, it only compares existing data.

### Comparing the data

To compare the files, we read in both files for each address. As we read the file, we enter each appearance into a map that maps the appearance into a `Diff` structure which is a pair of `booleans`. Like this:

```[go]
type Diff struct {
    app   types.SimpleAppearance
    left  bool
    right bool
}

type DiffMap map[types.SimpleAppearance]Diff
```

If the appearance is found in the `etherscan` file, we light up the left boolean. If it's found in the `list` file, we light up the right boolean. At the end of the process, we have a map that contains all the appearances found in both files (both booleans are lit), all the appearances found only in the `etherscan` file (only the left (i.e. etherscan) boolean is lit), and all the appearances found only in the `list` file (only the right (i.e. trueblocks) boolean is lit).

We write the appearances into the appropriate folders (`store/both`, `store/es_only`, and `store/tb_only`) and we're done.

## The Results

