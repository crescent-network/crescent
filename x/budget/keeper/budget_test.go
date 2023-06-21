package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/crescent-network/crescent/v5/app"
	appparams "github.com/crescent-network/crescent/v5/app/params"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

func (suite *KeeperTestSuite) TestCollectBudgets() {
	for _, tc := range []struct {
		name           string
		budgets        []types.Budget
		epochBlocks    uint32
		accAsserts     []sdk.AccAddress
		balanceAsserts []sdk.Coins
		expectErr      error
	}{
		{
			"basic budgets case",
			suite.budgets[:4],
			types.DefaultEpochBlocks,
			[]sdk.AccAddress{
				suite.destinationAddrs[0],
				suite.destinationAddrs[1],
				suite.destinationAddrs[2],
				suite.destinationAddrs[3],
				suite.sourceAddrs[0],
				suite.sourceAddrs[1],
				suite.sourceAddrs[2],
			},
			[]sdk.Coins{
				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				{},
				{},
				{},
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
			},
			nil,
		},
		{
			"only expired budget case",
			[]types.Budget{suite.budgets[3]},
			types.DefaultEpochBlocks,
			[]sdk.AccAddress{
				suite.destinationAddrs[3],
				suite.sourceAddrs[2],
			},
			[]sdk.Coins{
				{},
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
			},
			nil,
		},
		{
			"source has small balances case",
			suite.budgets[4:6],
			types.DefaultEpochBlocks,
			[]sdk.AccAddress{
				suite.destinationAddrs[0],
				suite.destinationAddrs[1],
				suite.sourceAddrs[3],
			},
			[]sdk.Coins{
				mustParseCoinsNormalized("1denom2,1denom3,500000000stake"),
				mustParseCoinsNormalized("1denom2,1denom3,500000000stake"),
				mustParseCoinsNormalized("1denom1,1denom3"),
			},
			nil,
		},
		{
			"source has small balances case 2",
			[]types.Budget{suite.budgets[6]},
			types.DefaultEpochBlocks,
			[]sdk.AccAddress{
				suite.destinationAddrs[4],
				suite.sourceAddrs[4],
			},
			[]sdk.Coins{
				mustParseCoinsNormalized(""),
				mustParseCoinsNormalized("1denom1,2denom2"),
			},
			nil,
		},
		{
			"empty source case",
			[]types.Budget{suite.budgets[7]},
			types.DefaultEpochBlocks,
			[]sdk.AccAddress{
				suite.destinationAddrs[5],
				suite.sourceAddrs[5],
			},
			[]sdk.Coins{
				{},
				{},
			},
			nil,
		},
		{
			"none budgets case",
			nil,
			types.DefaultEpochBlocks,
			[]sdk.AccAddress{
				suite.destinationAddrs[0],
				suite.destinationAddrs[1],
				suite.destinationAddrs[2],
				suite.destinationAddrs[3],
				suite.sourceAddrs[0],
				suite.sourceAddrs[1],
				suite.sourceAddrs[2],
				suite.sourceAddrs[3],
			},
			[]sdk.Coins{
				{},
				{},
				{},
				{},
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake"),
			},
			nil,
		},
		{
			"disabled budget epoch",
			nil,
			0,
			[]sdk.AccAddress{
				suite.destinationAddrs[0],
				suite.destinationAddrs[1],
				suite.destinationAddrs[2],
				suite.destinationAddrs[3],
				suite.sourceAddrs[0],
				suite.sourceAddrs[1],
				suite.sourceAddrs[2],
				suite.sourceAddrs[3],
			},
			[]sdk.Coins{
				{},
				{},
				{},
				{},
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake"),
			},
			nil,
		},
		{
			"disabled budget epoch with budgets",
			suite.budgets[:4],
			0,
			[]sdk.AccAddress{
				suite.destinationAddrs[0],
				suite.destinationAddrs[1],
				suite.destinationAddrs[2],
				suite.destinationAddrs[3],
				suite.sourceAddrs[0],
				suite.sourceAddrs[1],
				suite.sourceAddrs[2],
				suite.sourceAddrs[3],
			},
			[]sdk.Coins{
				{},
				{},
				{},
				{},
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake"),
			},
			nil,
		},
		{
			"relay-budget",
			suite.budgets[8:13],
			1,
			[]sdk.AccAddress{
				suite.destinationAddrs[1],
				suite.destinationAddrs[2],
				suite.destinationAddrs[3],
				suite.destinationAddrs[4],
				suite.destinationAddrs[5],
				suite.sourceAddrs[1],
			},
			[]sdk.Coins{
				{},
				mustParseCoinsNormalized("250000000denom1,250000000denom2,250000000denom3,250000000stake"),
				mustParseCoinsNormalized("397500000denom1,397500000denom2,397500000denom3,397500000stake"),
				mustParseCoinsNormalized("132500000denom1,132500000denom2,132500000denom3,132500000stake"),
				mustParseCoinsNormalized("132500000denom1,132500000denom2,132500000denom3,132500000stake"),
				mustParseCoinsNormalized("87500000denom1,87500000denom2,87500000denom3,87500000stake"),
			},
			nil,
		},
		{
			"relay-budget-fail",
			[]types.Budget{
				suite.budgets[8],
				suite.budgets[9],
				suite.budgets[10],
				suite.budgets[11],
				suite.budgets[12],
			},
			1,
			[]sdk.AccAddress{
				suite.destinationAddrs[6],
				suite.destinationAddrs[2],
				suite.destinationAddrs[3],
				suite.destinationAddrs[4],
				suite.destinationAddrs[5],
				suite.sourceAddrs[1],
			},
			[]sdk.Coins{
				{},
				mustParseCoinsNormalized("250000000denom1,250000000denom2,250000000denom3,250000000stake"),
				mustParseCoinsNormalized("397500000denom1,397500000denom2,397500000denom3,397500000stake"),
				mustParseCoinsNormalized("132500000denom1,132500000denom2,132500000denom3,132500000stake"),
				mustParseCoinsNormalized("132500000denom1,132500000denom2,132500000denom3,132500000stake"),
				mustParseCoinsNormalized("87500000denom1,87500000denom2,87500000denom3,87500000stake"),
			},
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := suite.keeper.GetParams(suite.ctx)
			params.Budgets = tc.budgets
			params.EpochBlocks = tc.epochBlocks
			suite.keeper.SetParams(suite.ctx, params)

			err := suite.keeper.CollectBudgets(suite.ctx)
			if tc.expectErr != nil {
				suite.Error(err)
			} else {
				suite.NoError(err)

				for i, acc := range tc.accAsserts {
					suite.True(suite.app.BankKeeper.GetAllBalances(suite.ctx, acc).IsEqual(tc.balanceAsserts[i]))
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBudgetChangeSituation() {
	encCfg := appparams.MakeTestEncodingConfig()
	params := suite.keeper.GetParams(suite.ctx)
	suite.keeper.SetParams(suite.ctx, params)
	height := 1
	suite.ctx = suite.ctx.WithBlockTime(types.MustParseRFC3339("2021-08-01T00:00:00Z"))
	suite.ctx = suite.ctx.WithBlockHeight(int64(height))

	// cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az
	// inflation occurs by 1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake every blocks
	budgetSource := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "InflationPool")

	for _, tc := range []struct {
		name                   string
		proposal               *proposal.ParameterChangeProposal
		budgetCount            int
		collectibleBudgetCount int
		govTime                time.Time
		nextBlockTime          time.Time
		expErr                 error
		accAsserts             []sdk.AccAddress
		balanceAsserts         []sdk.Coins
	}{
		{
			"add budget 1",
			testProposal(proposal.ParamChange{
				Subspace: types.ModuleName,
				Key:      string(types.KeyBudgets),
				Value: `[
					{
					"name": "gravity-dex-farming-1",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-12-31T00:00:00Z"
					}
				]`,
			}),
			1,
			0,
			types.MustParseRFC3339("2021-08-01T00:00:00Z"),
			types.MustParseRFC3339("2021-08-01T00:00:00Z"),
			nil,
			[]sdk.AccAddress{budgetSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
			[]sdk.Coins{
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				{},
				{},
				{},
			},
		},
		{
			"add budget 2",
			testProposal(proposal.ParamChange{
				Subspace: types.ModuleName,
				Key:      string(types.KeyBudgets),
				Value: `[
					{
					"name": "gravity-dex-farming-1",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-12-31T00:00:00Z"
					},
					{
					"name": "gravity-dex-farming-2",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-09-30T00:00:00Z"
					}
				]`,
			}),
			2,
			2,
			types.MustParseRFC3339("2021-09-03T00:00:00Z"),
			types.MustParseRFC3339("2021-09-03T00:00:00Z"),
			nil,
			[]sdk.AccAddress{budgetSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
			[]sdk.Coins{
				{},
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				{},
			},
		},
		{
			"add budget 3 overlapped with 2, invalid total rate, 1.5",
			testProposal(proposal.ParamChange{
				Subspace: types.ModuleName,
				Key:      string(types.KeyBudgets),
				Value: `[
					{
					"name": "gravity-dex-farming-1",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-12-31T00:00:00Z"
					},
					{
					"name": "gravity-dex-farming-2",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-09-30T00:00:00Z"
					},
					{
					"name": "gravity-dex-farming-3",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
					"start_time": "2021-09-29T00:00:00Z",
					"end_time": "2021-10-10T00:00:00Z"
					}
				]`,
			}),
			2, // left last budgets of 2nd tc
			1, // left last budgets of 2nd tc
			types.MustParseRFC3339("2021-09-29T00:00:00Z"),
			types.MustParseRFC3339("2021-09-30T00:00:00Z"),
			types.ErrInvalidTotalBudgetRate, // 1.5 ( 0.5 + 0.5 + 0.5 )
			[]sdk.AccAddress{budgetSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
			[]sdk.Coins{
				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
				mustParseCoinsNormalized("1500000000denom1,1500000000denom2,1500000000denom3,1500000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				{},
			},
		},
		{
			"add budget 3 with invalid total rate, 1.01",
			testProposal(proposal.ParamChange{
				Subspace: types.ModuleName,
				Key:      string(types.KeyBudgets),
				Value: `[
					{
					"name": "gravity-dex-farming-1",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-12-31T00:00:00Z"
					},
					{
					"name": "gravity-dex-farming-2",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-09-30T00:00:00Z"
					},
					{
					"name": "gravity-dex-farming-3",
					"rate": "0.510000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
					"start_time": "2021-09-30T00:00:00Z",
					"end_time": "2021-10-10T00:00:00Z"
					}
				]`,
			}),
			2, // left last budgets of 2nd tc
			1, // left last budgets of 2nd tc
			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
			types.ErrInvalidTotalBudgetRate, // 1.01 ( 0.5 + 0.0 + 0.51 )
			[]sdk.AccAddress{budgetSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
			[]sdk.Coins{
				mustParseCoinsNormalized("750000000denom1,750000000denom2,750000000denom3,750000000stake"),
				mustParseCoinsNormalized("2250000000denom1,2250000000denom2,2250000000denom3,2250000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				{},
			},
		},
		{
			"add budget 3",
			testProposal(proposal.ParamChange{
				Subspace: types.ModuleName,
				Key:      string(types.KeyBudgets),
				Value: `[
					{
					"name": "gravity-dex-farming-1",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-12-31T00:00:00Z"
					},
					{
					"name": "gravity-dex-farming-2",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-09-30T00:00:00Z"
					},
					{
					"name": "gravity-dex-farming-3",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
					"start_time": "2021-09-30T00:00:00Z",
					"end_time": "2021-10-10T00:00:00Z"
					}
				]`,
			}),
			3,
			2,
			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
			nil,
			[]sdk.AccAddress{budgetSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
			[]sdk.Coins{
				{},
				mustParseCoinsNormalized("3125000000denom1,3125000000denom2,3125000000denom3,3125000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("875000000denom1,875000000denom2,875000000denom3,875000000stake"),
			},
		},
		{
			"add budget 4 without date range overlap, remove 2, 3",
			testProposal(proposal.ParamChange{
				Subspace: types.ModuleName,
				Key:      string(types.KeyBudgets),
				Value: `[
					{
					"name": "gravity-dex-farming-1",
					"rate": "0.500000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
					"start_time": "2021-09-01T00:00:00Z",
					"end_time": "2021-12-31T00:00:00Z"
					},
					{
					"name": "gravity-dex-farming-4",
					"rate": "1.000000000000000000",
					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
					"start_time": "2021-12-31T00:00:00Z",
					"end_time": "2022-12-31T00:00:00Z"
					}
				]`,
			}),
			2,
			1,
			types.MustParseRFC3339("2021-10-02T00:00:00Z"),
			types.MustParseRFC3339("2021-10-14T00:00:00Z"),
			nil,
			[]sdk.AccAddress{budgetSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
			[]sdk.Coins{
				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
				mustParseCoinsNormalized("3625000000denom1,3625000000denom2,3625000000denom3,3625000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("875000000denom1,875000000denom2,875000000denom3,875000000stake"),
			},
		},
		{
			"remove all budgets",
			testProposal(proposal.ParamChange{
				Subspace: types.ModuleName,
				Key:      string(types.KeyBudgets),
				Value:    `[]`,
			}),
			0,
			0,
			types.MustParseRFC3339("2021-10-25T00:00:00Z"),
			types.MustParseRFC3339("2021-10-26T00:00:00Z"),
			nil,
			[]sdk.AccAddress{budgetSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
			[]sdk.Coins{
				mustParseCoinsNormalized("1500000000denom1,1500000000denom2,1500000000denom3,1500000000stake"),
				mustParseCoinsNormalized("3625000000denom1,3625000000denom2,3625000000denom3,3625000000stake"),
				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
				mustParseCoinsNormalized("875000000denom1,875000000denom2,875000000denom3,875000000stake"),
			},
		},
	} {
		suite.Run(tc.name, func() {
			proposalJson := paramscutils.ParamChangeProposalJSON{}
			bz, err := tc.proposal.Marshal()
			suite.Require().NoError(err)
			err = encCfg.Amino.Unmarshal(bz, &proposalJson)
			suite.Require().NoError(err)
			proposal := paramproposal.NewParameterChangeProposal(
				proposalJson.Title, proposalJson.Description, proposalJson.Changes.ToParamChanges(),
			)
			suite.Require().NoError(err)

			// endblock gov paramchange ->(new block)-> beginblock budget -> mempool -> endblock gov paramchange ->(new block)-> ...
			suite.ctx = suite.ctx.WithBlockTime(tc.govTime)
			err = suite.govHandler(suite.ctx, proposal)
			if tc.expErr != nil {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			// (new block)
			height += 1
			suite.ctx = suite.ctx.WithBlockHeight(int64(height))
			suite.ctx = suite.ctx.WithBlockTime(tc.nextBlockTime)

			params = suite.keeper.GetParams(suite.ctx)
			suite.Require().Len(params.Budgets, tc.budgetCount)
			for _, budget := range params.Budgets {
				err = budget.Validate()
				suite.Require().NoError(err)
			}

			budgets := types.CollectibleBudgets(params.Budgets, suite.ctx.BlockTime())
			suite.Require().Len(budgets, tc.collectibleBudgetCount)

			// BeginBlocker - inflation or mint on budgetSource
			// inflation occurs by 1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake every blocks
			err = app.FundAccount(suite.app.BankKeeper, suite.ctx, budgetSource, initialBalances)
			suite.Require().NoError(err)

			// BeginBlocker - Collect budgets
			err = suite.keeper.CollectBudgets(suite.ctx)
			suite.Require().NoError(err)

			// Assert budget collections
			for i, acc := range tc.accAsserts {
				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc)
				suite.Require().Equal(tc.balanceAsserts[i], balances)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetSetTotalCollectedCoins() {
	collectedCoins := suite.keeper.GetTotalCollectedCoins(suite.ctx, "budget1")
	suite.Require().Nil(collectedCoins)

	suite.keeper.SetTotalCollectedCoins(suite.ctx, "budget1", sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "budget1")
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)), collectedCoins))

	suite.keeper.AddTotalCollectedCoins(suite.ctx, "budget1", sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000000)))
	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "budget1")
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000), sdk.NewInt64Coin(denom2, 1000000)), collectedCoins))

	suite.keeper.AddTotalCollectedCoins(suite.ctx, "budget2", sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "budget2")
	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)), collectedCoins))
}

func (suite *KeeperTestSuite) TestTotalCollectedCoins() {
	budget := types.Budget{
		Name:               "budget1",
		Rate:               sdk.NewDecWithPrec(5, 2), // 5%
		SourceAddress:      suite.sourceAddrs[0].String(),
		DestinationAddress: suite.destinationAddrs[0].String(),
		StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
		EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
	}

	params := suite.keeper.GetParams(suite.ctx)
	params.Budgets = []types.Budget{budget}
	suite.keeper.SetParams(suite.ctx, params)

	balance := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.sourceAddrs[0])
	expectedCoins, _ := sdk.NewDecCoinsFromCoins(balance...).MulDec(sdk.NewDecWithPrec(5, 2)).TruncateDecimal()

	collectedCoins := suite.keeper.GetTotalCollectedCoins(suite.ctx, "budget1")
	suite.Require().Equal(sdk.Coins(nil), collectedCoins)

	suite.ctx = suite.ctx.WithBlockTime(types.MustParseRFC3339("2021-08-31T00:00:00Z"))
	err := suite.keeper.CollectBudgets(suite.ctx)
	suite.Require().NoError(err)

	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "budget1")
	suite.Require().True(coinsEq(expectedCoins, collectedCoins))
}
