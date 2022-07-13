<!-- order: 1 -->

# Concepts

## Liquidity Module

The liquidity module is a module that can be used on any Cosmos SDK-based application.
The liquidity module implements a decentralized exchange (DEX) that serves
liquidity providing and coin trading functions.
Anyone can create a liquidity pool with a pair of coins,
provide liquidity by depositing reserve coins into the liquidity pool,
and trade coins using the liquidity pool.
All the logic is designed to always protect the pool investors.

## Orderbook

Unlike general DEXs, we introduced the concept of an orderbook for versatility and visibility.
Users can select trading methods in traditional financial markets,
such as limit orders and market orders, as well as normal swap orders in DeFi ecosystem.
In that context, the liquidity pool will also break away from the existing passive role
and participate in the transaction as a single entity.
The matchable orders stacked in the orderbook are matched to the most reasonable price
by mathematical logic and unmatched requests are removed or remained according to user’s choice.

Read more about matching process in the [Liquidity pool white paper](../../../docs/whitepapers/liquidity/matching.md).

## Tick System

We introduce tick system in DEX, alongside with enabling order book feature.
This is a natural consequence because most exchanges with order book have its own tick system.
The size of tick is configured by using the `TickPreicision` governance parameter.

## Liquidity Pool

A liquidity pool is a coin reserve that contains two different types of coins in a trading pair.
The trading pair has to be unique.
A liquidity provider can be anyone who provides liquidity by depositing reserve coins into the pool.
The liquidity provider earns the accumulated swap fees with respect to their pool share.
The pool share is represented as possession of pool coins.
Liquidity pools locate limit orders on each tick with order amount
which is calculated from its AMM equations.

Read more about liquidity pool in the [Liquidity pool white paper](../../../docs/whitepapers/liquidity/pool.md).

## Constant Product Model (CPM)

This AMM has a particularly desirable feature where it can always provide liquidity,
no matter how large the order size nor how tiny the liquidity pool.
The trick is to asymptotically increase the price of the coin as the desired quantity increases.
The term “constant” refers to the fact that any trade must change the reserves in such a way
that the product of those reserves remains unchanged (i.e. equal to a constant).

## Batch Execution

The liquidity module uses a batch execution methodology.
Deposits and withdrawals are accumulated in a liquidity pool and
swap orders are accumulated in a market pair
for a pre-defined period that is one or more blocks in length.
Orders are then added to the orderbook and executed at the end of the batch.
The size of each batch is configured by using the `BatchSize` governance parameter.

## Escrow Process

The liquidity module uses a module account that acts as an escrow account.
The module account holds and releases the coin amount during batch execution.

## Refund

The liquidity module has a refunding logic when deposits, withdrawals and orders
were not successfully executed.

## Fees

You set liquidity module fees for pair creation, pool creation, withdrawal and swap.
These fees can be updated by the governance proposal.

### PoolCreationFee

The liquidity module pool creation fee set by the `PoolCreationFee` parameter
is paid on pool creation.
The purpose of this fee is to prevent users from creating useless pools and
making limitless transactions.
The funds from this fee go to the `FeeCollectorAddress`.

### WithdrawalFeeRate

The liquidity module has `WithdrawFeeRate` parameter that is paid upon withdrawal.
The purpose of this fee is to prevent liquidity providers from getting out of the pool.

### SwapFeeRate

Swap fees aren't paid upon swap orders directly.
Instead, the pool just adjust pool's quoting prices to reflect the swap fees.
In other words, the pool provides liquidity at a price that is slightly higher(or lower) than
what it can do, and the profit from the transaction made at this price is accumulated
in the pools and are shared among the liquidity providers.
In short, fee rate concept could be replaced by "QuoteSpread".
