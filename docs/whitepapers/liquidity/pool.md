This document explains liquidity pools of `liquidity` module.


## Basic Liquidity Pool

A liquidity pool is a coin reserve that contains two different types of coins in a trading pair.
The trading pair has to be unique.
A liquidity provider can be anyone who provides liquidity by depositing reserve coins into the pool.
The liquidity provider earns the accumulated swap fees with respect to their pool share.
The pool share is represented as possession of pool coins.
Liquidity pools locate limit orders on each tick with order amount
which is calculated from its AMM equations.
A basic liquidity pool is a liquidity pool providing its liquidity in all price range. 
This means that users can swap coins in any price point.
When a user creates a basic liquidity pool with $X$ and $Y$, then all of this coins become the initial reserve of the created basic liquidity pool.


## Constant Product Model (CPM)

This AMM has a particularly desirable feature where it can always provide liquidity,
no matter how large the order size nor how tiny the liquidity pool.
The trick is to asymptotically increase the price of the coin as the desired quantity increases.
The term “constant” refers to the fact that any trade must change the reserves in such a way
that the product of those reserves remains unchanged (i.e. equal to a constant).

A basic liquidity pool operates based on the following equation

$$X \cdot Y = k,$$

where $k$ is a constant over any swap operations.
When a user deposits two coins to the pool, then $k$ increases while $k$ decreases when a user withdraws its reserve from the pool.

The pool price $P$ is defined as 
$$P=\frac{X}{Y},$$
 which means an instant swap ratio of $X$ and $Y$ for very small swap amount.
This pool price does not vary by deposit and withdraw operations but varies by swap operations.

## Ranged Liquidity Pool

A ranged liquidity pool is another type of liquidity pools.
Differently from a basic liquidity pool that provides liquidity in all price range, a ranged liquidity pool is a liquidity pool that provides liquidity only in a pre-defined range of price.
A ranged liquidity pool locates limit orders on each price tick only between the minimum price and the maximum price of the ranged liquidity pool.
The following equation is the baseline for the limit orders by a ranged liquidity pool:

$$(X+a) (Y+b) = k,$$

where $k$, $a$, $b$ are constants for any swap operations.
When a user deposits two coins to the pool, then $k$, $a$, and $b$ increase while $k$, $a$, and $b$ decrease when a user withdraws its reserve from the pool.

### Creation of Ranged Liquidity Pool

Let $X$ and $Y$ be the deposit amount of coins of the base coin and the quote coin to create a ranged pool.
The following steps are used to calculate the initial reserve amount $X_0$ and $Y_0$ of the base coin and the quote coin to satisfy that the initial price, $P_i$, is equal to the pool price, where the remaining amount coins are refunded to the creator.

- If the initial prices is equal to $P_{min}$, $X_0 = 0$ and $Y_0 = Y$.
- If the initial prices is equal to $P_{max}$, $X_0 = X$ and $Y_0 = 0$.
- If $\frac{X}{\sqrt{P_i} - \sqrt{P_{min}}} \left( \frac{1}{\sqrt{P_i}}-\frac{1}{\sqrt{P_{max}}} \right) \leq Y$, $X_0 = X$ and $Y_0 = \frac{X}{\sqrt{P_i} - \sqrt{P_{min}}} \left( \frac{1}{\sqrt{P_i}}-\frac{1}{\sqrt{P_{max}}} \right)$.
- If $\frac{X}{\sqrt{P_i} - \sqrt{P_{min}}} \left( \frac{1}{\sqrt{P_i}}-\frac{1}{\sqrt{P_{max}}} \right) > Y$, $X_0 = \frac{Y}{\frac{1}{\sqrt{P_i}}-\frac{1}{\sqrt{P_{max}}} } \left( \sqrt{P_i} - \sqrt{P_{min}} \right)$ and $Y_0 = Y$. 

When the pool price becomes either the minimum price or the maximum price, the ranged liquidity pool consists of only single kind of coins.

The pool price $P$ is defined as
$$P=\frac{X+a}{Y+b},$$
which means an instant swap ratio of $X$ and $Y$ for very small swap amount.
This pool price does not vary by deposit and withdraw operations but varies by swap operations.

### Derivation of $k$, $a$, and $b$ for a given $X$ and $Y$

From the equations that the ranged liquidity pool should satisfy, the following relation can be derived.

- $\sqrt{k} = \frac{X}{\sqrt{P}-\sqrt{P_{min}}} =\frac{Y}{\frac{1}{\sqrt{P}}-\frac{1}{\sqrt{P_{max}}}}$
- $a=\sqrt{k} \sqrt{P_{min}}$
- $b=\frac{\sqrt{k}}{\sqrt{P_{max}}}$

### Deposit and Withdraw Ratio

A deposit to a pool is processed in a way that the deposit ratio of $X_d$ and $Y_d$ are the same as the ratio of $X$ and $Y$ of the pool reserve, where the remaining deposit coin is refunded.

A withdraw from a pool is processed in a way that the withdraw ratio of $X_w$ and $Y_w$ are the same as the ratio of $X$ and $Y$ of the pool reserve.

## Derivation of Pool's Order Amount for Orderbook

In the following, it is explained based on a ranged liquidity pool, where those can also be applied to a basic pool by using $a=0$, $b=0$, $P_{min}=0$, and $P_{max}=\infty$.

### Derivation of Buy Amount To A Given Price

This `Buy Amount To A Given Price` is how much a liquidity pool provides as a buy order when the pool price is higher than the highest price of the order book.
In otherwords, if this amount of the pool's buy order is matched, it is targeted to the pool price becomes the given price.
With this definition, the amount $\Delta x$ for the pool's buy order can be obtained by using
$$(X+a-\Delta x)(Y+b+\Delta y)=k$$
and 
$$\frac{X+a-\Delta x}{Y+b+\Delta y} = P.$$

### Derivation of Sell Amount To A Given Price

This `Sell Amount To A Given Price` is how much a liquidity pool provides as a sell order when the pool price is lower than the lowest price of the order book.
In otherwords, if this amount of the pool's sell order is matched, it is targeted to the pool price becomes the given price.
With this definition, the amount $\Delta y$ for the pool's sell order can be obtained by using
$$(X+a+\Delta x)(Y+b-\Delta y)=k$$
and 
$$\frac{X+a+\Delta x}{Y+b-\Delta y} = P.$$


### Derivation of Buy Amount Over A Given Price

This `Buy Amount Over A Given Price` is how much a liquidity pool provides as a buy order for price greater than or equal to given price.
With this definition, the amount $\Delta x$ for the pool's buy order can be obtained by using
$$(X+a-\Delta x)(Y+b+\Delta y)=k$$
and 
$$\frac{\Delta x}{\Delta y} = P.$$


### Derivation of Sell Amount Under A Given Price


This `Sell Amount Over A Given Price` is how much a liquidity pool provides as a sell order for price less than or equal to given price.
With this definition, the amount $\Delta y$ for the pool's sell order can be obtained by using
$$(X+a+\Delta x)(Y+b-\Delta y)=k$$
and 
$$\frac{\Delta x}{\Delta y} = P.$$