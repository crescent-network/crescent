package types

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	farmingtypes "github.com/crescent-network/crescent/v5/x/farming/types"
)

const (
	AddressType                             = farmingtypes.AddressType32Bytes
	ClaimableIncentiveReserveAccName string = "ClaimableIncentiveReserveAcc"
)

// Parameter store keys
var (
	KeyIncentiveBudgetAddress = []byte("IncentiveBudgetAddress")
	KeyDepositAmount          = []byte("DepositAmount")
	KeyCommon                 = []byte("Common")
	KeyIncentivePairs         = []byte("IncentivePairs")

	DefaultIncentiveBudgetAddress = farmingtypes.DeriveAddress(AddressType, farmingtypes.ModuleName, "ecosystem_incentive_mm")
	DefaultDepositAmount          = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000)))
	DefaultCommon                 = Common{
		MinOpenRatio:      sdk.MustNewDecFromStr("0.5"),
		MinOpenDepthRatio: sdk.MustNewDecFromStr("0.1"),
		MaxDowntime:       uint32(20),
		MaxTotalDowntime:  uint32(100),
		MinHours:          uint32(16),
		MinDays:           uint32(22),
	}

	ClaimableIncentiveReserveAcc = farmingtypes.DeriveAddress(AddressType, ModuleName, ClaimableIncentiveReserveAccName)
	DepositReserveAcc            = sdk.AccAddress(crypto.AddressHash([]byte(ModuleName)))
)

var _ paramstypes.ParamSet = (*Params)(nil)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the default marketmaker module parameters.
func DefaultParams() Params {
	return Params{
		IncentiveBudgetAddress: DefaultIncentiveBudgetAddress.String(),
		DepositAmount:          DefaultDepositAmount,
		Common:                 DefaultCommon,
		IncentivePairs:         []IncentivePair{},
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyIncentiveBudgetAddress, &p.IncentiveBudgetAddress, validateIncentiveBudgetAddress),
		paramstypes.NewParamSetPair(KeyDepositAmount, &p.DepositAmount, validateDepositAmount),
		paramstypes.NewParamSetPair(KeyCommon, &p.Common, validateCommon),
		paramstypes.NewParamSetPair(KeyIncentivePairs, &p.IncentivePairs, validateIncentivePairs),
	}
}

func (p Params) IncentiveBudgetAcc() sdk.AccAddress {
	acc, _ := sdk.AccAddressFromBech32(p.IncentiveBudgetAddress)
	return acc
}

func (p Params) IncentivePairsMap() map[uint64]IncentivePair {
	iMap := make(map[uint64]IncentivePair)
	for _, pair := range p.IncentivePairs {
		iMap[pair.PairId] = pair
	}
	return iMap
}

// String returns a human-readable string representation of the parameters.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates parameters.
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.IncentiveBudgetAddress, validateIncentiveBudgetAddress},
		{p.DepositAmount, validateDepositAmount},
		{p.IncentivePairs, validateIncentivePairs},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

func validateDepositAmount(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := v.Validate(); err != nil {
		return err
	}

	return nil
}

func validateIncentiveBudgetAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == "" {
		return fmt.Errorf("incentive budget address must not be empty")
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid account address: %v", v)
	}

	return nil
}

func validateCommon(i interface{}) error {
	_, ok := i.(Common)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateIncentivePairs(i interface{}) error {
	ips, ok := i.([]IncentivePair)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// pair id duplication checking
	pairMap := map[uint64]struct{}{}
	for _, ip := range ips {
		if _, ok := pairMap[ip.PairId]; ok {
			return fmt.Errorf("incentive pair id cannot be duplicated: %d", ip.PairId)
		}
		pairMap[ip.PairId] = struct{}{}
	}
	return nil
}
