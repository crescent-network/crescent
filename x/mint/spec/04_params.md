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


## InflationSchedules

InflationSchedules is a list struct of `InflationSchedules`.

`InflationSchedule` defines the start and end time of the inflation period, and the amount of inflation during that period.

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