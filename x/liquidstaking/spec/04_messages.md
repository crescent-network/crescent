<!-- order: 4 -->

# Messages

## MsgLiquidStake

Within this message the delegator provides coins, and in return receives some amount bTokens

This message is expected to fail if:

- the active validator does not exist
- the `Amount` `Coin` has a denomination different than one defined by `StakingKeeper.BondDenom()`
- the mint rate is invalid, meaning the active validator has no tokens
- the amount delegated is less than the minimum allowed liquid staking `params.MinLiquidStakingAmount`

The delegator receives newly minted bTokens at the current mint rate. The mint rate is the total supply of bTokens divided by the number of native tokens.

```go
type MsgLiquidStake struct {
   DelegatorAddress string
   Amount           types.Coin
}
```

## MsgLiquidUnstake

This message allows delegators to unstake their liquid stake position

This message is expected to fail if:

- the delegator doesn't have bTokens
- the active validator doesn't exist
- the amount of bTokens is less than the minimum allowed unstaking
- the `Amount` has a denomination different than one defined by `params.LiquidBondDenom`

When this message is processed the following actions occur:

- calcuate the amount of native tokens for bTokens to unstake
- burn the bTokens
- `LiquidStakingProxyAcc` unbond the active validator's LiquidTokens by calculated native token worth of bTokens divided by number of active validators
    - the `DelegatorAddress` of the `UnbondingDelegation` would be `MsgLiquidStake.DelegatorAddress` not `LiquidStakingProxyAcc`

```go
type MsgLiquidUnstake struct {
   DelegatorAddress string
   Amount           types.Coin
}
```
