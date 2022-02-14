package cli

import (
	"strings"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

// NormalizeActionType normalizes specified action type.
func NormalizeActionType(actionType string) types.ActionType {
	switch strings.ToLower(actionType) {
	case "d", "deposit":
		return types.ActionTypeDeposit
	case "s", "swap":
		return types.ActionTypeSwap
	case "f", "farming":
		return types.ActionTypeFarming
	default:
		return types.ActionTypeUnspecified
	}
}
