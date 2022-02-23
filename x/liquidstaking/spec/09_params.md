<!-- order: 7 -->

# Parameters

The liquid-staking module contains the following parameters:

| Key                    | Type                   | Example                |
|------------------------|------------------------|------------------------|
| LiquidBondDenom        | string                 | “bstake”               |
| WhitelistedValidators  | []WhitelistedValidator |                        |
| UnstakeFeeRate         | string (sdk.Dec)       | "0.001000000000000000" |
| MinLiquidStakingAmount | string (sdk.Int)       | "1000000"              |

### LiquidBondDenom

Denomination of the token receiving after LiquidStaking, The value is calculated through NetAmount.

### WhitelistedValidators

Validators elected to become Active Liquid Validators consist of a list of WhitelistedValidator.

### WhitelistedValidator

// WhitelistedValidator consists of the validator operator address and the target weight, which is a value for calculating the real weight to be derived according to the active status. In the case of inactive, it is calculated as zero.

```go
type WhitelistedValidator struct {
   // validator_address defines the bech32-encoded address that whitelisted validator
   ValidatorAddress
   // target_weight specifies the weight for liquid staking, unstaking amount
   TargetWeight github_com_cosmos_cosmos_sdk_types.Int
}
```

### UnstakeFeeRate

When liquid unstake is requested, unbonded by subtracting the UnstakeFeeRate from unbondingAmount, which remains the DelShares of LiquidStakingProxyAcc, increasing the value of netAmount and bToken.
Even if the UnstakeFeeRate is zero, a small loss may occur due to a decimal error in the process of dividing the staking/unstaking amount into weight of liquid validators, which is also accumulated in the netAmount value like fee.

### MinLiquidStakingAmount

Define the minimum liquid staking amount to minimize decimal loss and consider gas efficiency.

## Constant Variables

| Key                | Type             | Constant Value         |
|--------------------|------------------|------------------------|
| RebalancingTrigger | string (sdk.Dec) | "0.001000000000000000" |
| RewardTrigger      | string (sdk.Dec) | "0.001000000000000000" |

## RebalancingTrigger

if the maximum difference and needed each redelegation amount exceeds `RebalancingTrigger`, asset rebalacing will be executed.

## RewardTrigger

If the sum of balance(the withdrawn rewards, crumb) and the upcoming rewards(all delegations rewards) of `LiquidStakingProxyAcc` exceeds `RewardTrigger` of the total `DelShares`, the reward is automatically withdrawn and re-stake according to the weights.

### LiquidStakingProxyAcc

Proxy Reserve account for Delegation and Undelegation.

```go
LiquidStakingProxyAcc = farmingtypes.DeriveAddress(farmingtypes.AddressType32Bytes, ModuleName, "LiquidStakingProxyAcc")
```
