package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v2/x/claim/client/cli"
	"github.com/crescent-network/crescent/v2/x/claim/types"
)

func TestNormalizeConditionType(t *testing.T) {
	testCases := []struct {
		name string
		arg  string
		want types.ConditionType
	}{
		{"invalid", "order", types.ConditionTypeUnspecified},
		{"deposit", "deposit", types.ConditionTypeDeposit},
		{"deposit", "d", types.ConditionTypeDeposit},
		{"swap", "swap", types.ConditionTypeSwap},
		{"swap", "s", types.ConditionTypeSwap},
		{"liquidstake", "liquidstake", types.ConditionTypeLiquidStake},
		{"liquidstake", "ls", types.ConditionTypeLiquidStake},
		{"vote", "vote", types.ConditionTypeVote},
		{"vote", "v", types.ConditionTypeVote},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, cli.NormalizeConditionType(tc.arg))
		})
	}
}
