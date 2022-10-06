<!-- order: 1 -->
# Concepts

This is the concept for `Liquidfarming` module in Crescent Network.

## LiquidFarming Module

The `liquidfarming` module provides a functionality for farmers to have another option to use with their liquidity pool coins in Crescent Network. 

The module allows farmers to farm their pool coin and mint a synthetic version of the pool coin called LFCoin. 
Farmers can use the LFCoin to take a full advantage of Crescent functionality, such as Boost. 
On behalf of farmers, the module stakes their pool coin to the `farming` module and receives farming rewards for every epoch. 
The module provides auto-compounding of the rewards by going through an auction process, which results in the exchange of the farming rewards coin(s) into the pool coin.

## Liquid Farm

A `liquidFarm` corresponds to one unique pool id. A `liquidFarm` stakes and unstakes the pool coins as users’ requests. When the rewards are allocated by staking the pool coins, the `liquidFarm` creates and manages an auction in order to exchange the rewards to pool coins to be staked additionally.

A `liquidFarm` can be registered to the parameter `liquidFarms` by governance for activating the liquid farming of a pool.
When the `liquidFarm` is registered to the parameter, users can request to farm their pool coins.

A `liquidFarm` can be removed in the parameter `liquidFarms` by governance for deactivating the liquid farming of a pool.
When the `liquidFarm` is removed, the module unstakes all pool coins in the module, users cannot request to farm their pool coins, but users can still request to unfarm LF coins.
In this case, the ongoing rewards auction becomes finished, all bids are refunded, and a new auction is not started.

## Liquid Farm

Once a user farms their pool coin, the user receives LFCoin instantly minted.
The following formula is used for an exchange rate of `LFCoinMint` (which is also called `mintedCoin`) when a user farms with `LPCoinFarm` (which is also called `farmingCoin`).

$$LFCoinMint = \frac{LFCoinTotalSupply}{LPCoinTotalStaked+LPCoinTotalQueued} \times LPCoinFarm,$$

where `LFCoinTotalSupply` is not zero.
`LPCoinTotalQueued` is the amount of pool coins in `FarmQueued` by the `liquidfarming` module.
If `LFCoinTotalSupply` is zero, then the following formula is applied:

$$LFCoinMint = LPCoinFarm.$$

## Liquid Unfarm

When a user unfarms their LFCoin, the module burns the LFCoin and releases the corresponding amount of pool coin.
The following formula is used for an exchange rate of `LFCoinBurn` (which is also called `burningCoin`) to receive `LPCoinUnfarm` (which is also called `unfarmedCoin`):

$$LPCoinUnfarm = \frac{LPCoinTotalStaked+LPCoinTotalQueued-CompoundingRewards}{LFCoinTotalSupply} \times LFCoinBurn,$$

if $$LFCoinBurn < LFCoinTotalSupply$$, where `CompoundingRewards` is the amount of pool coins obtained from the last rewards auction, which is in `FarmQueued`.
If $$LFCoinBurn = LFCoinTotalSupply$$, the following formula is used:

$$LPCoinUnfarm = \frac{LPCoinTotalStaked+LPCoinTotalQueued}{LFCoinTotalSupply} \times LFCoinBurn.$$

If the `liquidfarm` is not registered in the parameter, the following formula is used an exchange rate of `LFCoinBurn`:

$$LPCoinUnfarm = \frac{LPCoinTotalQueued}{LFCoinTotalSupply} \times LFCoinBurn.$$

## Farming Rewards and Auction

On behalf of users, the module stakes their pool coins and claims farming rewards.
In order to exchange the rewards coin(s) into the pool coin to be additionally staked for farming, the module creates an auction to sell the rewards that will be received at the end of the epoch.
Note that the exact amount of the rewards being auctioned is not determined when the auction is created, but will be determined when the auction ends.
The amount of the rewards depends on the total amount of staked pool coins and the `liquidfarming` module’s staked pool coin, which can be varied during the epoch.
Therefore, a bidder to place a bid for the auction should be aware of this uncertainty of the rewards amount.

## Bidding for Auction and Winning Bid

A bidder can place a bid with the pool coin, which is the paying coin of the auction.
A bidder only can place a single bid per auction of a liquid farm.
The bid amount of the pool coin must be higher than the current winning bid amount that is the highest bid amount of the auction at the moment.
The bidder placing the bid with the highest amount of the pool coin becomes the winner of the auction and will takes all the accumulated rewards amount at the end of the auction.