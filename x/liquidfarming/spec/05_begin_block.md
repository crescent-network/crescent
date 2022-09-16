<!-- order: 5 -->

# Begin-Block

Begin block operations for the liquidfarming module synchronizes `liquidFarms` of the parameter and state. 

## Addition of new liquidFarm

When a new liquidFarm is added in the parameter `LiquidFarms`, the new liquidFarm is also added in the state `LiquidFarms`. 

## Deletion of liquidFarm

When a liquidFarm is deleted in the parameter `LiquidFarms`, the following opertions are done.
- Unstake all pool coins from the farming module with the reserve module account
- The ongoing auction status is set to `AUCTION_STATUS_FINISHED`.
- The new liquidFarm is also deleted in the state `LiquidFarms`. 