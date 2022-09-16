<!-- order: 7 -->

# Parameters

The `liquidfarming` module contains the following parameters:

| Key                        | Type         | Example                                        |
| -------------------------- | ------------ | ---------------------------------------------- |
| LiquidFarms                | []LiquidFarm | TBD                                            |

## LiquidFarms

`LiquidFarms` is a list of `LiquidFarm`, where a `LiquidFarm` is corresponding to a specific pool with `PoolId`. 
A single `LiquidFarm` can exist for a given pool.


```go
type LiquidFarm struct {
	PoolId               uint64
	MinimumFarmAmount sdk.Int
	MinimumBidAmount     sdk.Int
}
```
