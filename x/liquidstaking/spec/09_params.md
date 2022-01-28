<!-- order: 7 -->

# Parameters

The liquid-staking module contains the following parameters:

| Key                    | Type                   | Example                |
|------------------------| ---------------------- | ---------------------- |
| BondedBondDenom        | string                 | “bstake”               |
| WhitelistedValidators  | []WhitelistedValidator |                        |
| UnstakeFeeRate         | string (sdk.Dec)       | "0.001000000000000000" |
| MinLiquidStakingAmount | string (sdk.Int)       | "1000000"              |

### WhitelistedValidator

```go
type WhitelistedValidator struct {
   // validator_address defines the bech32-encoded address that whitelisted validator
   ValidatorAddress
   // target_weight specifies the weight for liquid staking, unstaking amount
   TargetWeight github_com_cosmos_cosmos_sdk_types.Int
}
```

### UnstakeFeeRate

### MinLiquidStakingAmount

## Constant Variables

| Key                | Type             | Constant Value         |
| ------------------ | ---------------- | ---------------------- |
| RebalancingTrigger | string (sdk.Dec) | "0.001000000000000000" |
| RewardTrigger      | string (sdk.Dec) | "0.001000000000000000" |

## RebalancingTrigger

## RewardTrigger

### LiquidStakingProxyAcc

```go
LiquidStakingProxyAcc = farmingtypes.DeriveAddress(farmingtypes.AddressType32Bytes, ModuleName, "LiquidStakingProxyAcc")
```
