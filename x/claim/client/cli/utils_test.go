package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/claim/client/cli"
	"github.com/cosmosquad-labs/squad/x/claim/types"
)

func TestNormalizeActionType(t *testing.T) {
	testCases := []struct {
		name string
		arg  string
		want types.ActionType
	}{
		{"invalid", "order", types.ActionTypeUnspecified},
		{"deposit", "deposit", types.ActionTypeDeposit},
		{"deposit", "d", types.ActionTypeDeposit},
		{"swap", "swap", types.ActionTypeSwap},
		{"swap", "s", types.ActionTypeSwap},
		{"farming", "farming", types.ActionTypeFarming},
		{"farming", "f", types.ActionTypeFarming},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, cli.NormalizeActionType(tc.arg))
		})
	}
}
