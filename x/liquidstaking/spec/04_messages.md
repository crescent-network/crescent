<!-- order: 4 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps `liquidstaking` module messages from transactions.

## MsgLiquidStake

Liquid stake with an amount. A liquid staker is expected to receive a synthetic version of the native token `bToken` at the current mint rate.

```go
type MsgLiquidStake struct {
	DelegatorAddress string     // the bech32-encoded address of the delegator
	Amount           types.Coin // the amount of coin to liquid stake
}
```

### Validity Checks

Validity checks are performed for `MsgLiquidStake` message. The transaction that is triggered with `MsgLiquidStake` fails if:

- The active liquid validators do not exist
- The amount of coin denomination is different from the one defined in `StakingKeeper.BondDenom()`
- The mint rate is invalid. It means that the active liquid validator set has no tokens
- Insufficient spendable balances (locked coins are not allowed to liquid stake)
- The amount of coin is less than the minimum liquid liquid staking amount defined in `params.MinLiquidStakingAmount`

## MsgLiquidUnstake

Liquid unstake with an amount. A liquid staker is expected to receive native token that corresponds to the synthetic version of coin `bToken` value.

```go
type MsgLiquidUnstake struct {
	DelegatorAddress string     // the bech32-encoded address of the delegator
	Amount           types.Coin // the amount of coin to liquid unstake
}
```

### Validity Checks

Validity checks are performed for `MsgLiquidUnstake` message. The transaction that is triggered with `MsgLiquidUnstake` fails if:

- The active liquid validators do not exist 
- The amount of coin denomination is different from the one defined in `params.LiquidBondDenom`
- The liquid staker has insufficient amount of `bTokens`; `params.UnstakeFeeRate` must be considered
- Insufficient liquid tokens or balance in proxy account
