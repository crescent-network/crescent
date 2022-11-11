<!-- order: 5 -->

# Begin-Block

At each BeginBlock, the following operations occur in the `liquidfarming` module:

- Synchronizes `LiquidFarms` registered in params with the ones stored in KVStore. When a new `LiquidFarm` is added by governance proposal, the `LiquidFarm` is stored in KVStore. When a new `LiquidFarm` is added by governance proposal, a rewards auction is started at 00:00UTC in the next day.

- When an existing `LiquidFarm` is removed by the governance proposal, it first calls `Unfarm` function in the `lpfarm` module with the reserve module account to unfarm all farming coin to prevent from having farming rewards accumulated and handle the ongoing `RewardsAuction`. It refunds all placed bids and change the auction status to `AuctionStatusFinished`. Lastly, it deletes the `LiquidFarm` in the store.

- Iterates all existing `LiquidFarms` in KVStore and create `RewardsAuction` for every `LiquidFarm` if it is not created before. It there is an ongoing `RewardsAuction` for the `LiquidFarm`, then it finishes by selecting the winning bid to give them the accumulated farming rewards and calls `Farm` function in the `lpfarm` module to farm the coin of the winning bid. This action is regarded as auto compounding rewards functionality for farmers.
