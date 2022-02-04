<!-- order: 2 -->

# State

## LiquidValidator

```go
type LiquidValidator struct {
   // operator_address defines the address of the validator's operator; bech encoded in JSON.
   OperatorAddress string 
}
```

LiquidValidators: `0xc0 | OperatorAddrLen (1 byte) | OperatorAddr -> ProtocolBuffer(liquidValidator)`

LiquidValidators can have one of two statuses

- `Active` : When some validators in the active set are elected to whiltelist by governance, liquid staker's delegating amount of native tokens are distributed to these vaidators. They can be slashed for misbehavior. They can be delisted. Liquid stakers who unbond their delegation must wait the duration of the UnStakingTime, a chain-specific param, during which time they are still slashable for offences of the active validators if those offences were committed during the period of time that the tokens were bonded.
- `Unspecified` : not existed or invalid validator on staking module

```go
const (
  ValidatorStatusUnspecified ValidatorStatus = 0
  ValidatorStatusActive ValidatorStatus = 1
)
```

## Liquid Staking

- Liquid stakers may delegate coins to active validators; under this circumstance their funds are held in a `LiquidStaking` data structure. It is owned by one liquid staker, and is associated with the bTokens which represent their shares for active validators.
- bTokens

  Liquid stakers receive bTokens in return for their liquid staking position. The amount of bTokens are minted is based on mint rate, calculated as follows from the total supply of bTokens and net amount of native tokens.

    - `MintRate = TotalSupply / NetAmount`
    - `MintAmount = MintRate * StakeAmount`

  MintRate = 1 for the special case of initial liquid staking


## UnStaking

- Shares in the `LiquidStaking` can be unstaked, but they must for some time exist as an `Unstaking`, where shares can be reduced if misbehavior is detected. The amount of native tokens returned is calculated as follows
    - `UnstakeAmount = bTokenAmount / MintRate * (1 - UnstakeFeeRate)`