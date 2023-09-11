---
Title: AMM
Description: A high-level overview of what gRPC-gateway REST routes are supported in the amm module.
---

# AMM Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the `amm` module.

## gRPC-gateway REST Routes

<!-- markdown-link-check-disable -->
++https://github.com/crescent-network/crescent/blob/main/proto/crescent/amm/v1beta1/query.proto (need check) 

- [Params](#Params)
- [AllPools](#AllPools)
- [Pool](#Pool)
- [AllPositions](#AllPositions)
- [Position](#Position)
- [AddLiquiditySimulation](#AddLiquiditySimulation)
- [RemoveLiquiditySimulation](#RemoveLiquiditySimulation)
- [CollectibleCoins](#CollectibleCoins)
- [AllTickInfos](#AllTickInfos)
- [TickInfo](#TickInfo)
- [AllFarmingPlans](#AllFarmingPlans)
- [FarmingPlan](#FarmingPlan)

## Params

Example Request

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/params
```

Example Response

```json
{
  "params": {
    "pool_creation_fee": [
      {
        "denom": "stake",
        "amount": "1000000"
      }
    ],
    "default_tick_spacing": 50,
    "private_farming_plan_creation_fee": [
      {
        "denom": "stake",
        "amount": "1000000"
      }
    ],
    "max_num_private_farming_plans": 50,
    "max_farming_block_time": "10s"
  }
}
```

## AllPools

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/pools
http://localhost:1317/crescent/amm/v1beta1/pools?market_id=1
```

Example Response

```json
{
  "pools": [
    {
      "id": "1",
      "market_id": "1",
      "balance0": {
        "denom": "uatom",
        "amount": "706867"
      },
      "balance1": {
        "denom": "uusd",
        "amount": "12020740"
      },
      "reserve_address": "cre1pf3w839s80c26z8dqccuzu4epyj58jrzglqv93yppqpj6yggu9qs8vpazr",
      "rewards_pool": "cre1srphgsfqllr85ndknjme24txux8m0sz0hhpnnksn2339d3a788rs3ax6tu",
      "tick_spacing": 50,
      "min_order_quantity": "1.000000000000000000",
      "min_order_quote": "1.000000000000000000",
      "current_tick": 90208,
      "current_price": "10.208470700277704312",
      "current_liquidity": "61622776",
      "total_liquidity": "61622776",
      "fee_growth_global": [
        {
          "denom": "uatom",
          "amount": "2417937.160117551341731180"
        },
        {
          "denom": "uusd",
          "amount": "8072226411.870831654841385269"
        }
      ],
      "farming_rewards_growth_global": [
        {
          "denom": "uatom",
          "amount": "1363132.358723988675875290"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Pool

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/pools/1
```

Example Response

```json
{
  "pool": {
    "id": "1",
    "market_id": "1",
    "balance0": {
      "denom": "uatom",
      "amount": "706867"
    },
    "balance1": {
      "denom": "uusd",
      "amount": "12020740"
    },
    "reserve_address": "cre1pf3w839s80c26z8dqccuzu4epyj58jrzglqv93yppqpj6yggu9qs8vpazr",
    "rewards_pool": "cre1srphgsfqllr85ndknjme24txux8m0sz0hhpnnksn2339d3a788rs3ax6tu",
    "tick_spacing": 50,
    "min_order_quantity": "1.000000000000000000",
    "min_order_quote": "1.000000000000000000",
    "current_tick": 90208,
    "current_price": "10.208470700277704312",
    "current_liquidity": "61622776",
    "total_liquidity": "61622776",
    "fee_growth_global": [
      {
        "denom": "uatom",
        "amount": "2417937.160117551341731180"
      },
      {
        "denom": "uusd",
        "amount": "8072226411.870831654841385269"
      }
    ],
    "farming_rewards_growth_global": [
      {
        "denom": "uatom",
        "amount": "6052956.783381521144065276"
      }
    ]
  }
}
```

## AllPositions

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/positions
http://localhost:1317/crescent/amm/v1beta1/positions?pool_id=1
http://localhost:1317/crescent/amm/v1beta1/positions?owner=cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p
```

Example Response

```json
{
  "positions": [
    {
      "id": "1",
      "pool_id": "1",
      "owner": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "lower_price": "9.000000000000000000",
      "upper_price": "11.000000000000000000",
      "liquidity": "61622776",
      "last_fee_growth_inside": [
      ],
      "owed_fee": [
      ],
      "last_farming_rewards_growth_inside": [
      ],
      "owed_farming_rewards": [
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Position

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/positions/1
```

Example Response

```json
{
  "position": {
    "id": "1",
    "pool_id": "1",
    "owner": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
    "lower_price": "9.000000000000000000",
    "upper_price": "11.000000000000000000",
    "liquidity": "61622776",
    "last_fee_growth_inside": [
    ],
    "owed_fee": [
    ],
    "last_farming_rewards_growth_inside": [
    ],
    "owed_farming_rewards": [
    ]
  }
}
```

## AddLiquiditySimulation

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/simulation/add_liquidity?pool_id=1&lower_price=9&upper_price=11&desired_amount=100uatom,100uusd
```

Example Response

```json
{
  "liquidity": "559",
  "amount": [
    {
      "denom": "uatom",
      "amount": "8"
    },
    {
      "denom": "uusd",
      "amount": "100"
    }
  ]
}
```

## RemoveLiquiditySimulation

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/simulation/remove_liquidity?position_id=1&liquidity=10000
```

Example Response

```json
{
  "amount": [
    {
      "denom": "uatom",
      "amount": "130"
    },
    {
      "denom": "uusd",
      "amount": "1785"
    }
  ]
}
```

## CollectibleCoins

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/collectible_coins?owner=cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p
http://localhost:1317/crescent/amm/v1beta1/collectible_coins?position_id=1
```

Example Response

```json
{
  "fee": [
    {
      "denom": "uatom",
      "amount": "147"
    },
    {
      "denom": "uusd",
      "amount": "2571"
    }
  ],
  "farming_rewards": [
    {
      "denom": "uatom",
      "amount": "63407"
    }
  ]
}
```

## AllTickInfos

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/pools/1/tick_infos
```

Example Response

```json
{
  "tick_infos": [
    {
      "tick": 80000,
      "gross_liquidity": "61622776",
      "net_liquidity": "61622776",
      "fee_growth_outside": [
      ],
      "farming_rewards_growth_outside": [
      ]
    },
    {
      "tick": 91000,
      "gross_liquidity": "61622776",
      "net_liquidity": "-61622776",
      "fee_growth_outside": [
      ],
      "farming_rewards_growth_outside": [
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

## TickInfo

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/pools/1/tick_infos/80000
```

Example Response

```json
{
  "tick_info": {
    "tick": 80000,
    "gross_liquidity": "61622776",
    "net_liquidity": "61622776",
    "fee_growth_outside": [
    ],
    "farming_rewards_growth_outside": [
    ]
  }
}
```

## AllFarmingPlans

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/farming_plans
```

Example Response

```json
{
  "farming_plans": [
    {
      "id": "1",
      "description": "New farming plan",
      "farming_pool_address": "cre1ll5dtdmug9n54fnhr9fpr8nmr840s72dstydd723ufgsxzrjg5qq9kzn3e",
      "termination_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "reward_allocations": [
        {
          "pool_id": "1",
          "rewards_per_day": [
            {
              "denom": "uatom",
              "amount": "1000000"
            }
          ]
        }
      ],
      "start_time": "2023-01-01T00:00:00Z",
      "end_time": "2024-01-01T00:00:00Z",
      "is_private": true,
      "is_terminated": false
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## FarmingPlan

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/amm/v1beta1/farming_plans/1
```

Example Response

```json
{
  "farming_plan": {
    "id": "1",
    "description": "New farming plan",
    "farming_pool_address": "cre1ll5dtdmug9n54fnhr9fpr8nmr840s72dstydd723ufgsxzrjg5qq9kzn3e",
    "termination_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
    "reward_allocations": [
      {
        "pool_id": "1",
        "rewards_per_day": [
          {
            "denom": "uatom",
            "amount": "1000000"
          }
        ]
      }
    ],
    "start_time": "2023-01-01T00:00:00Z",
    "end_time": "2024-01-01T00:00:00Z",
    "is_private": true,
    "is_terminated": false
  }
}
```
