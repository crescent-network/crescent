package types

import (
	"fmt"
	"time"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyRewardsAuctionDuration      = []byte("RewardsAuctionDuration")
	KeyMaxNumRecentRewardsAuctions = []byte("MaxNumRecentRewardsAuctions")
)

// Default parameters
var (
	DefaultRewardsAuctionDuration      = time.Hour
	DefaultMaxNumRecentRewardsAuctions = uint32(24)
)

var _ paramstypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for the module
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		RewardsAuctionDuration:      DefaultRewardsAuctionDuration,
		MaxNumRecentRewardsAuctions: DefaultMaxNumRecentRewardsAuctions,
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyRewardsAuctionDuration, &p.RewardsAuctionDuration, validateRewardsAuctionDuration),
		paramstypes.NewParamSetPair(KeyMaxNumRecentRewardsAuctions, &p.MaxNumRecentRewardsAuctions, validateMaxNumRecentRewardsAuctions),
	}
}

// Validate validates the set of parameters
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.RewardsAuctionDuration, validateRewardsAuctionDuration},
		{p.MaxNumRecentRewardsAuctions, validateMaxNumRecentRewardsAuctions},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

func validateRewardsAuctionDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("rewards auction duration must be positive: %d", v)
	}
	return nil
}

func validateMaxNumRecentRewardsAuctions(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
