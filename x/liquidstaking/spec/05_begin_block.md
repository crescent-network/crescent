<!-- order: 5 -->

# Begin-Block

At the beginning of every block, the `liquidstaking` module operates the following executions.

## Update Liquid Validator Set Changes

### New Liquid Validator

New liquid validator can be added and updated through governance process. When a new whitelisted validator is added, they become one of the active liquid validators as long as they meet the active conditions. The module redelgates the exiting `LiquidTokens` from an active liquid validator set to newly added liquid validators so that every liquid validator has the exact amount of tokens that correspond to their weight.

### Whitelisted -> Add Liquid Validator

When whitelisted validators meet the `Active Conditions`, certain amount of all the existing active liquid validator's LiquidTokens are redelegated to new active liquid validators so that every active liquid validators have balanced LiquidTokens as each weight
    
### Active -> Inactive

When out of the `Active Conditions` When active liquid validator is out of the Active Conditions, it begins the rebalancing process along with all its liquid staking amounts begin redelegating, whereby it will mature to be removed the LiquidValidator object after the redelegation/unbonding period has passed and no delShares

### Inactive -> Active

When meet again the `Active Conditions` before removed, it begins the rebalancing process

### Inactive -> Remove Liquid Validator

No delShares by redelegation, unbonding completed and out of the `Active Conditions`

## Rebalancing (Auto-Redelegation)

Due to the events like slashing, tombstoning, becoming inactive and policy related to serial redelegation, the actual current weights of the delegated amount(LiquidTokens) of the active liquid validators can be slightly different from what was target weight intended. Therefore, rebalancing of delegated assets is needed, and it is triggered by difference of power from the intended

- calculate the current weight of each active liquid validator's LiquidTokens and the difference between it and derived weight by status of each liquid validator
- if the maximum difference exceeds `params.RebalancingTrigger` ratio of total LiquidTokens, asset rebalacing will be executed by calling `BeginRedelegation` function of `cosmos-sdk/x/staking` module
- Depending on the restriction of the staking module, some redelegation may fail, which will be retried in the next rebalancing process.

## Auto-Withdraw-Re-Stake

- If the sum of balance(the withdrawn rewards, crumb) and the upcoming remaining rewards(all delegations rewards) of `LiquidStakingProxyAcc` exceeds `params.RewardTrigger` of the total LiquidTokens, the reward is automatically withdrawn and re-stake to active liquid validators according to each weight.

