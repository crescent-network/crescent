<!-- order: 3 -->

# State Transitions

## LiquidValidators

State transitions of liquid validators are performed on every `BeginBlock` to keep in track of any changes in active liquid validator set. The following state transition occurs when a validator is added or removed from an active liquid validator set.

### When new `LiquidValidator` is added

- Redelegation of `LiquidTokens` occurs from the existing active liquid validator set to newly added validators. This process rebalances their delegation shares that corresponds to their weight.

### When `LiquidValidator` becomes inactive

- Redelegation of the inactive liquid validator's `LiquidTokens` occurs to the remaining active liquid validators. If redelegation fails due to restrictions exist in `staking` module, then the module unbonds their delegation shares and remove the liquid validator from the store.

## Liquid Staking

- Reserve native token to `LiquidStakingProxyAcc`
- Mint the amount of `bToken` that is based on `MintRate`
  - Initial minting amount is the same as liquid staking amount
- Send the minted `bToken` amount to the liquid delegator
- `LiquidStakingProxyAcc` delegates delegation shares to all active liquid validators that correspond to their weight
  - Internally, the module calls `Delegate` function in `staking` module
  - First active liquid validator may receive slightly more delegation shares due to some crumb occuring from division

## Liquid Unstaking

- Calculate the unbonding amount from the requesting `bToken` 
- Burn the requesting `bToken`
- `LiquidStakingProxyAcc` unbonds the 
  - Internally, the module calls `Unbond` function in `staking` module and it takes `UnbondingTime` to be matured
  - `LiquidStakingProxyAcc` transfers an ownership of `UnbondingDelegation` to the liquid delegator. The liquid delegator is expected to receive unbonding amount after `UnbondingDelegation` is matured.
  - Crumb may occur due to decimal loss from division and it remains in `NetAmount`
  - Try to withdraw unstaking amount from `LiquidStakingProxyAcc` balance when 1) liquid validators don't have enough `LiquidTokens` to unbond and 2) there is no active liquid validator in the network. In case `LiquidStakingProxyAcc` doesn't have enough balance, liquid delegator must wait until active liquid validators are newly added or the proxy account gets sufficient balance that will be automatically filled when unbonding period is complete.