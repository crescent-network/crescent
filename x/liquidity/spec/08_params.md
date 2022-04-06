<!-- order: 8 -->

# Parameters

The `liquidity` module contains the following parameters:

| Key                      | Type               | Example                                                           |
|--------------------------|--------------------|-------------------------------------------------------------------|
| BatchSize                | uint32             | 1                                                                 |
| TickPrecision            | uint32             | 3                                                                 |
| FeeCollectorAddress      | string             | cre1zdew6yxyw92z373yqp756e0x4rvd2het37j0a2wjp7fj48eevxvq303p8d    |
| DustCollectorAddress     | string             | cre1suads2mkd027cmfphmk9fpuwcct4d8ys02frk8e64hluswfwfj0s4xymnj    |
| MinInitialPoolCoinSupply | string (sdk.Int)   | "1000000000000"                                                   |
| PairCreationFee          | string (sdk.Coins) | [{"denom":"stake","amount":"1000000"}]                            |
| PoolCreationFee          | string (sdk.Coins) | [{"denom":"stake","amount":"1000000"}]                            |
| MinInitialDepositAmount  | string (sdk.Int)   | "1000000"                                                         |
| MaxPriceLimitRatio       | string (sdk.Dec)   | "0.100000000000000000"                                            |
| MaxOrderLifespan         | time.Duration      | 24hours                                                           |
| SwapFeeRate              | string (sdk.Dec)   | "0.000000000000000000"                                            |
| WithdrawFeeRate          | string (sdk.Dec)   | "0.000000000000000000"                                            |
| DepositExtraGas          | uint64 (sdk.Gas)   | 60000                                                             |
| WithdrawExtraGas         | uint64 (sdk.Gas)   | 64000                                                             |
| OrderExtraGas            | uint64 (sdk.Gas)   | 37000                                                             |

## BatchSize

Block numbers for one batch.
A BatchSize of 1 means that one batch consists of one block.

## TickPrecision

Because our DEX adopts tick system, we have to set tick precision which
determines the gap between ticks.
Default TickPrecision of 3 means that the price will be displayed from
the highest digit to the last 3 digits.

## FeeCollectorAddress

Account address for fee collecting module

## DustCollectorAddress

Account address for dust collecting.
Dust means a small amount of tokens that cannot be avoided during the
order matching process.

## MinInitialPoolCoinSupply

Minimum amount of initial pool coin minted on pool creation.
Initial pool coin supply is calculated based on two deposit coins' amount.

## PairCreationFee

Fee paid for to create a pair.
This fee prevents spamming and is collected in the fee collector.

## PoolCreationFee

Fee paid for to create a pool.
This fee prevents spamming and is collected in the fee collector.

## MinInitialDepositAmount

Minimum number of coins to be deposited to the liquidity pool upon pool creation.

## MaxPriceLimitRatio

MaxPriceLimitRatio defines the range of valid swap order price.
Currently, the MaxPriceLimitRatio is 0.1 which means that the range of
valid swap order price is (1-0.1)*lastPrice ~(1+0.1)*lastPrice of each pair.
If a swap order with price outside that range is requested,
the module will reject the order.

## MaxOrderLifespan

Since our DEX allows partial execution of swap orders,
we need a parameter for how long the remaining swap orders will remain on-chain.
Leaving it for a long time needs lots of resources, the default is set to one day.

## SwapFeeRate 

Swap fee rate for swap.
In this version, swap fees aren't paid upon swap orders directly.
Instead, pool just adjust pool's quoting prices to reflect the swap fees.

## WithdrawFeeRate  

Reserve coin withdrawal with less proportion by WithdrawFeeRate.
This fee prevents attack vectors from repeated deposit/withdraw transactions.

## DepositExtraGas

Extra gas imposed to the depositor when they deposit to a pool, since the deposit
is happened in end-block, not in the msg handler.

## WithdrawExtraGas

Extra gas imposed to the withdrawer when they withdraw from a pool, since the withdrawal
is happened in end-block, not in the msg handler.

## OrderExtraGas

Extra gas imposed to the orderer when they make an order, since the order matching
is happened in end-block, not in the msg handler.

# Global Constants

## MinCoinAmount, MaxCoinAmount

`MinCoinAmount` and `MaxCoinAmount` are defined under the `x/liquidity/amm` package.
These are the minimum and maximum coin amount accepted by the liquidity module.
Any orders with amount out of this range will be rejected in the end blocker or
the msg server.
