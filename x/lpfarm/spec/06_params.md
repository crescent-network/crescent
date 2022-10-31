<!-- order: 6 -->

# Parameters

The lpfarm module contains the following parameters:

| Key                    | Type                  | Example                                |
|------------------------|-----------------------|----------------------------------------|
| PrivatePlanCreationFee | array (sdk.Coins)     | [{"denom":"stake","amount":"1000000"}] |
| FeeCollector           | string                | "cosmos1..."                           |
| MaxNumPrivatePlans     | uint32                | 50                                     |
| MaxBlockDuration       | int64 (time.Duration) | 10s                                    |
