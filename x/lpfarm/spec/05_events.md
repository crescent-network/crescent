<!-- order: 5 -->

# Events

## Handlers

### MsgCreatePrivatePlan

| Type                                           | Attribute Key        | Attribute Value                                |
|------------------------------------------------|----------------------|------------------------------------------------|
| message                                        | action               | /crescent.lpfarm.v1beta1.Msg/CreatePrivatePlan |
| crescent.lpfarm.v1beta1.EventCreatePrivatePlan | creator              | {planCreatorAddress}                           |
| crescent.lpfarm.v1beta1.EventCreatePrivatePlan | plan_id              | {planId}                                       |
| crescent.lpfarm.v1beta1.EventCreatePrivatePlan | farming_pool_address | {farmingPoolAddress}                           |

### MsgFarm

| Type                              | Attribute Key     | Attribute Value                   |
|-----------------------------------|-------------------|-----------------------------------|
| message                           | action            | /crescent.lpfarm.v1beta1.Msg/Farm |
| crescent.lpfarm.v1beta1.EventFarm | farmer            | {farmerAddress}                   |
| crescent.lpfarm.v1beta1.EventFarm | coin              | {coin}                            |
| crescent.lpfarm.v1beta1.EventFarm | withdrawn_rewards | {withdrawnRewards}                |

### MsgUnfarm

| Type                                | Attribute Key     | Attribute Value                     |
|-------------------------------------|-------------------|-------------------------------------|
| message                             | action            | /crescent.lpfarm.v1beta1.Msg/Unfarm |
| crescent.lpfarm.v1beta1.EventUnfarm | farmer            | {farmerAddress}                     |
| crescent.lpfarm.v1beta1.EventUnfarm | coin              | {coin}                              |
| crescent.lpfarm.v1beta1.EventUnfarm | withdrawn_rewards | {withdrawnRewards}                  |

### MsgHarvest

| Type                                 | Attribute Key     | Attribute Value                      |
|--------------------------------------|-------------------|--------------------------------------|
| message                              | action            | /crescent.lpfarm.v1beta1.Msg/Harvest |
| crescent.lpfarm.v1beta1.EventHarvest | farmer            | {farmerAddress}                      |
| crescent.lpfarm.v1beta1.EventHarvest | denom             | {farmingAssetDenom}                  |
| crescent.lpfarm.v1beta1.EventHarvest | withdrawn_rewards | {withdrawnRewards}                   |
