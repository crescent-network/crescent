package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

var (
	whitelistedValidators = []types.WhitelistedValidator{
		{
			ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			TargetWeight:     sdk.NewInt(10),
		},
		{
			ValidatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			TargetWeight:     sdk.NewInt(1),
		},
		{
			ValidatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			TargetWeight:     sdk.NewInt(-1),
		},
		{
			ValidatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
			TargetWeight:     sdk.NewInt(0),
		},
	}
)

func TestParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

	defaultParams := types.DefaultParams()

	paramsStr := `liquid_bond_denom: bstake
whitelisted_validators: []
unstake_fee_rate: "0.001000000000000000"
min_liquid_staking_amount: "1000000"
`
	require.Equal(t, paramsStr, defaultParams.String())
}

func TestValidateWhitelistedValidators(t *testing.T) {
	err := types.ValidateWhitelistedValidators([]types.WhitelistedValidator{whitelistedValidators[0]})
	require.NoError(t, err)
	err = types.ValidateWhitelistedValidators([]types.WhitelistedValidator{whitelistedValidators[1]})
	require.NoError(t, err)
	err = types.ValidateWhitelistedValidators([]types.WhitelistedValidator{whitelistedValidators[0], whitelistedValidators[0]})
	require.Error(t, err)
	err = types.ValidateWhitelistedValidators([]types.WhitelistedValidator{whitelistedValidators[2]})
	require.Error(t, err)
	err = types.ValidateWhitelistedValidators([]types.WhitelistedValidator{whitelistedValidators[3]})
	require.Error(t, err)
}

// TODO: add testcodes for params
