package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/tendermint/farming/x/farming/types"
)

func TestParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

	defaultParams := types.DefaultParams()

	paramsStr := `private_plan_creation_fee:
- denom: stake
  amount: "100000000"
staking_creation_fee: []
epoch_days: 0
`
	require.Equal(t, paramsStr, defaultParams.String())
}
