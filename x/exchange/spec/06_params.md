<!-- order: 6 -->

# Parameters

The amm module contains the following parameters:

| Key                     | Type                  | Example                                                                                                                                                      |
|-------------------------|-----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| MarketCreationFee       | sdk.Coins             | `[{"denom":"ucre","amount":"100000000"}]`                                                                                                                    |
| Fees                    | Fees                  | `{"default_maker_fee_rate":"-0.001500000000000000","default_taker_fee_rate":"0.003000000000000000","default_order_source_fee_ratio":"0.800000000000000000"}` |
| MaxOrderLifespan        | int64 (time.Duration) | 168h                                                                                                                                                         |
| MaxOrderPriceRatio      | sdk.Dec               | "0.100000000000000000"                                                                                                                                       |
| DefaultMinOrderQuantity | sdk.Int               | "1"                                                                                                                                                          |
| DefaultMinOrderQuote    | sdk.Int               | "1"                                                                                                                                                          |
| DefaultMaxOrderQuantity | sdk.Int               | "1000000000000000000000000000000"                                                                                                                            |
| DefaultMaxOrderQuote    | sdk.Int               | "1000000000000000000000000000000"                                                                                                                            |
| MaxSwapRoutesLen        | uint32                | 3                                                                                                                                                            |
| MaxNumMMOrders          | uint32                | 15                                                                                                                                                           |

## Fees

```go
type Fees struct {
    DefaultMakerFeeRate        sdk.Dec
    DefaultTakerFeeRate        sdk.Dec
    DefaultOrderSourceFeeRatio sdk.Dec
}
```
