<!-- order: 3 -->

# State Transitions

## LiquidValidators

State transitions in liquid validators are performed on every `BeginBlock` in order to check for changes in the active liquid validator set. A validator can be `Active` or `Inactive`, the following transition occurs when the state changes

### into Active
- redelegate certain amount from all the existing Active liquid validator's LiquidTokens to newly elected validator so that every Active liquid validators have the same amount of staked tokens

### expelled from Active
- redelegate all LiquidTokens to the remaining Active liquid validators

## Liquid Staking

- `LiquidStakingProxyAcc` reserve native tokens from the sending account to delegates it
- determine the amount of bTokens are minted is based on mint rate, calculated as follows from the total supply of bTokens and net amount of native tokens.
  - `MintAmount = StakeAmount * MintRate` by NativeTokenToBToken
  - when initial liquid staking, `MintAmount == StakeAmount`
- mint the calculated amount of bTokens and send it to delegator's account
- distribute the delegation from the `LiquidStakingProxyAcc` to the all the active liquid validators according to each weights
  - internally, it calls the `Delegate` function of module `cosmos-sdk/x/staking`.
  - crumb may occur due to a decimal point error in dividing the staking amount into the weight of liquid validators, It added on first active liquid validator

## Liquid Unstaking

- The amount of native tokens returned is calculated as `UnstakeAmount = bTokenAmount / MintRate * (1 - UnstakeFeeRate)` by BTokenToNativeToken
- burn the bTokens
- `LiquidStakingProxyAcc` unbond the liquid validator's delShares by calculated native token worth of bTokens divided by current weight of liquid validators
  - the `DelegatorAddress` of the `UnbondingDelegation` would be `MsgLiquidStake.DelegatorAddress` not `LiquidStakingProxyAcc`
  - internally, it calls the `Unbond` function of module `cosmos-sdk/x/staking`, it can take up to UnbondingTime to be matured
  - crumb may occur due to a decimal error in dividing the unstaking bToken into the weight of liquid validators, it will remain in the NetAmount
  - if liquid validators or liquid tokens to unbond doesn't exist, withdraw balance of proxy `LiquidStakingProxyAcc` or need to re-try after waiting for new liquid validator to be added or unbonding of proxy account to be completed

The following operations occur when the `UnbondingDelegation` element matures:

- Unbonding of `UnbondingDelegation` is completed according to the logic of module `cosmos-sdk/x/staking`, Then the delegator of liquid staking will receive the worth of native token.
