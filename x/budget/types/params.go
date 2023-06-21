package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// MaxBudgetNameLength is the maximum length of the name of each budget.
	MaxBudgetNameLength int = 50
	// DefaultEpochBlocks is the default epoch blocks.
	DefaultEpochBlocks uint32 = 1
)

// Parameter store keys
var (
	KeyBudgets     = []byte("Budgets")
	KeyEpochBlocks = []byte("EpochBlocks")
)

var _ paramstypes.ParamSet = (*Params)(nil)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the default budget module parameters.
func DefaultParams() Params {
	return Params{
		Budgets:     []Budget{},
		EpochBlocks: DefaultEpochBlocks,
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyBudgets, &p.Budgets, ValidateBudgets),
		paramstypes.NewParamSetPair(KeyEpochBlocks, &p.EpochBlocks, ValidateEpochBlocks),
	}
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
		{p.Budgets, ValidateBudgets},
		{p.EpochBlocks, ValidateEpochBlocks},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

// ValidateBudgets validates budget name and total rate.
// The total rate of budgets with the same source address must not exceed 1.
func ValidateBudgets(input interface{}) error {
	budgets, ok := input.([]Budget)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", input)
	}
	names := make(map[string]bool)
	for _, budget := range budgets {
		err := budget.Validate()
		if err != nil {
			return err
		}
		if _, ok := names[budget.Name]; ok {
			return sdkerrors.Wrap(ErrDuplicateBudgetName, budget.Name)
		}
		names[budget.Name] = true
	}
	budgetsBySourceMap, budgetSources := GetBudgetsBySourceMap(budgets)
	for _, source := range budgetSources {
		if budgetsBySourceMap[source].TotalRate.GT(sdk.OneDec()) {
			// If the TotalRate of Budgets with the same source address exceeds 1,
			// recalculate and verify the TotalRate of Budgets with overlapping time ranges.
			for i, budget := range budgetsBySourceMap[source].Budgets {
				totalRate := budget.Rate
				for j, budgetToCheck := range budgetsBySourceMap[source].Budgets {
					if i != j && budgetToCheck.Collectible(budget.StartTime) {
						totalRate = totalRate.Add(budgetToCheck.Rate)
					}
				}
				if totalRate.GT(sdk.OneDec()) {
					return sdkerrors.Wrapf(
						ErrInvalidTotalBudgetRate,
						"total rate for source address %s must not exceed 1: %v", source, totalRate)
				}
			}

		}
	}
	return nil
}

// ValidateEpochBlocks validates epoch blocks.
func ValidateEpochBlocks(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
