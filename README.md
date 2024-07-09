# TrueBlocks Comparison with other indexers

A repository used to compare other indexers against TrueBlocks.
The methodology and results are described in [Comparison](./results/with-etherscan-2023-12-13.md)

- [Running](#running)
- [Folder Structure](#folder-structure)
- [Code Structure](#code-structure)
- [The Addresses.txt File](#the-addressestxt-file)
- [The Code](#the-code)
  - [Downloading the data](#downloading-the-data)
  - [Comparing the data](#comparing-the-data)
  - [Why does TrueBlocks find more appearances?](#why-does-trueblocks-find-more-appearances)
- [List of Comparisons](#list-of-comparisons)

## Running
Prepare `addresses.txt` file with addresses that should be used for comparison (`data.tar.gz` file has a list of 1,000 addresses).
Then run:

```shell
go run . addresses.txt
```

For 1,000 addresses it takes couple of days to finish. Please see below for the explanation why it takes so much time (for impatient: the cause is other providers' rate limiting, which doesn't happen with local software like TrueBlocks).

After the comparison is done the results are printed to the screen and raw data is preserved in the SQLite database (date and time is used as the file name).
If you'd like to only print the results again, without downloading the data, you can call:
```shell
go run . --reuse path/to/database_file.sqlite
```

By the default the results are present in human-readable, textual form. You can change to CSV (which can be pasted to a spreadsheet app for example) using `format` flag:
```shell
go run . --reuse path/to/database_file.sqlite --format csv
```

## Folder Structure

This is what your directory structure should look like if you wish to reproduce the results:

```[shell]
.                      # The root of the repo. Where the code is stored.
├── bin                # The location of the built file.
└── store              # The location of all the data.
    └── addresses.txt  # The list of addresses used in the comparison
```

## Code Structure

The code is written in GoLang and is located in the root of the repo. It is split into 4 files:

```[shell]
├── main.go        # The main file. It is used to run the code
├── database.go    # The file containing the code to initialize and query SQLite database
├── setup.go       # The file containing preparation code
└── result.go      # The file containing the code to present the results
```

Results are obtained by querying the database. The queries can be found in `database.go` file.

## The Addresses.txt File

Also at the root of the repo is a file called `addresses.txt.`. This is the list of addresses we compared. It is used to download the data from EtherScan and `chifra`. Feel free to replace this file with your own list of addresses.

## The Code

The code to run the comparison is located in `main.go`. Read this very simple file.

### Downloading the data

The `setup.go` file contains the code used to download the data from each source. It reads the `addresses.txt` file and processes each line using TrueBlocks SDK. The code first filters out the addresses that have too few or too many appearances. If `chifra` is installed on the machine, `chifra list` is used for filtering. If TrueBlocks Key Endpoint is configured in `trueBlocks.toml` file (path to the file can be obtained by calling `chifra config --paths`), then TrueBlocks Key is used instead. If none of the mentioned can be used, an error is returned.

If there's not too many appearances (Etherscan doesn't download more than 10,000 records, so we ignore addresses with more than 10,000 records), we procede to download from the provider and store `address, block number, transaction index, provider name` in `appearances` SQL table.

If there are too many or too few appearances, the address is saved in `incompatible_addresses` together with the number of appearances.

Currently supported providers are Alchemy, Covalent Etherscan and TrueBlocks Key. However, the code will only use providers for which API keys (or Endpoint in case of TrueBlocks Key) are defined in `trueBlocks.toml`.

The code used to download the data looks like this:

```go
opts := sdk.SlurpOptions{
  Source: PROVIDER_ID,
  Addrs:  []string{address},
  Parts:  ALL_SUPPORTED,
}
appearances, _, err := opts.SlurpAppearances()
```

Note that the `SlurpOptions` command has a `Parts` field. It is set to every value supported by the given provider. This means it hits all eight of Etherscan's API endpoints: _normal transactions_, _external transactions_, _withdrawals_, etc., all five Alchemy's API and so on. This is the only way to get all the data from most providers. This, when combined with providers rate limiting, means that this process takes a long time to run. `chifra list` is WAY faster.

The download code runs unless you provide `--reuse path/to/existing/database_file.sqlite` flag.

### Comparing the data

To ease comparing the data, a view grouping appearances and providers is present in the database:

```sql
CREATE VIEW IF NOT EXISTS view_appearances_with_providers AS SELECT
  id,
  address,
  block_number,
  transaction_index,
  JSON_GROUP_ARRAY ( provider ) as providers
FROM (SELECT DISTINCT * FROM appearances)
GROUP BY address, block_number, transaction_index;
```

We need to use `SELECT DISTINCT * FROM appearances`, because Etherscan's API endpoints return duplicates.
An example record stored in the view would be:

```
132|0x007b003c4d0145b512286494d5ae123aeef29d9e|4982726|173|["key","etherscan","covalent","alchemy"]
```
Which can be read as: _an appearance with ID 132 of address 0x007b003c4d0145b512286494d5ae123aeef29d9e that has happened in block number 4982726, transaction index 173 was reported by all four providers_

To compare the data, different SQL queries are used. They can be found in`database.go` file.

Comparison has basically two possible outcomes:

1. An appearance is reported by more than 1 provider
2. An appearance is reported only by 1 provider

### Why does TrueBlocks find more appearances?

Hopefully TrueBlocks will find more appearances than other sources. In order to check where these additional appearances come from, for each appearance we call TrueBlocks SDK `TransactionsUniq()` method. It returns `reason` - a string explaining where the appearance has been found.

We store reasons together with provider name and appearance ID in `appearance_reasons` table:
```sql
SELECT * FROM appearance_reasons LIMIT 1;
-- Returns 1|key|log_923_topic_3|
```

For non-TrueBlocks sources the reason is the API endpoint used:
```sql
SELECT * FROM appearance_reasons WHERE provider = 'etherscan' LIMIT 1;
-- Returns 133|etherscan|ext|
```

We also check if the transaction involved a balance change. We detect it by calling TrueBlocks SDK again. Please refer to `getChifraBalanceChange` function for the details.
Information about balance change is stored in `appearance_balance_changes` table defined as follows:

```sql
CREATE TABLE appearance_balance_changes (
	appearance_id INTEGER NOT NULL,
	balance_change BOOLEAN,
	foreign key(appearance_id) references appearances(id)
);
```

## List of Comparisons

We've written a number of comparisons with other data sources. They are listed here:

| Name                                                                                                                                                             | Date       |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| [TrueBlocks / Alchemy, Covalent and Etherscan Comparison](./results/with-3-providers-2024-06-26.md)                                                                                      | 2024-06-26 |
| [TrueBlocks / EtherScan Comparison](./results/with-etherscan-2023-12-13.md)                                                                                      | 2023-12-13 |
| [TrueBlocks / Covalent Comparison](https://medium.com/coinmonks/trueblocks-covalent-comparison-7b42f3d1e6f7)                                                     | 2022-09-20 |
| [The Difference Between TrueBlocks and The Graph](https://trueblocks.io/papers/2021/the-difference-between-trueBlocks-and-rotki-and-trueBlocks-and-thegraph.pdf) | 2021-04-02 |
| [How Accurate is Etherscan](https://tjayrush.medium.com/how-accurate-is-etherscan-83dab12eeedd)                                                                  | 2020-06-11 |
