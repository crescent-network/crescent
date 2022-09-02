package types_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

func TestParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

	defaultParams := types.DefaultParams()

	paramsStr := `incentive_budget_address: cosmos1ddn66jv0sjpmck0ptegmhmqtn35qsg2vxyk2hn9sqf4qxtzqz3sqanrtcm
deposit_amount:
- denom: stake
  amount: "1000000000"
common:
  min_open_ratio: "0.500000000000000000"
  min_open_depth_ratio: "0.100000000000000000"
  max_downtime: 20
  max_total_downtime: 100
  min_hours: 16
  min_days: 22
incentive_pairs: []
`

	require.Equal(t, paramsStr, defaultParams.String())
}

func TestParamsValidate(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())

	testCases := []struct {
		name        string
		configure   func(*types.Params)
		expectedErr string
	}{
		{
			"Default",
			func(params *types.Params) {
			},
			"",
		},
		{
			"EmptyDepositAmount",
			func(params *types.Params) {
				params.DepositAmount = sdk.NewCoins()
			},
			"",
		},
		{
			"NegativeDepositAmount",
			func(params *types.Params) {
				params.DepositAmount = sdk.Coins{
					sdk.Coin{
						Denom:  "stake",
						Amount: sdk.NewInt(-1),
					},
				}
			},
			"coin -1stake amount is not positive",
		},
		{
			"EmptyBudgetAddr",
			func(params *types.Params) {
				params.IncentiveBudgetAddress = ""
			},
			"incentive budget address must not be empty",
		},
		{
			"WrongBudgetAddr",
			func(params *types.Params) {
				params.IncentiveBudgetAddress = "addr1"
			},
			"invalid account address: addr1",
		},
		{
			"IncentivePair",
			func(params *types.Params) {
				params.IncentivePairs = []types.IncentivePair{
					{
						PairId:          1,
						IncentiveWeight: sdk.MustNewDecFromStr("0.1"),
					},
				}
				a, _ := json.Marshal(params.IncentivePairs[0])
				fmt.Println(string(a))
			},
			"",
		},
		{
			"DuplicatedIncentivePair",
			func(params *types.Params) {
				params.IncentivePairs = []types.IncentivePair{
					{
						PairId:          1,
						IncentiveWeight: sdk.MustNewDecFromStr("0.1"),
					},
					{
						PairId:          1,
						IncentiveWeight: sdk.MustNewDecFromStr("0.2"),
					},
				}
			},
			"incentive pair id cannot be duplicated: 1",
		},
		{
			"MultipleIncentivePairs",
			func(params *types.Params) {
				params.IncentivePairs = []types.IncentivePair{
					{
						PairId:          1,
						IncentiveWeight: sdk.MustNewDecFromStr("0.1"),
					},
					{
						PairId:          2,
						IncentiveWeight: sdk.MustNewDecFromStr("0.2"),
					},
				}
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.configure(&params)
			err := params.Validate()

			var err2 error
			for _, p := range params.ParamSetPairs() {
				err := p.ValidatorFn(reflect.ValueOf(p.Value).Elem().Interface())
				if err != nil {
					err2 = err
					break
				}
			}
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				require.EqualError(t, err2, tc.expectedErr)
			} else {
				require.Nil(t, err)
				require.Nil(t, err2)
			}
		})
	}
}
