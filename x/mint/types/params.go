package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	squadtypes "github.com/cosmosquad-labs/squad/types"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyMintDenom          = []byte("MintDenom")
	KeyBlockTimeThreshold = []byte("BlockTimeThreshold")
	KeyInflationSchedules = []byte("InflationSchedules")

	// TODO: fix
	DefaultInflationSchedules = []InflationSchedule{
		{
			StartTime: squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"),
			EndTime:   squadtypes.MustParseRFC3339("2023-01-01T00:00:00Z"),
			Amount:    sdk.NewInt(300000000000000),
		},
		{
			StartTime: squadtypes.MustParseRFC3339("2023-01-01T00:00:00Z"),
			EndTime:   squadtypes.MustParseRFC3339("2024-01-01T00:00:00Z"),
			Amount:    sdk.NewInt(200000000000000),
		},
	}
	DefaultBlockTimeThreshold = 10 * time.Second
)

// ParamTable for mint module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	// TODO: []InflationSchedule or InflationSchedules
	mintDenom string, BlockTimeThreshold time.Duration, InflationSchedules []InflationSchedule,
) Params {

	return Params{
		MintDenom:          mintDenom,
		BlockTimeThreshold: BlockTimeThreshold,
		InflationSchedules: InflationSchedules,
	}
}

// default mint module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:          sdk.DefaultBondDenom,
		BlockTimeThreshold: DefaultBlockTimeThreshold,
		InflationSchedules: DefaultInflationSchedules,
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateBlockTimeThreshold(p.BlockTimeThreshold); err != nil {
		return err
	}

	// TODO: InflationSchedules
	return nil

}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
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
	for _, inflation := range v {
		if !inflation.Amount.IsPositive() {
			return fmt.Errorf("inflation schedule amount must be positive: %s", v)
		}
		if !inflation.EndTime.After(inflation.StartTime) {
			return fmt.Errorf("end time %s must be greater than start time %s", inflation.EndTime.Format(time.RFC3339), inflation.StartTime.Format(time.RFC3339))
		}
	}
	return nil
}
