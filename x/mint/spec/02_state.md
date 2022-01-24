<!--
order: 2
-->

# State

## InflationSchedules

InflationSchedule defines the start and end time of the inflation period, and the amount of inflation during that period.

```go
type InflationSchedules []InflationPeriod

type InflationSchedule struct {
	// start_time is a start date time of the inflation period
    StartTime time.Time
	// end_time is a start date time of the inflation period
    EndTime   time.Time
	// amount is the amount of inflation during that period.
    Amount    sdk.Int
}
```

## LastBlockTime

LastBlockTime defines block time of the last block's header, It used to calculate inflation.

- LastBlockTimeKey: `0x90 -> sdk.FormatTimeBytes(time.Time)`

## Params

Minting params are held in the global params store.

- Params: `mint/params -> ProtocolBuffer(params)`
