package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	utils "github.com/crescent-network/crescent/v2/types"
)

// Parameter store keys
var (
	KeyMintDenom          = []byte("MintDenom")
	KeyMintPoolAddress    = []byte("MintPoolAddress")
	KeyBlockTimeThreshold = []byte("BlockTimeThreshold")
	KeyInflationSchedules = []byte("InflationSchedules")

	DefaultBlockTimeThreshold = 10 * time.Second

	// DefaultMintPoolAddress is the fee collector of the auth module such as the mint module of the original cosmos-sdk
	DefaultMintPoolAddress = authtypes.NewModuleAddress(authtypes.FeeCollectorName)
	// MintModuleAcc is customized mint pool, If the fee collector as default mint pool is used,
	// there is a problem that the fee and inflation are mixed with the fee collector.
	// In order to solve this problem, mint pool can be set to CustomMintPoolAddress
	// and staking reward can be sent to the fee collector through Budget module functionality
	MintModuleAcc = authtypes.NewModuleAddress(ModuleName)

	// DefaultInflationSchedules is example of inflation schedules, It could be rearranged on genesis or gov
	DefaultInflationSchedules = []InflationSchedule{
		{
			StartTime: utils.ParseTime("2022-01-01T00:00:00Z"),
			EndTime:   utils.ParseTime("2023-01-01T00:00:00Z"),
			Amount:    sdk.NewInt(300000000000000),
		},
		{
			StartTime: utils.ParseTime("2023-01-01T00:00:00Z"),
			EndTime:   utils.ParseTime("2024-01-01T00:00:00Z"),
			Amount:    sdk.NewInt(200000000000000),
		},
	}
)

// ParamTable for mint module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// default mint module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:          sdk.DefaultBondDenom,
		MintPoolAddress:    DefaultMintPoolAddress.String(),
		BlockTimeThreshold: DefaultBlockTimeThreshold,
		InflationSchedules: DefaultInflationSchedules,
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateMintPoolAddress(p.MintPoolAddress); err != nil {
		return err
	}
	if err := validateBlockTimeThreshold(p.BlockTimeThreshold); err != nil {
		return err
	}
	if err := validateInflationSchedules(p.InflationSchedules); err != nil {
		return err
	}
	return nil

}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyMintPoolAddress, &p.MintPoolAddress, validateMintPoolAddress),
		paramtypes.NewParamSetPair(KeyBlockTimeThreshold, &p.BlockTimeThreshold, validateBlockTimeThreshold),
		paramtypes.NewParamSetPair(KeyInflationSchedules, &p.InflationSchedules, validateInflationSchedules),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateMintPoolAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if _, err := sdk.AccAddressFromBech32(v); err != nil {
		return fmt.Errorf("invalid mint pool address: %w", err)
	}

	return nil
}

func validateBlockTimeThreshold(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("block time threshold must be positive: %d", v)
	}

	return nil
}

func validateInflationSchedules(i interface{}) error {
	v, ok := i.([]InflationSchedule)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for j, inflation := range v {
		if !inflation.Amount.IsPositive() {
			return fmt.Errorf("inflation schedule amount must be positive: %s", inflation.Amount)
		}
		if inflation.Amount.LT(sdk.NewInt(int64(inflation.EndTime.Sub(inflation.StartTime).Seconds()))) {
			return fmt.Errorf("inflation amount too small, it should be over period duration seconds: %s", inflation.Amount)
		}
		if !inflation.EndTime.After(inflation.StartTime) {
			return fmt.Errorf("inflation end time %s must be greater than start time %s", inflation.EndTime.Format(time.RFC3339), inflation.StartTime.Format(time.RFC3339))
		}
		for _, inflationOther := range v[j+1:] {
			if utils.DateRangesOverlap(inflation.StartTime, inflation.EndTime, inflationOther.StartTime, inflationOther.EndTime) {
				return fmt.Errorf("inflation periods cannot be overlapped %s ~ %s with %s ~ %s", inflation.StartTime.Format(time.RFC3339), inflation.EndTime.Format(time.RFC3339), inflationOther.StartTime.Format(time.RFC3339), inflationOther.EndTime.Format(time.RFC3339))
			}
		}
	}
	return nil
}
