package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/claim/client/cli"
	"github.com/cosmosquad-labs/squad/x/claim/types"
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
		{"farming", "farming", types.ConditionTypeFarming},
		{"farming", "f", types.ConditionTypeFarming},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, cli.NormalizeConditionType(tc.arg))
		})
	}
}
