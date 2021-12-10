package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/farming/x/liquidstaking/types"
)

var (
	whitelistedValidators = []types.WhitelistedValidator{
		{
			ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			Weight:           sdk.OneDec(),
		},
	}
)

func TestParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

	defaultParams := types.DefaultParams()

	paramsStr := `whitelisted_validators: []
unstake_fee_rate: "0.001000000000000000"
`
	require.Equal(t, paramsStr, defaultParams.String())
}

func TestValidateWhitelistedValidators(t *testing.T) {
	err := types.ValidateWhitelistedValidators([]types.WhitelistedValidator{whitelistedValidators[0]})
	require.NoError(t, err)
}
