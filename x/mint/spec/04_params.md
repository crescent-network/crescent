<!--
order: 4
-->

# Parameters

The mint module contains the following parameters:

| Key                  | Type                | Example |
|----------------------|---------------------|---------|
| mint_denom           | string              | "stake" |
| block_time_threshold | time.duration       | "10s"   |
| inflation_schedules  | []InflationSchedule |         |


## MintDenom

MintDenom is the denomination of the coin to be minted.

## BlockTimeThreshold

If the difference in block time from the `LastBlockTime` is greater than `BlockTimeThreshold`, the block only generates inflation as long as `BlockTimeThreshold.
so actual minted amount could be less than the defined `InflationSchedule.Amount` depending on the number of times blocks having the block time length over `BlockTimeThreshold` occurs.

This is to prevent inflationary manipulation attacks caused by stopping chains or block time manipulation.

## InflationSchedules

InflationSchedules is a list of `InflationSchedules`, Those inflation schedules cannot overlap, start time is inclusive for the schedule and end time is exclusive so the end times and other start times could be the same.

`InflationSchedule` defines the start and end time of the inflation period, and the amount of inflation during that period.

`InflationSchedule.Amount` should be over the inflation schedule duration seconds to avoid decimal loss

```go
type InflationSchedule struct {
	// start_time is a start date time of the inflation period
    StartTime time.Time
	// end_time is a start date time of the inflation period
    EndTime   time.Time
	// amount is the amount of inflation during that period.
    Amount    sdk.Int
}
```

Example of inflation schedules

```go
ExampleInflationSchedules = []InflationSchedule{
    {
        StartTime: squadtypes.ParseTime("2022-01-01T00:00:00Z"),
        EndTime:   squadtypes.ParseTime("2023-01-01T00:00:00Z"),
        Amount:    sdk.NewInt(300000000000000),
    },
    {
        StartTime: squadtypes.ParseTime("2023-01-01T00:00:00Z"),
        EndTime:   squadtypes.ParseTime("2024-01-01T00:00:00Z"),
        Amount:    sdk.NewInt(200000000000000),
    },
}
```