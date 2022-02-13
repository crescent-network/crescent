<!-- order: 2 -->

# State

## LiquidValidator

LiquidValidator is the Validator that can be the target of LiquidStaking and LiquidUnstaking, Active, Weight, etc. fields are derived as functions to deal with by maintaining consistency with the state of the staking module.

```go
type LiquidValidator struct {
   // operator_address defines the address of the validator's operator; bech encoded in JSON.
   OperatorAddress string 
}
```

LiquidValidators: `0xc0 | OperatorAddrLen (1 byte) | OperatorAddr -> ProtocolBuffer(liquidValidator)`

LiquidValidators can have one of two statuses

- `Active` : When some validators in the active set are elected to whitelist by governance, liquid staker's delegating amount of native tokens are distributed to these vaidators. They can be slashed for misbehavior. They can be delisted. Liquid stakers who unbond their delegation must wait the duration of the UnStakingTime, a chain-specific param, during which time they are still slashable for offences of the active liquid validators if those offences were committed during the period of time that the tokens were bonded.
- `Inactive` : Not meet Active Condition, but has delegation shares of LiquidStakingProxyAcc, **Inactive liquid validator's TargetWeight Would be zero**

```go
const (
  ValidatorStatusUnspecified ValidatorStatus = 0
  ValidatorStatusActive ValidatorStatus = 1
  ValidatorStatusInactive ValidatorStatus = 2
)
```

## NetAmount
 NetAmount = `LiquidStakingProxyAcc's native token balance + total liquid tokens(slashing applied delegation shares) + remaining rewards + unbonding amount`
  - `MintRate = bTokenTotalSupply / NetAmount`
  - NativeTokenToBToken : `nativeTokenAmount * bTokenTotalSupply / netAmount` with truncations
  - BTokenToNativeToken : `bTokenAmount * netAmount / bTokenTotalSupply * (1-UnstakeFeeRate)` with truncations

## Liquid Staking

- Liquid stakers may delegate coins to active liquid validators; under this circumstance their funds are held in a `LiquidStaking` data structure. It is owned by one liquid staker, and is associated with the bTokens which represent their shares for active liquid validators.
- bTokens

  Liquid stakers receive bTokens in return for their liquid staking position. The amount of bTokens are minted is based on mint rate, calculated as follows from the total supply of bTokens and net amount of native tokens.
    - `MintAmount = StakeAmount * MintRate` by NativeTokenToBToken
    - when initial liquid staking, `MintAmount == StakeAmount`


## UnStaking

- Shares in the `LiquidStaking` can be unstaked, but they must for some time exist as an `Unstaking`, where shares can be reduced if misbehavior is detected. The amount of native tokens returned is calculated as follows
    - `UnstakeAmount = bTokenAmount / MintRate * (1 - UnstakeFeeRate)` by BTokenToNativeToken