# Matching engine

This document describes the matching algorithm of the liquidity module.
Basically, matching is done in batch-style in the liquidity module,
which means the timing of the order is not important like in other exchanges.

## Matching type

There are two types of order matching: *Single price auction* and
*Two-stage matching*.

### Single price auction

When there's no last price in the pair, *Single price auction* matching is done
for that pair.
With this type of matching, pools don't place orders on every tick but place orders at
single matching price.
The matching price is the best price where the largest portion of orders can be matched,
including pools.

### Two-stage matching

When there's last price in the pair, *Two-stage matching* is done for that pair.

The first stage is like *Single price auction*, but with the matching price set to
the last price.
If there are no orders to match at the last price, this stage is skipped.

In the second stage, buy orders with higher price are matched with sell orders with
lower price, which is typical.
The matching price depends on the direction of the price.
If the next price(last price of the batch after matching) is going to be higher than
the current last price(price increasing case), orders are matched at each sell order's price.
If the price is going to be decreased, then orders are matched at each buy order's price.
Otherwise, this stage is skipped since matching should have done in the first stage.

#### Price range

Order books are constructed from orders above and below 10%(the percentage can be
changed through a parameter change proposal) based on the pair's last price.
This restriction makes lower limit and upper limit of price within a pair.

#### Price direction

In the second stage of *Two-stage matching*, the direction of price is estimated and used
to decide matching price.
To estimate the price direction, following variables are calculated.

- $B_o$: Sum of buy orders amount over the last price
- $S_u$: Sum of sell orders amount under the last price
- $B_e$: Sum of buy orders amount at the last price
- $S_e$: Sum of sell orders amount at the last price

Then, price direction is determined as below:

- If $B_o$ > $S_u + S_e$, price is **increasing**
- If $S_u$ > $B_o + B_e$, price is **decreasing**
- Otherwise, price is **staying**

## Proportion based partial matching

When matching orders, orders in the same tick takes amount from opposite side
fairly based on their proportion and priority.
Since there is no time priority in matching, orders with larger amount have
more amount to be matched.

After distributing order amount to a tick, there can be small amount of orders
remaining because the amount isn't always divided into integers exactly.
These remaining amount is then distributed to the order with the highest priority.
If there's still a remaining amount, the next highest priority orders
take the amount.

For sell orders, if the matching amount is too small so that the order receives
zero quote coin after matching, the order is excluded from matching to prevent
unfair trade.
The proportion of other orders is re-calculated after dropping the orders.

### Order priority

There should be priorities within orders in the same tick to distribute order
amount as described as above.
There are three criteria to decide the order's priority:

1. Order amount
2. Type of orderer
3. Order(or pool) ID

The order with larger amount gets higher priority than other orders.
If the amount was same, then the type of the orderer determines the precedence.
In the liquidity module user orders always gets higher priority than pool orders.
If even the type of the orderer was same, then time priority takes part in.
It is inevitable to consider time priority even if the liquidity module tries to be
as fair as possible, because there must be a precedence between orders when distributing
remaining order amount after dividing orders based on their proportion.
If two orders are from users then the order with lower order ID takes precedence,
and if two orders are pool orders, the order from pool with lower pool ID takes precedence.
