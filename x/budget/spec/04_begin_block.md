<!-- order: 4 -->

# Begin-Block

At the beginning of each block, the `BeginBlock`, the budget module gets all budgets that are registered in `params.Budgets`, then selects the valid budgets to collect budgets for the block by its respective plan (defined rate, source address, destination address, start time, and end time). Then, distributes the collected amount of coins from `SourceAddress` to `DestinationAddress`.

## Workflow

1. Get all the budgets registered in `params.Budgets` and proceed with the valid and unexpired budgets. Otherwise, exit and wait for the next block. 

2. Create a map by `SourceAddress` to handle the budgets for the same `SourceAddress` based on the same balance when calculating rates for the same `SourceAddress`.

3. Collect budgets from `SourceAddress` and send amount of coins to `DestinationAddress` relative to the rate of each budget`.

4. Cumulate `TotalCollectedCoins` and emit events about the successful budget collection for each budget.

