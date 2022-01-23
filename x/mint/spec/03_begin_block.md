<!--
order: 3
-->

# Begin-Block

- BlockInflation = `InflationAmountThisPeriod * min(CurrentBlockTime-LastBlockTime,10second)/(InflationPeriodEndDate-InflationPeriodStartDate)`
- BlockInflationRate = `BlockInflation * BlocksPerYear / TotalSupply`
- if no LastBlockTime(genesis block) → no inflation