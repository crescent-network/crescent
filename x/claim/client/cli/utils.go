package cli

import (
	"strings"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

// NormalizeConditionType normalizes specified action type.
func NormalizeConditionType(ConditionType string) types.ConditionType {
	switch strings.ToLower(ConditionType) {
	case "d", "deposit":
		return types.ConditionTypeDeposit
	case "s", "swap":
		return types.ConditionTypeSwap
	case "f", "farming":
		return types.ConditionTypeFarming
	default:
		return types.ConditionTypeUnspecified
	}
}
