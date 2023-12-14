# TrueBlocks Comparison with EtherScan

- [TrueBlocks Comparison with EtherScan](#trueblocks-comparison-with-etherscan)
  - [The Problem](#the-problem)
  - [What We Compared](#what-we-compared)
  - [The Results](#the-results)
    - [Bug in EtherScan related to Uncles](#bug-in-etherscan-related-to-uncles)
  - [Why Does This Matter?](#why-does-this-matter)
  - [Replicating The Test](#replicating-the-test)
  - [What is an Appearance?](#what-is-an-appearance)

## The Problem

TrueBlocks' command line tool `chifra scrape` produces the [Unchained Index](https://trueblocks.io/papers/2023/specification-for-the-unchained-index-v2.0.0-release.pdf). The best _index of appearances_ that we know of. The Unchained Index includes a record each time an address appears anywhere on the chain. No other indexer, to our knowledge, is as complete.

In this article, we compare the TrueBlocks' indexer with EtherScan's web 2.0 API. Spoiler alert: TrueBlocks wins.

## What We Compared

We queried both TrueBlocks and EtherScan for "appearances" for about 1,000 randomly-selected addresses. The results are presented below.

There is an important distinction to be made between TrueBlocks and EtherScan. TrueBlocks is local-first running against a local Ethereum archive node. This means TrueBlocks does not rate limit, nor does it cost anything to operate, nor does it paginate. EtherScan, on the other hand, is a web 2.0 API--they have no choice but to rate limit, charge for access, and paginate.

We think these differences are the reason for the surprising results presented below.

## The Results

We checked 1,000 randomly selected addresses against both Etherscan and TrueBlocks. The data produced is available in `data.tar.gz`.

Of those 1,000 addresses, 328 were discarded because they had more than 5,000 appearances. Etherscan's free service limits its return to less than 10,000 records. We wanted to stay as far away from that limit as possible. (Plus, waiting for more than 5,000 records from Etherscan was way too slow. TrueBlocks can easily return 100,000s of records for any address almost instaneously.)

<img src="../assets/results1.png" alt="Results" width="50%" height="auto">

Of the remaining **672** addresses:

- **555** (**83%**) addresses had appearances found by TrueBlocks but not Etherscan. That's **166,481** more appearances!
- **NO** appearances were found by EtherScan that were not also found by TrueBlocks
- for **15** addresses, Etherscan found 364 different appearances than TrueBlocks, but in all 15 cases, the difference was due to a bug in EtherScan. (See below.)
- all **672** addresses had appearances in common. This constituted **282,478** appearances

We recognize that the huge number of additional appearances found by TrueBlocks seems like a mistake. But one needs to realize that TrueBlocks looks for more than just a small set of known behaviours (such as `Transfers`). TrueBlocks looks everywhere. In particulate, TrueBlocks looks in:

- the transaction's `input` data
- the `topics` of the transaction's logs
- the `data` field of the transaction's logs
- the `data` and `output` field of the transaction's traces

Here's the breakout of where those **166,481** additional appearances were found.

<img src="../assets/results2.png" alt="Results" width="60%" height="auto">

As you can see, TrueBlocks and Etherscan find mostly the same things for `from`, `to`, and some log topics. But TrueBlocks finds a lot more in the `input` data, the `data` field of the logs, and the `data` and `output` fields of the traces. Only the first appearances of each type are counted. For example, if an address appears in the `input` data of a transaction and then again in the `log_topic` field of a log, only the `input` data appearance is counted.

### Bug in EtherScan related to Uncles

In the **672** addresses we searched, we found **15** addresses where EtherScan found appearances that TrueBlocks did not. In all cases, however,the difference was due to a bug in EtherScan related to uncles. The bug is that EtherScan returns the block number when the uncle was "located". TrueBlocks returns the block number in which the uncle reward was credited to the address's account. We know this because we ran the following analysis on all **364** appearances of this issue.

First, we extracted just the block number from the `etherscan` files. We then calculated 1 block prior to that block number (P) and seven blocks after that block number (A). We then ran:

```[shell]
chifra state --parts balance P-A <address> --changes
```

which uses another one of the `chifra` tools to extract the balances for the given address at the given blocks. For example, for address `0x3f98e477a361f777da14611a7e419a75fd238b6b`, Etherscan reports the following appearances:

```[shell]
485,uncle
940,uncle
1114,uncle
...
```

This command

```[shell]
chifra state --parts balance 484-492 0x48040276e9c17ddbe5c8d2976245dcd0235efa43
```

returns

```[shell]
blockNumber,address,balance
484,0x48040276e9c17ddbe5c8d2976245dcd0235efa43,0
485,0x48040276e9c17ddbe5c8d2976245dcd0235efa43,0
486,0x48040276e9c17ddbe5c8d2976245dcd0235efa43,0
487,0x48040276e9c17ddbe5c8d2976245dcd0235efa43,3750000000000000000
488,0x48040276e9c17ddbe5c8d2976245dcd0235efa43,3750000000000000000
489,0x48040276e9c17ddbe5c8d2976245dcd0235efa43,3750000000000000000
490,0x48040276e9c17ddbe5c8d2976245dcd0235efa43,3750000000000000000
491,0x48040276e9c17ddbe5c8d2976245dcd0235efa43,3750000000000000000
```

As you can see, EtherScan reports the uncle block at block 485. However, the uncle reward was not credited to the miner's account until block 487. TrueBlocks reports the uncle at block 487.

In all 364 cases, the block EtherScan reports as the uncle block is technically correct. However, the uncle reward was not credited to the miner's account until a few blocks later. In each case, that block was the block that TrueBlocks reported. EtherScan got it wrong. Unless you want to lean on a technicality. I would argue that a change in balance of an account is the correct place to note in an address's history. I'll leave it up to EtherScan to decide if they want to fix this bug.

Total number of place where EtherScan legitmately found more appearances than TrueBlocks: **ZERO**!

## Why Does This Matter?

Blockchains are perfect, 18-decimal place accurate accounting systems. Every 12 seconds, they come to balance on many hundreds of millions of accounts. That's true on-chain.

The fact that even Etherscan, our industries leading data provider can't get it right, is imporant because blockchains should balance perfectly off-chain as well as on-chain. What the hell are we even building otherwise if we can't account for every single wei for every single account. That's what TrueBlocks does. That's what TrueBlocks is.

## Replicating The Test

Staying objective and working on permissionless public goods has always been important for us. As with our software, which requires no trust between the parties (us, our users, data consumers, you!) we want to make sure anyone can replicate this test.

In order to run the test on your own, you will need TrueBlocks Core (`chifra`) [installed](https://trueblocks.io/docs/install/install-core/). You will also need a local Ethereum archive node (Erigon or Reth recommended) or REALLY fast access to a remote node. The [DAppNode](https://dappnode.com) can be handy if you want to run the node yourself.

Next, clone [this repo](https://github.com/TrueBlocks/trueblocks-compare). Follow the instructions from the README to set up the test code.

Share your results with us, on [X/Twitter](https://twitter.com/trueblocks) or [GitHub](https://github.com/TrueBlocks/trueblocks-compare/issues/new). If you have questions, we invite you to join [our Discord server](https://discord.com/invite/kAFcZH2x7K).

## What is an Appearance?

"Appearances" are seemingly simple. For any address, the address's list of appearances is a list of blocknumber.transactionId pairs noting whereever the address appears on the chain.

For example, the first three appearances for `trueblocks.eth` are:

```[csv]
blockNumber,transactionIndex
8854723,61
8856290,62
8856316,91
```

Easy enough. Just look at `from`, `to`, `log topic 0` and a few other places. That's what most indexers do. But as we've demonstrated above, there's way more to the story. Please see a very detailed discussion in the [Specification of the Unchained Index](https://trueblocks.io/papers/2023/specification-for-the-unchained-index-v2.0.0-release.pdf).
