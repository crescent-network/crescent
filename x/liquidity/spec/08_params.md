<!-- order: 7 -->

# Parameters

The `liquidity` module contains the following parameters:

| Key                     | Type                | Example                                                           |
|-------------------------| ------------------- | ----------------------------------------------------------------- |
| BatchSize               | uint32              | 1
| TickPrecision           | uint32              | 3
| FeeCollectorAddress     | string              |
| DustCollectorAddress    | string              |
| InitialPoolCoinSupply   | string (sdk.Int)    | "1000000000000"
| PairCreationFee         | string (sdk.Coins)  | [{"denom":"stake","amount":"1000000"}]
| PoolCreationFee         | string (sdk.Coins)  | [{"denom":"stake","amount":"1000000"}]
| MinInitialDepositAmount | string (sdk.Int)    | "1000000"
| MaxPriceLimitRatio      | string (sdk.Dec)    | "0.100000000000000000"
| MaxOrderLifespan        | time.Duration       | 24hours
| SwapFeeRate             | string (sdk.Dec)    | "0.000000000000000000"
| WithdrawFeeRate         | string (sdk.Dec)    | "0.000000000000000000"

## BatchSize

Block numbers for one batch. A BatchSize of 1 means that one batch consists of one block.

## TickPrecision

Because our DEX adopt tick system, we have to set tick precision which determine the gap between ticks. Default TickPrecision of 3 means that the price will be displayed from the highest digit to the last 3 digits.

## FeeCollectorAddress

Account address for fee collecting module

## DustCollectorAddress

Account address for dust collecting. Dust means a small amount of tokens that cannot be avoided during the order matching process.

## InitialPoolCoinSupply

Initial mint amount of pool coin on pool creation

## PairCreatinoFee

Fee paid for to create a pair. This fee prevents spamming and is collected in the community pool through fee collector

## PoolCreatinoFee

Fee paid for to create a pool. This fee prevents spamming and is collected in the community pool through fee collector

## MinInitialDepositAmount

Minimum number of coins to be deposited to the liquidity pool upon pool creation.

## MaxPriceLimitRatio

MaxPriceLimitRatio defines the range of valid swap order price. Currently, the MaxPriceLimitRatio is 0.1 which means that the range of valid swap order price is (1-0.1)*lastPrice ~(1+0.1)*lastPrice of each pair. If a swap order with price outside that range is requested, the module will reject the order.

## MaxOrderLifespan

Since our DEX allows partial execution of swap orders, we need a parameter for how long the remaining swap orders will remain on-chain. Leaving it for a long time needs lots of resources, the default is set to one day.

## SwapFeeRate 

Swap fee rate for swap. In this version, swap fees aren’t paid upon swap orders directly. Instead, pool just adjust pool's quoting prices to reflect the swap fees.

## WithdrawFeeRate  

Reserve coin withdrawal with less proportion by WithdrawFeeRate. This fee prevents attack vectors from repeated deposit/withdraw transactions.
