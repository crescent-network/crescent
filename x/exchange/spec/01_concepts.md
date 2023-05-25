<!-- order: 1 -->

# Concepts

## Exchange

### Tick

In an orderbook-based exchange, a "tick" refers to the minimum price movement
that a particular asset can make.
Each asset on the exchange is assigned a certain tick size, which is the
smallest possible increment that the price can change by.
For example, if the tick size for an asset is $0.01, the price can only move up
or down in increments of $0.01.

The tick size is an important consideration for traders because it can affect
the cost and liquidity of their trades.
For instance, if the tick size is very small, it can be easier to enter and exit
trades at specific price levels, which can increase liquidity.
On the other hand, if the tick size is too large, it can be difficult to enter
and exit trades at specific price levels, and traders may have to accept less
favorable prices in order to execute their trades.

This doesn't mean that blindly reducing the tick size is a good idea.
If we had infinite computing power, we could manage countless ticks, but in
reality, we don't need to tolerate the inefficiency of managing all the
information in ticks that we don't need.
Exchange operators can adjust the tick size for different assets based on a
variety of factors, including market volatility, trading volume, and the
specific needs of traders on their platform.
Ultimately, the tick size is a key aspect of the orderbook model, helping to
ensure that trades can be executed efficiently and accurately while maintaining
a market.

### Order types

Orderbook-based DEXs can offer traders several order types and the most common
orders are limit orders and market orders.

A limit order is an order to buy or sell an asset at a specified price or
better.
When a limit order is placed, it remains on the order book until either it is
executed or it is canceled by the trader.
If the market price reaches the specified limit price, the order will be
executed, and the trade will be made at or better than the specified price.
However, if the market price does not reach the limit price, the order will not
be filled.

On the other hand, a market order is an order to buy or sell an asset at the
best available price in the order book.
This means that market orders are executed immediately at the prevailing market
price, regardless of the specified price.
Market orders are useful when the trader wants to enter or exit a position
quickly and does not want to wait for the order to be filled at a specific
price.

In addition to the standard limit and market orders, many order book exchanges
offer a variety of advanced order types that allow users to implement more
complex trading strategies.
Two popular examples are the stop loss order and the take profit order.
A stop loss order is designed to help mitigate potential losses by automatically
triggering a market sell order when a certain price level is reached.
A take profit order is similar in nature, but instead of limiting losses, it
allows users to lock in profits by triggering a market sell order when a
specified price target is reached.
These advanced order types can be particularly useful for traders who want to
automate their trading strategies and manage risk more effectively.

### Batch (batch-sequential hybrid)

In the context of MEV(Miner Extractable Value), there is a significant
inequality between professional miners and ordinary users that cannot be
resolved.
Since miners have the power to determine the order in which transactions are
executed and included in a block, they can prioritize transactions that benefit
their own financial interests, potentially at the expense of other market
participants.
Even in traditional finance, many have long been concerned about the inequality
between specialized institutions and ordinary individuals, and some have made
significant efforts to resolve this issue.
Cryptocurrencies were actually born out of a lack of trust in traditional
financial institutions, and the spirit of decentralization that permeates this
ecosystem is also a spirit of resistance to the dominance of large institutions.
This is why we handle the matching algorithm in orderbook in a batch
format(although technically it's a hybrid that allows for sequential matching
after the batch to provide additional functionality).
We want to prevent professional participants from taking advantage of common
users and ensure that everyone has an equal footing in the market.

A one-word definition of batch matching is to eliminate the order priority of
orders in the matching phase.
Each orderer gets the same matching result regardless of the order in which one
or more orders are sorted, making MEV attacks essentially impossible.
Also, the advantages of batch matching include greater liquidity, reduced price
volatility, and increased transparency.

Most orders are processed in batches, but some orders are processed sequentially
for additional features like multi-hop.
Of course, since sequential matching happens after batch matching, it has no
effect on batch matching and will require higher fees than normal, so it should
only be used when absolutely necessary.
It may be difficult to completely prevent MEVs during the sequential match
phase, so in the future, we may decide to open up slots to users who want MEVs,
charge a large fee, and return the revenue to ecosystem participants.
