package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyMintDenom          = []byte("MintDenom")
	KeyBlockTimeThreshold = []byte("BlockTimeThreshold")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, BlockTimeThreshold time.Duration,
) Params {

	return Params{
		MintDenom:          mintDenom,
		BlockTimeThreshold: BlockTimeThreshold,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:          sdk.DefaultBondDenom,
		BlockTimeThreshold: time.Duration(10 * time.Second),
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
		return fmt.Errorf("unbonding time must be positive: %d", v)
	}

	return nil
}
