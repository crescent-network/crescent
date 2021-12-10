package keeper_test

//func (suite *KeeperTestSuite) TestCollectBiquidStakings() {
//	for _, tc := range []struct {
//		name           string
//		liquidStakings       []types.BiquidStaking
//		epochBlocks    uint32
//		accAsserts     []sdk.AccAddress
//		balanceAsserts []sdk.Coins
//		expectErr      bool
//	}{
//		{
//			"basic liquidStakings case",
//			suite.liquidStakings[:4],
//			types.DefaultEpochBlocks,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.destinationAddrs[2],
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[0],
//				suite.sourceAddrs[1],
//				suite.sourceAddrs[2],
//			},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
//				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//				{},
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//			},
//			false,
//		},
//		{
//			"only expired liquidstaking case",
//			[]types.BiquidStaking{suite.liquidStakings[3]},
//			types.DefaultEpochBlocks,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[2],
//			},
//			[]sdk.Coins{
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//			},
//			false,
//		},
//		{
//			"source has small balances case",
//			suite.liquidStakings[4:6],
//			types.DefaultEpochBlocks,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.sourceAddrs[3],
//			},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("1denom2,1denom3,500000000stake"),
//				mustParseCoinsNormalized("1denom2,1denom3,500000000stake"),
//				mustParseCoinsNormalized("1denom1,1denom3"),
//			},
//			false,
//		},
//		{
//			"none liquidStakings case",
//			nil,
//			types.DefaultEpochBlocks,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.destinationAddrs[2],
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[0],
//				suite.sourceAddrs[1],
//				suite.sourceAddrs[2],
//				suite.sourceAddrs[3],
//			},
//			[]sdk.Coins{
//				{},
//				{},
//				{},
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake"),
//			},
//			false,
//		},
//		{
//			"disabled liquidstaking epoch",
//			nil,
//			0,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.destinationAddrs[2],
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[0],
//				suite.sourceAddrs[1],
//				suite.sourceAddrs[2],
//				suite.sourceAddrs[3],
//			},
//			[]sdk.Coins{
//				{},
//				{},
//				{},
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake"),
//			},
//			false,
//		},
//		{
//			"disabled liquidstaking epoch with liquidStakings",
//			suite.liquidStakings[:4],
//			0,
//			[]sdk.AccAddress{
//				suite.destinationAddrs[0],
//				suite.destinationAddrs[1],
//				suite.destinationAddrs[2],
//				suite.destinationAddrs[3],
//				suite.sourceAddrs[0],
//				suite.sourceAddrs[1],
//				suite.sourceAddrs[2],
//				suite.sourceAddrs[3],
//			},
//			[]sdk.Coins{
//				{},
//				{},
//				{},
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake"),
//			},
//			false,
//		},
//	} {
//		suite.Run(tc.name, func() {
//			suite.SetupTest()
//			params := suite.keeper.GetParams(suite.ctx)
//			params.BiquidStakings = tc.liquidStakings
//			params.EpochBlocks = tc.epochBlocks
//			suite.keeper.SetParams(suite.ctx, params)
//
//			err := suite.keeper.CollectBiquidStakings(suite.ctx)
//			if tc.expectErr {
//				suite.Error(err)
//			} else {
//				suite.NoError(err)
//
//				for i, acc := range tc.accAsserts {
//					suite.True(suite.app.BankKeeper.GetAllBalances(suite.ctx, acc).IsEqual(tc.balanceAsserts[i]))
//				}
//			}
//		})
//	}
//}
//
//func (suite *KeeperTestSuite) TestBiquidStakingChangeSituation() {
//	encCfg := app.MakeTestEncodingConfig()
//	params := suite.keeper.GetParams(suite.ctx)
//	suite.keeper.SetParams(suite.ctx, params)
//	height := 1
//	suite.ctx = suite.ctx.WithBlockTime(types.MustParseRFC3339("2021-08-01T00:00:00Z"))
//	suite.ctx = suite.ctx.WithBlockHeight(int64(height))
//
//	// cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az
//	// inflation occurs by 1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake every blocks
//	liquidStakingSource := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "InflationPool")
//
//	for _, tc := range []struct {
//		name                    string
//		proposal                *proposal.ParameterChangeProposal
//		liquidStakingCount            int
//		collectibleBiquidStakingCount int
//		govTime                 time.Time
//		nextBlockTime           time.Time
//		expErr                  error
//		accAsserts              []sdk.AccAddress
//		balanceAsserts          []sdk.Coins
//	}{
//		{
//			"add liquidstaking 1",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					}
//				]`,
//			}),
//			1,
//			0,
//			types.MustParseRFC3339("2021-08-01T00:00:00Z"),
//			types.MustParseRFC3339("2021-08-01T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//				{},
//				{},
//			},
//		},
//		{
//			"add liquidstaking 2",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-2",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2021-09-30T00:00:00Z"
//					}
//				]`,
//			}),
//			2,
//			2,
//			types.MustParseRFC3339("2021-09-03T00:00:00Z"),
//			types.MustParseRFC3339("2021-09-03T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				{},
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//			},
//		},
//		{
//			"add liquidstaking 3 with invalid total rate case 1",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-2",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2021-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-3",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
//					"start_time": "2021-09-30T00:00:00Z",
//					"end_time": "2021-10-10T00:00:00Z"
//					}
//				]`,
//			}),
//			2, // left last liquidStakings of 2nd tc
//			1, // left last liquidStakings of 2nd tc
//			types.MustParseRFC3339("2021-09-29T00:00:00Z"),
//			types.MustParseRFC3339("2021-09-30T00:00:00Z"),
//			types.ErrInvalidTotalBiquidStakingRate,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
//				mustParseCoinsNormalized("1500000000denom1,1500000000denom2,1500000000denom3,1500000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//			},
//		},
//		{
//			"add liquidstaking 3 with invalid total rate case 2",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-2",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1czyx0dj2yd26gv3stpxzv23ddy8pld4j6p90a683mdcg8vzy72jqa8tm6p",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2021-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-3",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
//					"start_time": "2021-09-30T00:00:00Z",
//					"end_time": "2021-10-10T00:00:00Z"
//					}
//				]`,
//			}),
//			2, // left last liquidStakings of 2nd tc
//			1, // left last liquidStakings of 2nd tc
//			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
//			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
//			types.ErrInvalidTotalBiquidStakingRate,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("750000000denom1,750000000denom2,750000000denom3,750000000stake"),
//				mustParseCoinsNormalized("2250000000denom1,2250000000denom2,2250000000denom3,2250000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				{},
//			},
//		},
//		{
//			"add liquidstaking 3",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-3",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
//					"start_time": "2021-09-30T00:00:00Z",
//					"end_time": "2021-10-10T00:00:00Z"
//					}
//				]`,
//			}),
//			2,
//			2,
//			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
//			types.MustParseRFC3339("2021-10-01T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				{},
//				mustParseCoinsNormalized("3125000000denom1,3125000000denom2,3125000000denom3,3125000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("875000000denom1,875000000denom2,875000000denom3,875000000stake"),
//			},
//		},
//		{
//			"add liquidstaking 4 without date range overlap",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value: `[
//					{
//					"name": "gravity-dex-farming-1",
//					"rate": "0.500000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1qceyjmnrl6hapntjq3z25vn38nh68u7yxvufs2thptxvqm7huxeqj7zyrq",
//					"start_time": "2021-09-01T00:00:00Z",
//					"end_time": "2031-09-30T00:00:00Z"
//					},
//					{
//					"name": "gravity-dex-farming-4",
//					"rate": "1.000000000000000000",
//					"source_address": "cosmos10wy60v3zuks7rkwnqxs3e878zqfhus6m98l77q6rppz40kxwgllsruc0az",
//					"destination_address": "cosmos1e0n8jmeg4u8q3es2tmhz5zlte8a4q8687ndns8pj4q8grdl74a0sw3045s",
//					"start_time": "2031-09-30T00:00:01Z",
//					"end_time": "2031-12-10T00:00:00Z"
//					}
//				]`,
//			}),
//			2,
//			1,
//			types.MustParseRFC3339("2021-09-29T00:00:00Z"),
//			types.MustParseRFC3339("2021-09-30T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("500000000denom1,500000000denom2,500000000denom3,500000000stake"),
//				mustParseCoinsNormalized("3625000000denom1,3625000000denom2,3625000000denom3,3625000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("875000000denom1,875000000denom2,875000000denom3,875000000stake"),
//			},
//		},
//		{
//			"remove all liquidStakings",
//			testProposal(proposal.ParamChange{
//				Subspace: types.ModuleName,
//				Key:      string(types.KeyBiquidStakings),
//				Value:    `[]`,
//			}),
//			0,
//			0,
//			types.MustParseRFC3339("2021-10-25T00:00:00Z"),
//			types.MustParseRFC3339("2021-10-26T00:00:00Z"),
//			nil,
//			[]sdk.AccAddress{liquidStakingSource, suite.destinationAddrs[0], suite.destinationAddrs[1], suite.destinationAddrs[2]},
//			[]sdk.Coins{
//				mustParseCoinsNormalized("1500000000denom1,1500000000denom2,1500000000denom3,1500000000stake"),
//				mustParseCoinsNormalized("3625000000denom1,3625000000denom2,3625000000denom3,3625000000stake"),
//				mustParseCoinsNormalized("1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake"),
//				mustParseCoinsNormalized("875000000denom1,875000000denom2,875000000denom3,875000000stake"),
//			},
//		},
//	} {
//		suite.Run(tc.name, func() {
//			proposalJson := paramscutils.ParamChangeProposalJSON{}
//			bz, err := tc.proposal.Marshal()
//			suite.Require().NoError(err)
//			err = encCfg.Amino.Unmarshal(bz, &proposalJson)
//			suite.Require().NoError(err)
//			proposal := paramproposal.NewParameterChangeProposal(
//				proposalJson.Title, proposalJson.Description, proposalJson.Changes.ToParamChanges(),
//			)
//			suite.Require().NoError(err)
//
//			// endblock gov paramchange ->(new block)-> beginblock liquidstaking -> mempool -> endblock gov paramchange ->(new block)-> ...
//			suite.ctx = suite.ctx.WithBlockTime(tc.govTime)
//			err = suite.govHandler(suite.ctx, proposal)
//			if tc.expErr != nil {
//				suite.Require().Error(err)
//			} else {
//				suite.Require().NoError(err)
//			}
//
//			// (new block)
//			height += 1
//			suite.ctx = suite.ctx.WithBlockHeight(int64(height))
//			suite.ctx = suite.ctx.WithBlockTime(tc.nextBlockTime)
//
//			params := suite.keeper.GetParams(suite.ctx)
//			suite.Require().Len(params.BiquidStakings, tc.liquidStakingCount)
//			for _, liquidStaking := range params.BiquidStakings {
//				err := liquidStaking.Validate()
//				suite.Require().NoError(err)
//			}
//
//			liquidStakings := types.CollectibleBiquidStakings(params.BiquidStakings, suite.ctx.BlockTime())
//			suite.Require().Len(liquidStakings, tc.collectibleBiquidStakingCount)
//
//			// BeginBlocker - inflation or mint on liquidStakingSource
//			// inflation occurs by 1000000000denom1,1000000000denom2,1000000000denom3,1000000000stake every blocks
//			err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, liquidStakingSource, initialBalances)
//			suite.Require().NoError(err)
//
//			// BeginBlocker - Collect liquidStakings
//			err = suite.keeper.CollectBiquidStakings(suite.ctx)
//			suite.Require().NoError(err)
//
//			// Assert liquidstaking collections
//			for i, acc := range tc.accAsserts {
//				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, acc)
//				suite.Require().Equal(tc.balanceAsserts[i], balances)
//			}
//		})
//	}
//}
//
//func (suite *KeeperTestSuite) TestGetSetTotalCollectedCoins() {
//	collectedCoins := suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().Nil(collectedCoins)
//
//	suite.keeper.SetTotalCollectedCoins(suite.ctx, "liquidStaking1", sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
//	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)), collectedCoins))
//
//	suite.keeper.AddTotalCollectedCoins(suite.ctx, "liquidStaking1", sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000000)))
//	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000), sdk.NewInt64Coin(denom2, 1000000)), collectedCoins))
//
//	suite.keeper.AddTotalCollectedCoins(suite.ctx, "liquidStaking2", sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)))
//	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking2")
//	suite.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000000)), collectedCoins))
//}
//
//func (suite *KeeperTestSuite) TestTotalCollectedCoins() {
//	liquidStaking := types.BiquidStaking{
//		Name:               "liquidStaking1",
//		Rate:               sdk.NewDecWithPrec(5, 2), // 5%
//		SourceAddress:      suite.sourceAddrs[0].String(),
//		DestinationAddress: suite.destinationAddrs[0].String(),
//		StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
//		EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
//	}
//
//	params := suite.keeper.GetParams(suite.ctx)
//	params.BiquidStakings = []types.BiquidStaking{liquidStaking}
//	suite.keeper.SetParams(suite.ctx, params)
//
//	balance := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.sourceAddrs[0])
//	expectedCoins, _ := sdk.NewDecCoinsFromCoins(balance...).MulDec(sdk.NewDecWithPrec(5, 2)).TruncateDecimal()
//
//	collectedCoins := suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().Equal(sdk.Coins(nil), collectedCoins)
//
//	suite.ctx = suite.ctx.WithBlockTime(types.MustParseRFC3339("2021-08-31T00:00:00Z"))
//	err := suite.keeper.CollectBiquidStakings(suite.ctx)
//	suite.Require().NoError(err)
//
//	collectedCoins = suite.keeper.GetTotalCollectedCoins(suite.ctx, "liquidStaking1")
//	suite.Require().True(coinsEq(expectedCoins, collectedCoins))
//}
