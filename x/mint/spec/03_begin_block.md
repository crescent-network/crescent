<!--
order: 3
-->

# Begin-Block

Begin block operation for the `mint` module calculates `BlockInflation` to mint coins to be sent to the mint pool and sets `LastBlockTime` value. It is worth noting that there is no inflation in genesis block as it doesn't have `LastBlockTime`.

## Inflation Calculation

At the beginning of each block, block inflation is calculated with the following calculation.

```
BlockInflation = InflationScheduleAmount * min(BlockDurationForInflation, BlockTimeThreshold) / (InflationScheduleEndTime - InflationScheduleStartTime)
```
