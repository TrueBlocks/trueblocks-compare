# TrueBlocks Comparison with Alchemy, Covalent and Etherscan

- [The Problem](#the-problem)
- [What We Compared](#what-we-compared)
- [The Results](#the-results)
  - [Bug in Etherscan related to Uncles](#bug-in-etherscan-related-to-uncles)
- [What is an Appearance?](#what-is-an-appearance)
- [Why Does This Matter?](#why-does-this-matter)
- [Replicating The Test](#replicating-the-test)

## The Problem

TrueBlocks' command line tool `chifra scrape` produces the [Unchained Index](https://trueblocks.io/papers/2023/specification-for-the-unchained-index-v2.0.0-release.pdf). The best _index of appearances_ that we know of. ([See below if you don't know what an appearance is](#what-is-an-appearance).) The Unchained Index includes a record each time an address appears anywhere on the chain. No other indexer, to our knowledge, is as complete.

In this article, we compare the TrueBlocks' indexer with web 2.0 providers: Alchemy, Covalent and Etherscan. Spoiler alert: TrueBlocks wins.

## What We Compared

We queried TrueBlocks, Alchemy, Covalent and Etherscan for "appearances" for about 1,000 randomly-selected addresses. The results are presented below.

There is an important distinction to be made between TrueBlocks and the above mentioned providers. TrueBlocks is local-first indexer running against a local Ethereum archive node. This means TrueBlocks does not rate limit, nor does it cost anything to operate, nor does it paginate. Alchemy, Covalent and Etherscan, on the other hand, are web 2.0 APIs &mdash; they have no choice but to rate limit, charge for access, and paginate.

These differences, we think, are part of the reason for the surprising results presented below. Given its lightweight nature, TrueBlocks can dig deeper.

We made sure that every provider had been queried for all addresses. Had there been a single error response, the testing tool would exit without finishing the test.

## The Results

We checked 1,000 randomly-selected addresses against TrueBlocks and web 2.0 providers mentioned earlier. The list of addresses is available in `addresses.txt` file in this repo. README explains how to rerun the test. The results are saved to SQLite database to allow additional queries.

Of those 1,000 addresses, 102 had no appearances at all and 207 were discarded because they had more than 5,000 appearances. Etherscan's free service limits its return to less than 10,000 records. We wanted to stay as far away from that limit as possible. (Plus, waiting for more than 5,000 records from Etherscan was way too slow. TrueBlocks can easily return 100,000s of records for any address almost instaneously.)

|                                                         |TrueBlocks|Covalent|Etherscan|Alchemy|
|-------------------------------------------------------- |----------|--------|---------|-------|
|Addresses Queried                                        |1,000     |1,000   |1,000    |1,000  |
|Addresses with Too Many Appearances 	                    |207       |207     |207      |207    |
|Addresses with No Appearances                            |102       |121     |106      |344    |
|Addresses with Appearances                               |691       |672     |687      |449    |
|Appearances Reported                                     |456,269   |311,929 |289,312  |80,851 |
|Addresses for which only the provider found a transaction|506       |0       |14       |0      |
|Unique Appearances Found                                 |120,744   |0       |364*     |0      |
|Balance Changing Unique Appearances                      |1,194     |0       |0        |0      |
\* &mdash; Please see [Bug in Etherscan related to Uncles](#bug-in-etherscan-related-to-uncles) for the explanaition why Etherscan has found 364 unique appearances.

### Row description
- *Addresses Queried* is the initial number of addresses before filtering any addresses out
- *Addresses with Too Many Appearances* is the number of addresses exceeding allowed appearance total (5,000 in this case, see the paragraph above)
- *Addresses with No Appearances* is the number of addresses for which the given provider returned no appearances
- *Addresses with Appearances* is the total number of addresses not exceeding allowed appearance total that a provider returned data for
- *Appearances Reported* is the total number of appearances returned by a provider
- *Addresses for which only the provider found a transaction* is the number of addresses for which _only_ the given provider returned appearances
- *Unique Appearances Found* is the number of appearances reported _only_ by the given provider
- *Balance Changing Appearances* shows how many Unique Appearances involved balance change (ETH only)

### Summary

Of the remaining **691** addresses:

- **506** (**73%**) addresses had appearances found only by TrueBlocks. That's **120,744** more appearances!
- **1,194** appearances found by TrueBlocks changed address' ETH balance
- **NO** appearances were found by other providers that were not also found by TrueBlocks
- **Only** TrueBlocks found appearances for all **691** addresses
- for **15** addresses, Etherscan found 364 different appearances than TrueBlocks, but in all 15 cases, the difference was due to a bug in Etherscan. ([See below](#bug-in-etherscan-related-to-uncles).)

We recognize that the huge number of additional appearances found by TrueBlocks seems like a mistake. But one needs to realize that TrueBlocks looks for more than just a small set of known behaviours (such as `Transfers`). TrueBlocks looks everywhere. In particulate, TrueBlocks looks in:

- the transaction's `input` data
- the `topics` of the transaction's logs
- the `data` field of the transaction's logs
- the `data` and `output` field of the transaction's traces

### Bug in Etherscan related to Uncles

In the **687** addresses searched by Etherscan, for **14** addresses it found appearances that TrueBlocks did not. In all cases, however, the difference was due to a bug in Etherscan related to uncles. The bug is that Etherscan returns the block number when the uncle was "located". TrueBlocks returns the block number in which the uncle reward was credited to the address's account. We know this because we ran the following analysis on all **364** appearances of this issue.

First, we extracted just the block number from the appearances found by Etherscan. We then calculated 1 block prior to that block number (P) and seven blocks after that block number (A). We then ran:

```shell
chifra state --parts balance P-A <address> --changes
```

which uses another one of the `chifra` tools to extract the balances for the given address at the given blocks. For example, for address `0x3f98e477a361f777da14611a7e419a75fd238b6b`, Etherscan reports the following appearances:

```shell
485,uncle
940,uncle
1114,uncle
...
```

This command

```shell
chifra state --parts balance 484-492 0x48040276e9c17ddbe5c8d2976245dcd0235efa43
```

returns

|blockNumber|address|balance|
|-----------|-------|-------|
|484|0x48040276e9c17ddbe5c8d2976245dcd0235efa43|0|
|485|0x48040276e9c17ddbe5c8d2976245dcd0235efa43|0|
|486|0x48040276e9c17ddbe5c8d2976245dcd0235efa43|0|
|487|0x48040276e9c17ddbe5c8d2976245dcd0235efa43|3750000000000000000|
|488|0x48040276e9c17ddbe5c8d2976245dcd0235efa43|3750000000000000000|
|489|0x48040276e9c17ddbe5c8d2976245dcd0235efa43|3750000000000000000|
|490|0x48040276e9c17ddbe5c8d2976245dcd0235efa43|3750000000000000000|
|491|0x48040276e9c17ddbe5c8d2976245dcd0235efa43|3750000000000000000|

As you can see, Etherscan reports the uncle block at block 485. However, the uncle reward was not credited to the miner's account until block 487. TrueBlocks reports the uncle at block 487.

In all 364 cases, the block Etherscan reports as the uncle block is technically correct. However, the uncle reward was not credited to the miner's account until a few blocks later. In each case, that block was the block that TrueBlocks reported. Etherscan got it wrong. Unless you want to lean on a technicality. I would argue that a change in balance of an account is the correct place to note in an address's history. I'll leave it up to Etherscan to decide if they want to fix this bug.

Total number of place where Etherscan legitmately found more appearances than TrueBlocks: **ZERO**!

## What is an Appearance?

"Appearances" are seemingly simple. For any address, the address's list of appearances is a list of blocknumber.transactionId pairs noting whereever the address appears on the chain.

For example, the first three appearances for `trueblocks.eth` are:

|blockNumber|transactionIndex|
|-----------|----------------|
|8854723|61|
|8856290|62|
|8856316|91|

Easy enough. Just look at `from`, `to`, `log topic 0` and a few other places. That's what most indexers do. But as we've demonstrated above, there's way more to the story. Please see a very detailed discussion in the [Specification of the Unchained Index](https://trueblocks.io/papers/2023/specification-for-the-unchained-index-v2.0.0-release.pdf).

## Why Does This Matter?

Blockchains are perfect, 18-decimal place accurate accounting systems. Every 12 seconds, they come to balance on many hundreds of millions of accounts. That's true on-chain.

The fact that even Etherscan, our industries leading data provider can't get it right, is imporant because blockchains should balance perfectly off-chain as well as on-chain. What the hell are we even building otherwise if we can't account for every single wei for every single account. That's what TrueBlocks does. That's what TrueBlocks is.

## Replicating The Test

Staying objective and working on permissionless public goods has always been important for us. As with our software, which requires no trust between the parties (us, our users, data consumers, you!) we want to make sure anyone can replicate this test.

In order to run the test on your own, you will need TrueBlocks Core (`chifra`) [installed](https://trueblocks.io/docs/install/install-core/). You will also need a local Ethereum archive node (Erigon or Reth recommended) or REALLY fast access to a remote node. The [DAppNode](https://dappnode.com) can be handy if you want to run the node yourself.

Next, clone [this repo](https://github.com/TrueBlocks/trueblocks-compare). Follow the instructions from the README to set up the test code.

Share your results with us, on [X/Twitter](https://twitter.com/trueblocks) or [GitHub](https://github.com/TrueBlocks/trueblocks-compare/issues/new). If you have questions, we invite you to join [our Discord server](https://discord.com/invite/kAFcZH2x7K).
