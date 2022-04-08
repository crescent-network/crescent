<!-- order: 2 -->

# State

## LiquidValidator

LiquidValidator is a validator for liquid staking. Liquid validators are set from the whitelisted validators that are defined in global parameter `params.WhitelistedValidators`. Whitelisted validators must meet the active conditions (see below section). Otherwise they become inactive status; this results to no delegation shares and being removed from the active liquid validator set. This occurs during rebalancing at every begin block. 

```go
// LiquidValidator is a validator for liquid staking
type LiquidValidator struct {
   // operator_address defines the bech32-encoded address of the validator operator
   OperatorAddress string 
}
```

LiquidValidatorState contains the validator's state of status, weight, delegation shares, and liquid tokens. Each field has derived function that syncs with the state of the `staking` module. This object is not stored in KVStore and only used for querying state of a liquid validator.

```go
// LiquidValidatorState is a liquid validator state
type LiquidValidatorState struct {
	// operator_address defines the bech32-encoded address of the validator operator
	OperatorAddress string
	// weight defines the weight that corresponds to liquid staking and unstaking amount
	Weight sdk.Int
	// status defines the liquid validator status
	Status ValidatorStatus
	// del_shares defines the delegation shares of the liquid validator
	DelShares sdk.Dec
	// liquid_tokens defines the token amount worth of delegaiton shares (slashing applied amount)
	LiquidTokens sdk.Int
}
```

LiquidValidators: `0xc0 | OperatorAddrLen (1 byte) | OperatorAddr -> ProtocolBuffer(LiquidValidator)`

### Status

A liquid validator has the following status:

- `Active`: active validators are the whitelisted validators who are governed and elected by governance process. Delegators' delegations are distributed to all liquid validators that correspond to their weight.

- `Inactive`: inactive validators are the ones that do not meet active conditions (see below section)

```go
const (
	// VALIDATOR_STATUS_UNSPECIFIED defines the unspecified invalid status
	ValidatorStatusUnspecified ValidatorStatus = 0
	// VALIDATOR_STATUS_ACTIVE defines the active, valid status
	ValidatorStatusActive ValidatorStatus = 1
	// VALIDATOR_STATUS_INACTIVE defines the inactive, invalid status
	ValidatorStatusInactive ValidatorStatus = 2
)
```

### Active Conditions

- Must exist in `params.WhitelistedValidators`
- Must be a validator in `staking` module
- Must not be tombstoned

### Weight

The weight of a liquid validator is derived depending on their status:

- Active LiquidValidator: `TargetWeight` value defined in `params.WhitelistedValidators` by governance

- Inactive LiquidValidator: zero (`0`)

## NetAmount

NetAmount is the sum of the following items that belongs to `LiquidStakingProxyAcc`:

- Native token balance
- Token amount worth of delegation shares from all liquid validators
- Remaining rewards 
- Unbonding balance

`MintRate` is the rate that is calculated from total supply of `bTokens` divided by `NetAmount`. 
- `MintRate = bTokenTotalSupply / NetAmount` 

Depending on the equation, the value transformation between native tokens and bTokens can be calculated as follows:

- NativeTokenToBToken : `nativeTokenAmount * bTokenTotalSupply / netAmount` with truncations
- BTokenToNativeToken : `bTokenAmount * netAmount / bTokenTotalSupply * (1-params.UnstakeFeeRate)` with truncations


### NetAmountState

NetAmountState provides states with each field for `NetAmount`. Each field is derived from several modules. This object is not stored in KVStore and only used for querying `NetAmount` state.

```go
// NetAmountState is type for NetAmount
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
