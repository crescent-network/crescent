<!-- order: 6 -->

# Events

The `farming` module emits the following events:

## EndBlocker

| Type              | Attribute Key        | Attribute Value        |
|-------------------|----------------------|------------------------|
| plan_terminated   | plan_id              | {planID}               |
| plan_terminated   | farming_pool_address | {farmingPoolAddress}   |
| plan_terminated   | termination_address  | {terminationAddress}   |
| rewards_allocated | plan_id              | {planID}               |
| rewards_allocated | amount               | {totalAllocatedAmount} |
| rewards_withdrawn | farmer               | {farmer}               |
| rewards_withdrawn | staking_coin_denom   | {stakingCoinDenom}     |
| rewards_withdrawn | rewards_coins        | {rewardCoins}          |

## Handlers

### MsgCreateFixedAmountPlan

| Type                     | Attribute Key        | Attribute Value          |
|--------------------------|----------------------|--------------------------|
| create_fixed_amount_plan | plan_id              | {planID}                 |
| create_fixed_amount_plan | plan_name            | {planName}               |
| create_fixed_amount_plan | farming_pool_address | {farmingPoolAddress}     |
| create_fixed_amount_plan | start_time           | {startTime}              |
| create_fixed_amount_plan | end_time             | {endTime}                |
| create_fixed_amount_plan | epoch_amount         | {epochAmount}            |
| message                  | module               | farming                  |
| message                  | action               | create_fixed_amount_plan |
| message                  | sender               | {senderAddress}          |

### MsgCreateRatioPlan

| Type              | Attribute Key        | Attribute Value      |
|-------------------|----------------------|----------------------|
| create_ratio_plan | plan_id              | {planID}             |
| create_ratio_plan | plan_name            | {planName}           |
| create_ratio_plan | farming_pool_address | {farmingPoolAddress} |
| create_ratio_plan | start_time           | {startTime}          |
| create_ratio_plan | end_time             | {endTime}            |
| create_ratio_plan | epoch_ratio          | {epochRatio}         |
| message           | module               | farming              |
| message           | action               | create_ratio_plan    |
| message           | sender               | {senderAddress}      |

### MsgStake

| Type    | Attribute Key | Attribute Value |
|---------|---------------|-----------------|
| stake   | farmer        | {farmer}        |
| stake   | staking_coins | {stakingCoins}  | 
| message | module        | farming         |
| message | action        | stake           |
| message | sender        | {senderAddress} |

### MsgUnstake

| Type              | Attribute Key      | Attribute Value    |
|-------------------|--------------------|--------------------|
| unstake           | farmer             | {farmer}           |
| unstake           | unstaking_coins    | {unstakingCoins}   |
| rewards_withdrawn | farmer             | {farmer}           |
| rewards_withdrawn | staking_coin_denom | {stakingCoinDenom} |
| rewards_withdrawn | rewards_coins      | {rewardCoins}      |
| message           | module             | farming            |
| message           | action             | unstake            |
| message           | sender             | {senderAddress}    |

### MsgHarvest

| Type    | Attribute Key       | Attribute Value     |
|---------|---------------------|---------------------|
| harvest | farmer              | {farmer}            |
| harvest | staking_coin_denoms | {stakingCoinDenoms} |
| message | module              | farming             |
| message | action              | harvest             |
| message | sender              | {senderAddress}     |

### MsgRemovePlan

| Type        | Attribute Key | Attribute Value |
|-------------|---------------|-----------------|
| remove_plan | creator       | {creator}       |
| remove_plan | plan_id       | {planId}        |
| message     | module        | farming         |
| message     | action        | remove_plan     |
| message     | sender        | {senderAddress} |

### MsgAdvanceEpoch

The `MsgAdvanceEpoch` message is for testing purposes only and requires that you build the `crescentd` binary. See [MsgAdvanceEpoch](04_messages.md#MsgAdvanceEpoch).
