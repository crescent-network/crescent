<!-- order: 8 -->

# Parameters

The `liquidstaking` module contains the following parameters:

| Key                    | Type                   | Example                |
|------------------------|------------------------|------------------------|
| LiquidBondDenom        | string                 | “bstake”               |
| WhitelistedValidators  | []WhitelistedValidator |                        |
| UnstakeFeeRate         | string (sdk.Dec)       | "0.001000000000000000" |
| MinLiquidStakingAmount | string (sdk.Int)       | "1000000"              |

## LiquidBondDenom

The denomination of the token that liquid stakers receive after they liquid stake. It acts as staking representation. 

## WhitelistedValidators

It is a list of `WhitelistedValidator`. A list of whitelisted validator is defined in `params.WhitelistedValidators` and they are being governed and elected through governance process. `WhitelistedValidator` has validator operator address and target weight. A target weight is a value used for calculating the real weight considering the active status. It is calculated to zero when a liquid validator's status is inactive.

```go
type WhitelistedValidator struct {
   // validator_address defines the bech32-encoded address that whitelisted validator
   ValidatorAddress
   // target_weight defines the weight for liquid staking and unstaking amount
   TargetWeight github_com_cosmos_cosmos_sdk_types.Int
}
```

## UnstakeFeeRate

It is the fee rate that liquid stakers pay when they liquid unstake. When liquid unstake is requested, unbonded by subtracting the UnstakeFeeRate from unbondingAmount, which remains the DelShares of LiquidStakingProxyAcc, increasing the value of netAmount and bToken. Even if the `UnstakeFeeRate` is zero, a small loss may occur due to a decimal loss in the process of dividing the staking/unstaking amount into weight of liquid validators, which is also accumulated in the netAmount value like fee.

## MinLiquidStakingAmount

It is the minimum liquid staking amount. It is used for minimizing decimal loss during calculation and gas efficiency.

## Constant Variables

| Key                | Type             | Constant Value         |
|--------------------|------------------|------------------------|
| RebalancingTrigger | string (sdk.Dec) | "0.001000000000000000" |
| RewardTrigger      | string (sdk.Dec) | "0.001000000000000000" |

## RebalancingTrigger

It is the maximum difference and required rate that triggers asset rebalancing (redelegation) for all liquid validators.

## RewardTrigger

It is the rate that triggers to withdraw rewards and re-stake amounts to active validators. Specifically, if the sum of balances including the withdrawn rewards, crumb, and the upcoming rewards of `LiquidStakingProxyAcc` exceeds the rate of `RewardTrigger` of the total `DelShares`, the rewards are automatically withdrawn and re-stake according to each validator's weight.

### LiquidStakingProxyAcc

The proxy reserve account for all delegations and undelegations. It is derived by the following code snippet.

```go
LiquidStakingProxyAcc = farmingtypes.DeriveAddress(farmingtypes.AddressType32Bytes, ModuleName, "LiquidStakingProxyAcc")
```
