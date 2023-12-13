# TrueBlocks Comparison with EtherScan

- [TrueBlocks Comparison with EtherScan](#trueblocks-comparison-with-etherscan)
  - [The Problem](#the-problem)
  - [What We Compared](#what-we-compared)
  - [The Results](#the-results)
    - [Bug in EtherScan related to Uncles](#bug-in-etherscan-related-to-uncles)
  - [Why Does It Matter?](#why-does-it-matter)
  - [Replicating The Test](#replicating-the-test)
  - [What is an Appearance?](#what-is-an-appearance)

## The Problem

TrueBlocks' command line tool `chifra scrape` produces what we call the Unchained Index (or sometimes an _index of appearances_). This index records each time an address appears anywhere on the chain. In this way, we can find complete transaction histories. No other indexer, to our knowledge, is as complete.

In this article, we compare the TrueBlocks' indexer with EtherScan's web 2.0 API.

## What We Compared

We queried both TrueBlocks and EtherScan for "appearances" for about 1,000 randomly-selected addresses. The results are presented below.

There is an important distinction to be made between TrueBlocks and EtherScan. TrueBlocks is local-first running against a local Ethereum node. This means TrueBlocks does not rate limit, nor does it generate any cost, nor paginates. EtherScan, on the other hand, is a web 2.0 API--they have no choice but to rate limit. Access to some of their data is expensive. And, they paginate everything.

Keep this in mind as we proceed.

## The Results

We checked 1,000 randomly selected addresses. The data produced is available in `data.tar.gz`.

Of these 1,000 addresses, we discarded 328 because they had more than 5,000 appearances. Etherscan's free service does not return more than 10,000 records for any given address. So we stayed well short of that limit. TrueBlocks easily returns 100,000s of records nearly instantaneously.

![Results](../assets/results2.png)

We were left with **672** addresses to compare.

- for **83%** (555) of those addresses, TrueBlocks found found more appearances than Etherscan. **166,481** more!
- for **NO** addresses did EtherScan find more appearances than TrueBlocks
- **282,478** appearances showed up in both data sets

In our initial results, EtherScan seemingly return more appearances than TrueBlocks (15 addresses--364 appearances). However we found that this was due to a bug in EtherScan, see _EtherScan's uncles bug_ below.

We recognize that the huge number of additional appearances found by TrueBlocks seems like a mistake. But one needs to remember that TrueBlocks not only looks for token transfers. In a large percentage of cases, TrueBlocks finds appearances where EtherScan doesn't look:

- in the transaction's input data
- in the topics of the transaction's logs
- in the data field of the transaction's logs
- in the data field of the transaction's traces

This is to be expected as these four places are exactly where the Unchained Index looks where other systems do not. In fact, TrueBlocks always finds more appearances than others system because they don't even look.

![Results](../assets/results3.png)

As you can see. We find things others don't.

### Bug in EtherScan related to Uncles

In the 672 addresses we searched, we found 15 where EtherScan found more appearances than TrueBlocks. In all 15 cases, however,the difference was due to a bug in EtherScan related to uncles. The bug is that EtherScan returns the block number when the uncle was found. TrueBlocks returns the block number in which the uncle reward was credited to the miner's account. We know this because we ran the following analysis on all 364 appearances of this issue.

First, we extracted just the block number from the `etherscan` files. We then calculated 1 block prior to that block number (P) and seven blocks after that block number (A). We then ran:

```[shell]
chifra state --parts balance P-A <address> --changes
```

which extracts the balances for the given address at the given blocks. For example, for address `0x3f98e477a361f777da14611a7e419a75fd238b6b`, Etherscan reports the following appearances:

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

## Why Does It Matter?

Blockchains are perfect, 18-decimal place accurate accounting systems. Every 12 seconds, they come to balance on many hundreds of millions of accounts. That's true on-chain.

It's imporant because we think it should be not only true, but expected, off-chain as well. We think that if you're going to build a system that claims to be a blockchain explorer, you should be able to account for every wei of every account. That's what TrueBlocks does. That's what TrueBlocks is.

## Replicating The Test

Staying objective and working on permission-less public goods has always been important for us. And like with our software, which require no trust between the parties (us - software authors and users - data consumers) we want anyone to be able to replicate the test.

In order to run the test on your own, you will need TrueBlocks Core (`chifra`) [installed](https://trueblocks.io/docs/install/install-core/). You will also need a local Ethereum archive node (Erigon or Reth recommended). [DAppNode](https://dappnode.com) can be handy if you need to run the node on another machine.

Next, clone [compare repository](https://github.com/TrueBlocks/trueblocks-compare). Follow the instructions from the README to set up the test code.

Please share the results with us, on [X/Twitter](https://twitter.com/trueblocks) or [GitHub](https://github.com/TrueBlocks/trueblocks-compare/issues/new). If you have questions, we invite you to join [our Discord server](https://discord.com/invite/kAFcZH2x7K).

## What is an Appearance?

"Appearances" are seemingly simple. For any address, its appearances are whereever it appears on the chain.

For example, the first three appearances for `trueblocks.eth` are:

```[csv]
blockNumber,transactionIndex
8854723,61
8856290,62
8856316,91
```
