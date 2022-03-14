<!--
order: 0
-->

# Concepts

## The Constant Inflation Minting Mechanism

The constant inflation is minted in proportion to the block time length according to the amount and the inflation schedules defined in params, not the existing existing dynamic inflation rate algorithm in original Cosmos-SDK.
Actual minted amount could be less than the defined `InflationSchedule.Amount` depending on the number of times blocks having the block time length over `BlockTimeThreshold` occurs and decimal loss.

