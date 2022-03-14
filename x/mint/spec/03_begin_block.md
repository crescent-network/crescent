<!--
order: 3
-->

# Begin-Block

- BlockInflation = `InflationAmountThisPeriod * min(CurrentBlockTime-LastBlockTime,params.block_time_threshold)/(InflationPeriodEndDate-InflationPeriodStartDate)`
- BlockInflationRate = `BlockInflation * BlocksPerYear / TotalSupply`
- if no LastBlockTime(genesis block) → no inflation
- Set LastBlockTime for this block's block time on end of begin-block 