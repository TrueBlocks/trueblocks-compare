# DRAFT: TrueBlocks Comparison with Alchemy, Covalent and Etherscan

- [The Problem](#the-problem)
- [What We Compared](#what-we-compared)
- [The Results](#the-results)
  - [Bug in EtherScan related to Uncles](#bug-in-etherscan-related-to-uncles)
- [What is an Appearance?](#what-is-an-appearance)
- [Why Does This Matter?](#why-does-this-matter)
- [Replicating The Test](#replicating-the-test)

## The Problem

TrueBlocks' command line tool `chifra scrape` produces the [Unchained Index](https://trueblocks.io/papers/2023/specification-for-the-unchained-index-v2.0.0-release.pdf). The best _index of appearances_ that we know of. The Unchained Index includes a record each time an address appears anywhere on the chain. No other indexer, to our knowledge, is as complete.

In this article, we compare the TrueBlocks' indexer with web 2.0 providers: Alchemy, Covalent and Etherscan. Spoiler alert: TrueBlocks wins.

## What We Compared

We queried TrueBlocks, Alchemy, Covalent and Etherscan for "appearances" for about 1,000 randomly-selected addresses. The results are presented below.

There is an important distinction to be made between TrueBlocks and the above mentioned providers. TrueBlocks is local-first indexer running against a local Ethereum archive node. This means TrueBlocks does not rate limit, nor does it cost anything to operate, nor does it paginate. Alchemy, Covalent and Etherscan, on the other hand, is a web 2.0 API--they have no choice but to rate limit, charge for access, and paginate.

These differences, we think, are part of the reason for the surprising results presented below. Given its lightweight nature, TrueBlocks can dig deeper.

## The Results

We checked 1,000 randomly-selected addresses against TrueBlocks and web 2.0 providers mentioned earlier. The list of addresses is available in `addresses.txt` file in this repo. README explains how to rerun the test. The results are saved to SQLite database to allow additional queries.

Of those 1,000 addresses, 102 had no appearances at all and 207 were discarded because they had more than 5,000 appearances. Etherscan's free service limits its return to less than 10,000 records. We wanted to stay as far away from that limit as possible. (Plus, waiting for more than 5,000 records from Etherscan was way too slow. TrueBlocks can easily return 100,000s of records for any address almost instaneously.)

|                             |TrueBlocks|Covalent|Etherscan|Alchemy|
|-----------------------------|----------|--------|---------|-------|
|Addresses Reported           |691       |672     |687      |449    |
|Appearances Reported         |456,269   |311,929 |289,312  |80,851 |
|Unique Addresses Found       |506       |3       |14       |0      |
|Unique Appearances Found     |120,744   |4       |364      |0      |
|Balance Changing Unique Appearances |1,194     |0       |0        |0      |

- *Addresses Reported* is the total number of addresses that a provider returned data for
- *Appearances Reported* is the total number of appearances returned by a provider
- *Unique Addresses Found* is the number of addresses for which ONLY the given provider returned appearances
- *Unique Appearances Found* is the number of appearances reported ONLY by the given provider
- *Balance Changing Appearances* shows how many Unique Appearances involved balance change (ETH)


Of the remaining **691** addresses:

- **506** (**73%**) addresses had appearances found only by TrueBlocks. That's **120,744** more appearances!



> The current draft ends here. It should continue to cover the same stats and sections as the first comparison with Etherscan