<!-- order: 7 -->

# Parameters

The `liquidfarming` module contains the following parameters:

| Key                    | Type             | Example                                                        |
| ---------------------- | ---------------- | -------------------------------------------------------------- |
| LiquidFarms            | []LiquidFarm     | []LiquidFarm{}                                                 |
| RewardsAuctionDuration | string (time ns) | 43200000000000 (12 hours)                                      |
| FeeCollector           | string           | cre1lsvtflq2gau8ha7zvlethfy85qus59eserphyhc3tumua7upx6eqckll2q |

## LiquidFarms

`LiquidFarms` is a list of `LiquidFarm`, where a `LiquidFarm` is corresponding to a specific pool with `PoolId`.

```go
type LiquidFarm struct {
	PoolId        uint64  // the pool id
	MinFarmAmount sdk.Int // the minimum farm amount; it allows zero value
	MinBidAmount  sdk.Int // the minimum bid amount; it allows zero value
	FeeRate       sdk.Dec // the fee rate that deducts from auction winner's rewards; default value is 0
}
```

## RewardsAuctionDuration

`RewardsAuctionDuration` is the duration that triggers the module to create new `RewardsAuction`.
If there is an ongoing `RewardsAuction`, then it finishes it and it creates next one.
