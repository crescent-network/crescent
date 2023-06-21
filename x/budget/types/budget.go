package types

import (
	"fmt"
	"regexp"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	reBudgetNameString = fmt.Sprintf(`[a-zA-Z][a-zA-Z0-9-]{0,%d}`, MaxBudgetNameLength-1)
	reBudgetName       = regexp.MustCompile(fmt.Sprintf(`^%s$`, reBudgetNameString))
)

// String returns a human-readable string representation of the budget.
func (budget Budget) String() string {
	out, _ := budget.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a Budget.
func (budget Budget) MarshalYAML() (interface{}, error) {
	bz, err := codec.MarshalYAML(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), &budget)
	if err != nil {
		return nil, err
	}
	return string(bz), err
}

// Validate validates the budget.
func (budget Budget) Validate() error {
	if err := ValidateName(budget.Name); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(budget.DestinationAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid destination address %s: %v", budget.DestinationAddress, err)
	}

	if _, err := sdk.AccAddressFromBech32(budget.SourceAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid source address %s: %v", budget.SourceAddress, err)
	}

	if budget.DestinationAddress == budget.SourceAddress {
		return sdkerrors.Wrapf(ErrSameSourceDestinationAddr, "budget source address and destination address cannot be the same %s", budget.SourceAddress)
	}

	if !budget.EndTime.After(budget.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidStartEndTime, "end time %s must be greater than start time %s", budget.EndTime, budget.StartTime)
	}

	if !budget.Rate.IsPositive() {
		return sdkerrors.Wrapf(ErrInvalidBudgetRate, "budget rate must not be positive: %s", budget.Rate)
	} else if budget.Rate.GT(sdk.OneDec()) {
		return sdkerrors.Wrapf(ErrInvalidBudgetRate, "budget rate must not exceed 1: %s", budget.Rate)
	}

	return nil
}

// Collectible validates the budget has reached its start time and that the end time has not elapsed.
func (budget Budget) Collectible(blockTime time.Time) bool {
	return DateRangeIncludes(budget.StartTime, budget.EndTime, blockTime)
}

// CollectibleBudgets returns only the valid and started and not expired budgets based on the given block time.
func CollectibleBudgets(budgets []Budget, blockTime time.Time) (collectibleBudgets []Budget) {
	for _, budget := range budgets {
		if budget.Collectible(blockTime) {
			collectibleBudgets = append(collectibleBudgets, budget)
		}
	}
	return
}

// ValidateName is the default validation function for Budget.Name.
// A budget name only allows alphabet letters(`A-Z, a-z`), digit numbers(`0-9`), and `-`.
// It doesn't allow spaces and the maximum length is 50 characters.
func ValidateName(name string) error {
	if !reBudgetName.MatchString(name) {
		return sdkerrors.Wrap(ErrInvalidBudgetName, name)
	}
	return nil
}

// BudgetsBySource defines the total rate of budget lists.
type BudgetsBySource struct {
	Budgets         []Budget
	CollectionCoins []sdk.Coins
	TotalRate       sdk.Dec
}

type BudgetsBySourceMap map[string]BudgetsBySource

// GetBudgetsBySourceMap returns BudgetsBySourceMap that has a list of budgets and their total rate
// which contain the same SourceAddress. It can be used to track of what budgets are available with SourceAddress
// and validate their total rate.
func GetBudgetsBySourceMap(budgets []Budget) (BudgetsBySourceMap, []string) {
	budgetsMap := make(BudgetsBySourceMap)
	budgetSources := []string{}
	for _, budget := range budgets {
		if budgetsBySource, ok := budgetsMap[budget.SourceAddress]; ok {
			budgetsBySource.TotalRate = budgetsBySource.TotalRate.Add(budget.Rate)
			budgetsBySource.Budgets = append(budgetsBySource.Budgets, budget)
			budgetsMap[budget.SourceAddress] = budgetsBySource
		} else {
			budgetsMap[budget.SourceAddress] = BudgetsBySource{
				Budgets:   []Budget{budget},
				TotalRate: budget.Rate,
			}
			budgetSources = append(budgetSources, budget.SourceAddress)
		}
	}
	return budgetsMap, budgetSources
}
