<!-- order: 5 -->

# Begin-Block

Each abci begin block call, the operations to update active liquid validator set changes are specified to execute

## Active Liquid Validator set Changes

The active liquid validator set is updated during this process by state transitions that run at the begin of every block. Operations are as following:

- the previous validator set is compared with the new validator set:
    - missing validators begin delisting and their `Tokens` are redelegated to the remaining active liquid validators
    - new active liquid validators and certain amount of all the existing active liquid validator's `Tokens` are redelegated to new validators so that every active liquid validators have same power

### Active Conditions
- included on whitelist
- existed valid validator on staking module ( existed, not nil DelShares and tokens, valid exchange rate)
- not tombstoned

### Active -> Inactive
- when out of the `Active Conditions`

When active liquid validator is kicked out of the list, it begins the delisting process along with all its liquid staking amounts begin redelegating. At this point the validator is said to be an "Inactive liquid validator", whereby it will mature to be removed the LiquidValidator object after the redelegation/unbonding period has passed and no DelShares

### Whitelisted -> Active
- when meet the `Active Conditions`

### Inactive -> Active
- when meet again the `Active Conditions` before removed

### Inactive -> Removed
- no DelShares(redelegation completed) and out of the `Active Conditions`

## Auto-Redelegation

Due to the events like slashing, tombstoning, become Inactive and policy related to serial redelegation, the actual current weights of the delegated amount of the active liquid validators can be slightly different from what was target weight intended. Therefore, rebalancing of delegated assets is needed, and it is triggered by difference of power from the intended

- calculate the current weight of each active liquid validator and the difference between it and target weight
- if the maximum difference and needed each redelegation amount exceeds `RebalancingTrigger`, asset rebalacing will be executed

## Auto-Withdraw-Re-Stake

- If the sum of balance(the withdrawn rewards, crumb) and the upcoming rewards(all delegations rewards) of `LiquidStakingProxyAcc` exceeds `RewardTrigger` of the total `DelShares`, the reward is automatically withdrawn and re-stake according to the weights.

