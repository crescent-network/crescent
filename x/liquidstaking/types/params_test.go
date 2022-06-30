package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

func TestParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

	params := types.DefaultParams()

	paramsStr := `liquid_bond_denom: bstake
whitelisted_validators: []
unstake_fee_rate: "0.001000000000000000"
min_liquid_staking_amount: "1000000"
`
	require.Equal(t, paramsStr, params.String())

	params.WhitelistedValidators = []types.WhitelistedValidator{
		{
			ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			TargetWeight:     sdk.NewInt(10),
		},
	}
	paramsStr = `liquid_bond_denom: bstake
whitelisted_validators:
- validator_address: cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv
  target_weight: "10"
unstake_fee_rate: "0.001000000000000000"
min_liquid_staking_amount: "1000000"
`
	require.Equal(t, paramsStr, params.String())
}

func TestWhitelistedValsMap(t *testing.T) {
	params := types.DefaultParams()
	require.EqualValues(t, params.WhitelistedValsMap(), types.WhitelistedValsMap{})

	params.WhitelistedValidators = []types.WhitelistedValidator{
		whitelistedValidators[0],
		whitelistedValidators[1],
	}

	wvm := params.WhitelistedValsMap()
	require.Len(t, params.WhitelistedValidators, len(wvm))

	for _, wv := range params.WhitelistedValidators {
		require.EqualValues(t, wvm[wv.ValidatorAddress], wv)
		require.True(t, wvm.IsListed(wv.ValidatorAddress))
	}

	require.False(t, wvm.IsListed("notExistedAddr"))
}

func TestValidateWhitelistedValidators(t *testing.T) {
	for _, tc := range []struct {
		name     string
		malleate func(*types.Params)
		errStr   string
	}{
		{
			"valid default params",
			func(params *types.Params) {},
			"",
		},
		{
			"blank liquid bond denom",
			func(params *types.Params) {
				params.LiquidBondDenom = ""
			},
			"liquid bond denom cannot be blank",
		},
		{
			"invalid liquid bond denom",
			func(params *types.Params) {
				params.LiquidBondDenom = "a"
			},
			"invalid denom: a",
		},
		{
			"duplicated whitelisted validators",
			func(params *types.Params) {
				params.WhitelistedValidators = []types.WhitelistedValidator{
					{
						ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
						TargetWeight:     sdk.NewInt(10),
					},
					{
						ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
						TargetWeight:     sdk.NewInt(10),
					},
				}
			},
			"liquidstaking validator cannot be duplicated: cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
		},
		{
			"invalid whitelisted validator address",
			func(params *types.Params) {
				params.WhitelistedValidators = []types.WhitelistedValidator{
					{
						ValidatorAddress: "invalidaddr",
						TargetWeight:     sdk.NewInt(10),
					},
				}
			},
			"decoding bech32 failed: invalid separator index -1",
		},
		{
			"nil whitelisted validator target weight",
			func(params *types.Params) {
				params.WhitelistedValidators = []types.WhitelistedValidator{
					{
						ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
						TargetWeight:     sdk.Int{},
					},
				}
			},
			"liquidstaking validator target weight must not be nil",
		},
		{
			"negative whitelisted validator target weight",
			func(params *types.Params) {
				params.WhitelistedValidators = []types.WhitelistedValidator{
					{
						ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
						TargetWeight:     sdk.NewInt(-1),
					},
				}
			},
			"liquidstaking validator target weight must be positive: -1",
		},
		{
			"zero whitelisted validator target weight",
			func(params *types.Params) {
				params.WhitelistedValidators = []types.WhitelistedValidator{
					{
						ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
						TargetWeight:     sdk.ZeroInt(),
					},
				}
			},
			"liquidstaking validator target weight must be positive: 0",
		},
		{
			"nil unstake fee rate",
			func(params *types.Params) {
				params.UnstakeFeeRate = sdk.Dec{}
			},
			"unstake fee rate must not be nil",
		},
		{
			"negative unstake fee rate",
			func(params *types.Params) {
				params.UnstakeFeeRate = sdk.NewDec(-1)
			},
			"unstake fee rate must not be negative: -1.000000000000000000",
		},
		{
			"too large unstake fee rate",
			func(params *types.Params) {
				params.UnstakeFeeRate = sdk.MustNewDecFromStr("1.0000001")
			},
			"unstake fee rate too large: 1.000000100000000000",
		},
		{
			"nil min liquid staking amount",
			func(params *types.Params) {
				params.MinLiquidStakingAmount = sdk.Int{}
			},
			"min liquid staking amount must not be nil",
		},
		{
			"negative min liquid staking amount",
			func(params *types.Params) {
				params.MinLiquidStakingAmount = sdk.NewInt(-1)
			},
			"min liquid staking amount must not be negative: -1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.malleate(&params)
			err := params.Validate()
			if tc.errStr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.errStr)
			}
		})
	}
}
