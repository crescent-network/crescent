<!-- order: 6 -->

# Parameters

The amm module contains the following parameters:

| Key                 | Type                  | Example                                                                                                                                                           |
|---------------------|-----------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Fees                | Fees                  | `{"market_creation_fee":[{"denom":"stake","amount":"1000000"}],"default_maker_fee_rate":"-0.001500000000000000","default_taker_fee_rate":"0.003000000000000000"}` |
| MaxOrderLifespan    | int64 (time.Duration) | 168h                                                                                                                                                              |
| MaxOrderPriceRatio  | sdk.Dec               | "0.100000000000000000"                                                                                                                                            |
| MaxSwapRoutesLen    | uint32                | 3                                                                                                                                                                 |
| MaxNumMMOrders      | uint32                | 15                                                                                                                                                                |
