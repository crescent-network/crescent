<!-- order: 6 -->

# End-Block

Each abci end block call, the operations to update queues and active validator set changes are specified to execute

## Active validator set Changes

The active validator set is updated during this process by state transitions that run at the end of every block. Operations are as following:

- the previous validator set is compared with the new validator set:
  - missing validators begin delisting and their `Tokens` are redelegated to the remaining active validators
  - new active validators and certain amount of all the existing active validator's `Tokens` are redelegated to new validators so that every active validators have same power
  
### Active Validator Conditions
- included on whitelist
- existed valid validator on staking module ( existed, not nil del shares and tokens, valid exchange rate)

### Active -> Delisting
- when out of `Active Validator Conditions`

### Whitelisted -> Active
- when `Active Validator Conditions`

### Delisted -> Active 
- same above 

### Delisting -> Active
- same above

When a active validator is kicked out of the list, it begins the delisting process along with all its liquid staking amounts begin redelegating. At this point the validator is said to be an "delisting validator", whereby it will mature to become an "delisted validator" after the delisting period has passed

Each block the validator queue is to be checked for mature unbonding validators (namely with a completion time <= current time and completion height <= current block height). For all mature delisting validators, the `validator.Status` is switched from `types.delisting` to `types.delisted`


## Auto-Redelegation

Due to the events like slashing, tombstoning and policy related to serial redelegation, the actual weights of the delegated amount of the active validators can be slightly different from what was initially intended. Therefore, rebalancing of delegated assets is needed and it is triggered by difference of power from the intended

- calculate the current weight of each active validator and the difference between it and original target weight
- if the maximum difference exceeds `RebalancingTrigger`, asset rebalacing will be executed

## Auto-Withdraw-Re-Stake

- If the sum of the withdrawn rewards(balance) and the upcoming rewards(all delegations rewards) of `LiquidStakingProxyAcc` exceeds `RewardTrigger` of the `DelShares`, the reward is automatically withdrawn and re-stake.

