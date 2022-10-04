<!-- order: 5 -->

# Events

## Handlers

### MsgCreatePrivatePlan

| Type                                         | Attribute Key        | Attribute Value                              |
|----------------------------------------------|----------------------|----------------------------------------------|
| message                                      | action               | /crescent.farm.v1beta1.Msg/CreatePrivatePlan |
| crescent.farm.v1beta1.EventCreatePrivatePlan | creator              | {planCreatorAddress}                         |
| crescent.farm.v1beta1.EventCreatePrivatePlan | plan_id              | {planId}                                     |
| crescent.farm.v1beta1.EventCreatePrivatePlan | farming_pool_address | {farmingPoolAddress}                         |

### MsgFarm

| Type                            | Attribute Key     | Attribute Value                 |
|---------------------------------|-------------------|---------------------------------|
| message                         | action            | /crescent.farm.v1beta1.Msg/Farm |
| crescent.farm.v1beta1.EventFarm | farmer            | {farmerAddress}                 |
| crescent.farm.v1beta1.EventFarm | coin              | {coin}                          |
| crescent.farm.v1beta1.EventFarm | withdrawn_rewards | {withdrawnRewards}              |

### MsgUnfarm

| Type                              | Attribute Key     | Attribute Value                   |
|-----------------------------------|-------------------|-----------------------------------|
| message                           | action            | /crescent.farm.v1beta1.Msg/Unfarm |
| crescent.farm.v1beta1.EventUnfarm | farmer            | {farmerAddress}                   |
| crescent.farm.v1beta1.EventUnfarm | coin              | {coin}                            |
| crescent.farm.v1beta1.EventUnfarm | withdrawn_rewards | {withdrawnRewards}                |

### MsgHarvest

| Type                               | Attribute Key     | Attribute Value                    |
|------------------------------------|-------------------|------------------------------------|
| message                            | action            | /crescent.farm.v1beta1.Msg/Harvest |
| crescent.farm.v1beta1.EventHarvest | farmer            | {farmerAddress}                    |
| crescent.farm.v1beta1.EventHarvest | denom             | {farmingAssetDenom}                |
| crescent.farm.v1beta1.EventHarvest | withdrawn_rewards | {withdrawnRewards}                 |
