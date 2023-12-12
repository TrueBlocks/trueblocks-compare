# trueblocks-compare

A repository used to compare EtherScan against TrueBlocks.

## Folder Structure

```[shell]
.                  # The root of the repo. Where the code is stored
├── bin            # The location of the built file
└── store          # The location of all the data
    ├── etherscan  # A folder containing all data downloaded from EtherScan
    ├── list       # A folder containing all the data produced by chifra list
    ├── both       # A folder containing appearances found in both data sources
    ├── es_only    # A folder containing appearances found only in EtherScan
    └── tb_only    # A folder containing appearances found only in TrueBlocks
```

## Code Structure

The code is written in GoLang and is located in the root of the repo. It is split into 3 files:

```[shell]
├── main.go        # The main file. It is used to run the code
├── compare.go     # The file containing the code to compare the data
└── download.go    # The file containing the code to download the data
```

## The Addresses.txt File

The `addresses.txt` file is a file containing a list of addresses. It is used to download the data from EtherScan and `chifra`. The file is located in the root of the repo. You may add or remove addresses from the file to change the data that is downloaded.

## The Code

### Downloading the data

The code to download the data is located in the `download.go` file. It reads the `addresses.txt` file and processes each line using `chifra`. The code, which calls into `os.System`, first makes a list of all appearances using `chifra list`. It stores this data into the `store/list` folder. The data has the form:

```[shell]
blockNumber,transactionIndex
```

We then count how many records there are and possibly skip over the file (if it's too big to be useful). The code then downloads the data from Etherscan into the `store/etherscan` folder. This is the command it uses:

```[shell]
chifra list --no_header --last_block 18517000 --fmt csv <address> | cut -d, -f 1,2 >store/list/<address>.csv

# then count and if in range

chifra slurp --types all 0-18517000 --fmt csv <address> | cut -d, -f 1,2 >store/etherscan/<address>.csv
```

At the end of this process, we have one file in each of the two folders (`store/list` and `store/etherscan`) for each address in the `addresses.txt` file with the name `<address>.csv`. This allows us to compare the results easily.

The process will only download if you give it the `--download` flag. Otherwise, it will only compare existing data.

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

