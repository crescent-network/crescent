<!-- order: 7 -->

# Events

## BeginBlocker

| Type                                | Attribute Key           | Attribute Value                |
|-------------------------------------|-------------------------|--------------------------------|
| add_liquid_validator                | liquid_validator        | {liquidValidatorAddress}       |
| remove_liquid_validator             | liquid_validator        | {liquidValidatorAddress}       |
| begin_rebalancing                   | delegator               | {liquidStakingProxyAccAddress} |
| begin_rebalancing                   | redelegation_count      | {NeededRedelegationCount}      |
| begin_rebalancing                   | redelegation_fail_count | {RedelegationFailCount}        |
| EventTypeReStake                    | delegator               | {liquidStakingProxyAccAddress} |
| EventTypeReStake                    | amount                  | {liquidStakingProxyAccBalance} |
| EventTypeUnbondInactiveLiquidTokens | liquid_validator        | {liquidValidatorAddress}       |
| EventTypeUnbondInactiveLiquidTokens | unbonding_amount        | {unbondAmount}                 |
| EventTypeUnbondInactiveLiquidTokens | completion_time         | {completionTime}               |


## Handlers

### MsgLiquidStake

| Type         | Attribute Key        | Attribute Value    |
|--------------|----------------------|--------------------|
| liquid_stake | delegator            | {delegatorAddress} |
| liquid_stake | amount               | {delegationAmount} |
| liquid_stake | btoken_minted_amount | {newDelShares}     |
| liquid_stake | amount               | {bTokenMintAmount} |
| message      | module               | liquidstaking      |
| message      | action               | liquid_stake       |
| message      | sender               | {senderAddress}    |

### MsgLiquidUnstake

| Type           | Attribute Key    | Attribute Value    |
|----------------|------------------|--------------------|
| liquid_unstake | validator        | {validatorAddress} |
| liquid_unstake | amount           | {bTokenBurnAmount} |
| liquid_unstake | unbonding_amount | {unbondingAmount}  |
| liquid_unstake | unbonded_amount  | {unbondedAmount}   |
| liquid_unstake | completion_time  | {completionTime}   |
| message        | module           | liquidstaking      |
| message        | action           | liquid_unstake     |
| message        | sender           | {senderAddress}    |
