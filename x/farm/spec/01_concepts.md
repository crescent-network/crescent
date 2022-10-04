<!-- order: 1 -->

# Concepts

## Farming Module V2

The farming module is a module that implements farming functionality that keeps
track of staking and provides farming rewards to farmers.
A primary use case of this module is to provide incentives for liquidity pool
investors for their pool participation.
The main differences between Farming V1 and V2 are:

1. The target of the farming plan
2. The shortened farming epoch

These upgrades are designed in such a way that it can more easily guarantee
convenience in operation and fairness among farmers.
Furthermore, as the DeFi service evolves, the possibility of collaboration
between multiple modules or projects will continue to be important, and this
upgrade will make such compatibility easier.

### Target changes from pool to pair

Most DEXs using AMM represented by the constant product model (CPM) have
operated only one pool for one token pair.
However, pools operated by CPM inevitably spread liquidity over a very wide
price range from 0 to infinity, and thus have been criticized for significantly
lowering capital efficiency.
Later, to solve such problems, a new AMM appeared in the form of providing
liquidity only within a predetermined range(concentrated liquidity, trident,
ranged liquidity, etc).

The problem is that if this happens, different pools can be created in one pair,
and there is no choice but to think about how to distribute the farming reward
to the pools in the most fair and convenient way.
Of course, there are ways in which the team operating the AMM DEX decides
arbitrarily or by collecting the opinions of the community, but this is not
fair, confused and even wastes energy and resources.

Therefore, instead of creating and managing a farming plan for all newly created
pools in the future, we would like to present a new type of farming module in
which the target of the farming plan changes from pool to the corresponding
token pair and all pools that share token pair as an underlying asset receive
farming reward according to its liquidity contribution.

### Liquidity

The part that investors should be most concerned about in pool investing is the
so-called impermanent loss(IL).
Leveraged pools such as ranged pools have a more risky IL curve instead of
providing liquidity to the effective range much more efficiently.
Therefore, it is natural that investors of ranged pool who took the risk and
improved the capital efficiency of the DEX as a whole should receive more
rewards than the basic pool investors.
Perhaps the fairest way is to actually calculate and reflect how much liquidity
a ranged pool provides around the current price compared to a basic pool of the
same size. As we know, in the case of AMM using the constant product model, the
amount of liquidity that the pool can provide is equal to the geometric mean of
the reserves of two tokens in the pool.

$$
Liquidity\space of\space basic\space pool : L = \sqrt{XY}
$$

$$
Liquidity\space of\space ranged\space pool : L = \sqrt{(X+a)(Y+b)}
$$

So we can calculate reward weight by using above liquidity equation.
It can be standardized by dividing with the sum of liquidity for all LP with the
same token pair, so that the sum of reward weight become 1.

$$
W_{i} = \frac{L_{i}}{\sum_{k=1}^n L_k}
$$

### Farming Epoch

Basically, in Farming V2, the rewards distributing epoch is designed as 1 block.
A short reward epoch has a big advantage.
First, because the rewards are accumulated in near real time, the rewards are
distributed more fairly and further, the freedom of choice for investors is
expanded.
Investors can take their own rewards no matter how short the pool investing
period is, and can respond to market conditions by withdrawing the assets in the
pool at any time.
A second, much greater advantage is that it allowing collaboration with other
modules or projects due to the overall farming process become much more
flexible.
For example, in a situation where farming rewards are distributed once a day,
when considering compatibility with other modules, flexible connection of
functions is difficult because the farming epoch (1 day) must be filled to
compensate for the loss of farming rewards.
Immediate interaction between modules is also not possible. In addition, in
order to neutralize the economic advantages and disadvantages between new pool
investors, existing pool investors, and withdrawers through other modules,
control logics may be attached here and there, resulting in an unintuitive
collaboration.
