package cli

import (
	"strings"

	"github.com/crescent-network/crescent/v2/x/claim/types"
)

// NormalizeConditionType normalizes specified condition type.
func NormalizeConditionType(ConditionType string) types.ConditionType {
	switch strings.ToLower(ConditionType) {
	case "d", "deposit":
		return types.ConditionTypeDeposit
	case "s", "swap":
		return types.ConditionTypeSwap
	case "ls", "liquidstake":
		return types.ConditionTypeLiquidStake
	case "v", "vote":
		return types.ConditionTypeVote
	default:
		return types.ConditionTypeUnspecified
	}
}
