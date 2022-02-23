<!-- order: 4 -->

# Messages

## MsgLiquidStake

Within this message the delegator provides coins, and in return receives newly minted bTokens at the current mint rate.

This message is expected to fail if:

- the active liquid validator does not exist
- the `Amount` `Coin` has a denomination different than one defined by `StakingKeeper.BondDenom()`
- the mint rate is invalid, meaning the active liquid validator has no tokens
- insufficient spendable balances (not allowed locked coins)
- the amount delegated is less than the minimum allowed liquid staking `params.MinLiquidStakingAmount`

```go
type MsgLiquidStake struct {
   DelegatorAddress string
   Amount           types.Coin
}
```

## MsgLiquidUnstake

This message allows delegators to unstake their liquid stake position by burn the requested bToken amount and begins unbonding the corresponding value.

This message is expected to fail if:

- the `Amount` has a denomination different than one defined by `params.LiquidBondDenom`
- the delegator doesn't have sufficient bTokens `Amount`
- liquid validators to unbond doesn't exist with insufficient liquid tokens or balance of proxy account


```go
type MsgLiquidUnstake struct {
   DelegatorAddress string
   Amount           types.Coin
}
```
