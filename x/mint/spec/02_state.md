<!--
order: 2
-->

# State

## InflationSchedules

```go
type InflationSchedules []InflationPeriod

type InflationPeriod struct {
    StartTime time.Time
    EndTime   time.Time
    Amount    sdk.Int
}
```

## LastBlockTime

## Params

Minting params are held in the global params store.

- Params: `mint/params -> ProtocolBuffer(params)`
