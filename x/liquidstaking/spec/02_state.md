<!-- order: 2 -->

# State

## LiquidValidator

LiquidValidator is the validator that can be the target of liquid staking and liquid unstaking.
Active Status, Weight, LiquidTokens, DelShares, etc. fields are derived as functions to deal with by maintaining consistency with the state of the staking module.
LiquidValidators are generated when the validator included in params.WhitelistedValidators meets the active conditions, and then if it becomes inactive status and no delShares involved liquid staking due to rebalancing and unbonding, removes the LiquidValidator.

```go
type LiquidValidator struct {
   // operator_address defines the address of the validator's operator; bech encoded in JSON.
   OperatorAddress string 
}
```

```go
// LiquidValidatorState is type LiquidValidator with state added to return to query results.
type LiquidValidatorState struct {
	// operator_address defines the address of the validator's operator; bech encoded in JSON.
	OperatorAddress string
	// weight specifies the weight for liquid staking, unstaking amount
	Weight sdk.Int
	// status is the liquid validator status
	Status ValidatorStatus
	// del_shares define the delegation shares of the validator
	DelShares sdk.Dec
	// liquid_tokens define the token amount worth of delegation shares of the validator (slashing applied amount)
	LiquidTokens sdk.Int
}
```

LiquidValidators: `0xc0 | OperatorAddrLen (1 byte) | OperatorAddr -> ProtocolBuffer(liquidValidator)`

### Status
LiquidValidators can have one of two statuses

- `Active` : When some validators in the active set are elected to whitelist by governance, delegator's delegating amount of native tokens are distributed to these vaidators. They can be slashed for misbehavior. They can be delisted. Liquid stakers who unbond their delegation must wait the duration of the UnStakingTime, a chain-specific param, during which time they are still slashable for offences of the active liquid validators if those offences were committed during the period of time that the tokens were bonded.
- `Inactive` : Not meet Active Condition, but has delegation shares of LiquidStakingProxyAcc, **Inactive liquid validator's Weight would be derived zero**

```go
const (
  ValidatorStatusUnspecified ValidatorStatus = 0
  ValidatorStatusActive ValidatorStatus = 1
  ValidatorStatusInactive ValidatorStatus = 2
)
```

### Active Conditions
- included on whitelist
- existed valid validator on staking module ( existed, not nil DelShares and tokens, valid exchange rate)
- not tombstoned


### Weight

Weight of LiquidValidator is derived as follows depending on the Status.

- Active LiquidValidator : `TargetWeight` value defined in `params.WhitelistedValidators` by governance
- Inactive LiquidValidator : zero (`0`)

## NetAmount

NetAmount is the sum of the items below of `LiquidStakingProxyAcc`. 
- token amount worth of delegation shares of all liquid validators
- remaining rewards 
- native token balance
- unbonding balance

MintRate is the total supply of bTokens divided by NetAmount `bTokenTotalSupply / NetAmount` depending on the equation, the value transformation between native tokens and bTokens can be calculated as follows.
- NativeTokenToBToken : `nativeTokenAmount * bTokenTotalSupply / netAmount` with truncations
- BTokenToNativeToken : `bTokenAmount * netAmount / bTokenTotalSupply * (1-params.UnstakeFeeRate)` with truncations


### NetAmountState

NetAmountState is type for net amount raw data and mint rate, This is a value that depends on the several module state every time, so it is used only for calculation and query and is not stored in kv.

```go
type NetAmountState struct {
	// mint_rate is bTokenTotalSupply / NetAmount
	MintRate sdk.Dec
	// btoken_total_supply returns the total supply of btoken(liquid_bond_denom)
	BtokenTotalSupply sdk.Int
	// net_amount is proxy account's native token balance + total liquid tokens + total remaining rewards + total unbonding balance
	NetAmount sdk.Dec
	// total_del_shares define the delegation shares of all liquid validators
	TotalDelShares sdk.Dec
	// total_liquid_tokens define the token amount worth of delegation shares of all liquid validator (slashing applied amount)
	TotalLiquidTokens sdk.Int
	// total_remaining_rewards define the sum of remaining rewards of proxy account by all liquid validators
	TotalRemainingRewards sdk.Dec
	// total_unbonding_balance define the unbonding balance of proxy account by all liquid validator (slashing applied amount)
	TotalUnbondingBalance sdk.Int
	// proxy_acc_balance define the balance of proxy account for the native token
	ProxyAccBalance sdk.Int
}
```
