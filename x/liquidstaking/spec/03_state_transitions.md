<!-- order: 3 -->

# State Transitions

## LiquidValidators

- State transitions in liquid validators are performed on every `EndBlock` in order to check for changes in the active validator set.
- A validator can be `Delisting` or `Active`. A validator can move directly between the states.
- into Active

  The following transition occurs when a validator elected to whiltelist by governance.

  - set `LiquidValidator.Status` to `Active`
  - redelegate certain amount from all the existing Active validator's `LiquidValidator.GetDelShares()` to newly elected validator so that every Active validators have the same amount of staked tokens
  - update the `LiquidValidator` object for this validator
  - if it exists, delete any `ValidatorQueue` record for this validator

- Active to Delisting

  The following transition occurs when a validator expelled from Active

  - redelegate all `LiquidValidator.GetDelShares()` to the remaining Active validators
  - set `LiquidValidator.Status` to `Delisting`
  - update the `LiquidValidator` object for this validator

## Staking

When a staking occurs both the validator and the liquid staking objects are affected

- determine the amount of bTokens based on native tokens delegated and the mint rate
- remove native tokens from the sending account
- distribute the `staking.Amount` from the staker's account to the all the active validator's account
- update the `LiquidValidator` object for active validators
- mint the calculated amount of bTokens and send it to liquid staker's account

## Begin Unstaking

As a part of the Complete Unstaking state transitions Begin Unstaking will be called

- determine the amount of native tokens based on amount of bTokens, mint rate and unstake fee rate
- burn the bTokens
- `LiquidStakingProxyAcc` unbond the active validator's DelShares by calculated native token worth of bTokens divided by number of active validators
  - the `DelegatorAddress` of the `UnbondingDelegation` would be `MsgLiquidStake.DelegatorAddress` not `LiquidStakingProxyAcc`

## Complete Unstaking

the following operations occur when the `UnbondingDelegation` element matures:

- Unbonding of `UnbondingDelegation` is completed according to the logic of module `cosmos-sdk/x/distribution`.
