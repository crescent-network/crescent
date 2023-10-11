<!-- order: 6 -->

# Parameters

The amm module contains the following parameters:

| Key                           | Type                  | Example                               |
|-------------------------------|-----------------------|---------------------------------------|
| PoolCreationFee               | array (sdk.Coins)     | [{"denom":"ucre","amount":"1000000"}] |
| DefaultTickSpacing            | uint32                | 25                                    |
| PrivateFarmingPlanCreationFee | array (sdk.Coins)     | [{"denom":"ucre","amount":"1000000"}] |
| MaxNumPrivateFarmingPlans     | uint32                | 50                                    |
| MaxFarmingBlockTime           | int64 (time.Duration) | 10s                                   |
