<!-- order: 3 -->

# State Transitions

The state transaction operations for the budget module are:

- CollectBudgets

- TotalCollectedCoins

## CollectBudgets

Get all budgets registered in `params.Budgets` and select the valid budgets to collect budgets for the block by its respective plan.

This state transition occurs at each `BeginBlock`. See [Begin-Block](04_begin_block.md).

## TotalCollectedCoins

`TotalCollectedCoins` are accumulated coins in a budget since the creation of the budget.

This state transition occurs at each `BeginBlock` at the same time as the `CollectBudgets` operation. See [Begin-Block](04_begin_block.md).
