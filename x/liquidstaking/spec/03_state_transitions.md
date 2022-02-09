<!-- order: 3 -->

# State Transitions

## LiquidValidators

- State transitions in liquid validators are performed on every `BeginBlock` in order to check for changes in the active liquid validator set.
- A validator can be `Active` or `InActive`. A validator can move directly between the states.
- into Active

  The following transition occurs when a validator elected to whiltelist by governance and meet the Active Conditions.

  - redelegate certain amount from all the existing Active liquid validator's `LiquidValidator.GetLiquidTokens()` to newly elected validator so that every Active liquid validators have the same amount of staked tokens

- expelled from Active

  The following transition occurs when a validator expelled from Active

  - redelegate all `LiquidValidator.GetLiquidTokens()` to the remaining Active liquid validators

## Staking

When a staking occurs both the validator and the liquid staking objects are affected

- determine the amount of bTokens based on native tokens delegated and the mint rate
- `LiquidStakingProxyAcc` reserve native tokens from the sending account to delegates it
- distribute the `staking.Amount` from the `LiquidStakingProxyAcc` to the all the active liquid validator's account according to the weights 
- mint the calculated amount of bTokens and send it to liquid staker's account

## Begin Unstaking

As a part of the Complete Unstaking state transitions Begin Unstaking will be called

- determine the amount of native tokens based on amount of bTokens, mint rate and unstake fee rate
- burn the bTokens
- `LiquidStakingProxyAcc` unbond the active liquid validator's DelShares by calculated native token worth of bTokens divided by current weight of active liquid validators
  - the `DelegatorAddress` of the `UnbondingDelegation` would be `MsgLiquidStake.DelegatorAddress` not `LiquidStakingProxyAcc`

## Complete Unstaking

the following operations occur when the `UnbondingDelegation` element matures:

- Unbonding of `UnbondingDelegation` is completed according to the logic of module `cosmos-sdk/x/distribution`, Then liquid staker will receive the worth of native token.
