# TrueBlocks Comparison with EtherScan

- [TrueBlocks Comparison with EtherScan](#trueblocks-comparison-with-etherscan)
  - [The Problem](#the-problem)
  - [What We Compare](#what-we-compare)
  - [The Results](#the-results)
    - [Bug in EtherScan related to Uncles](#bug-in-etherscan-related-to-uncles)
  - [Why Does It Matter?](#why-does-it-matter)
  - [Replicating The Test](#replicating-the-test)

## The Problem

TrueBlocks command line tool called `chifra` produces what we call an _index of appearances_, because each time an address appears on chain, it is extracted and saved. This way we can find all transactions for given address and to our knowledge every other indexer creates a similar database.

But can TrueBlocks find more appearances than EtherScan?

## What We Compare

We will ask both TrueBlocks' `chifra` and EtherScan for appearances of a number of random addresses.

There is an important difference between TrueBlocks and EtherScan: TrueBlocks is local software, running against user's own local node. This means that TrueBlocks do not need to care about rate limits, request costs or results pagination.

EtherScan on the other hand is a service - they have to pay for every request they get and hence need to impose limits. We need to keep those in mind in order to get a complete data set to compare.

## The Results

We checked 1000 randomly selected addresses. Of those, we discarded 328 which had more than 5,000 appearances. Etherscan, as mentioned in _What We Compare_ section is a service and does not return more than 10,000 records for any given address. We wanted to stay clear of that limit to avoid confusion.

It gave us **672** addresses to compare.
- for **83%** of addresses (555) TrueBlocks found **166,481** appearances more
- for **no** addresses did EtherScan find more appearances than TrueBlocks
- both has found the same 282,478 appearances

Initially it seemed that for 2% (15) addresses EtherScan returned 364 appearances more. However it was due to EtherScan bug, see _Bug in EtherScan related to Uncles_ below.

We understand that the number of appearances found by TrueBlocks is huge and may look like a mistake. But one needs to remember that TrueBlocks is not only looking for token transfers. In huge percentage of the cases, the difference was due to the fact that TrueBlocks finds appearances in one of four places:
- in the transaction's input data
- in the topics of the transaction's logs
- in the data field of the transaction's logs
- in the data field of the transaction's traces

This is totally to be expected as this is exactly where TrueBlocks' Unchained Index looks where other systems do not. In fact, we would expect TrueBlocks to find more appearances than any other system that does not look in these places.

![Results](./assets/results3.png)

As you can see. We find things in places others don't look.

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
