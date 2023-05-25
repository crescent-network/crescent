<!-- order: 6 -->

# Parameters

The amm module contains the following parameters:

| Key                 | Type                  | Example                                |
|---------------------|-----------------------|----------------------------------------|
| MarketCreationFee   | array (sdk.Coins)     | [{"denom":"stake","amount":"1000000"}] |
| DefaultMakerFeeRate | sdk.Dec               | "-0.001500000000000000"                |
| DefaultTakerFeeRate | sdk.Dec               | "0.003000000000000000"                 |
| MaxOrderLifespan    | int64 (time.Duration) | 168h                                   |
| MaxOrderPriceRatio  | sdk.Dec               | "0.100000000000000000"                 |
| MaxSwapRoutesLen    | uint32                | 3                                      |
