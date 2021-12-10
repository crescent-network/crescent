package types_test

//func TestValidateGenesis(t *testing.T) {
//	startTime, _ := time.Parse(time.RFC3339, "0000-01-01T00:00:00Z")
//	endTime, _ := time.Parse(time.RFC3339, "9999-12-31T00:00:00Z")
//	testCases := []struct {
//		name        string
//		configure   func(*types.GenesisState)
//		expectedErr string
//	}{
//		{
//			"default case",
//			func(genState *types.GenesisState) {
//				genState.Params = types.DefaultParams()
//			},
//			"",
//		},
//		{
//			"normal liquidstaking case",
//			func(genState *types.GenesisState) {
//				genState.Params = types.DefaultParams()
//				genState.Params.BiquidStakings = []types.BiquidStaking{
//					{
//						Name:               "liquidStaking1",
//						Rate:               sdk.NewDecWithPrec(5, 2), // 5%
//						SourceAddress:      sdk.AccAddress(crypto.AddressHash([]byte("SourceAddress"))).String(),
//						DestinationAddress: sdk.AccAddress(crypto.AddressHash([]byte("DestinationAddress"))).String(),
//						StartTime:          startTime,
//						EndTime:            endTime,
//					},
//				}
//			},
//			"",
//		},
//		{
//			"invalid liquidstaking case",
//			func(genState *types.GenesisState) {
//				genState.Params = types.DefaultParams()
//				genState.Params.BiquidStakings = []types.BiquidStaking{
//					{
//						Name:               "liquidStaking1",
//						Rate:               sdk.NewDecWithPrec(5, 2), // 5%
//						SourceAddress:      "cosmos1invalidaddress",
//						DestinationAddress: sdk.AccAddress(crypto.AddressHash([]byte("DestinationAddress"))).String(),
//						StartTime:          startTime,
//						EndTime:            endTime,
//					},
//				}
//			},
//			"invalid source address cosmos1invalidaddress: decoding bech32 failed: failed converting data to bytes: invalid character not part of charset: 105: invalid address",
//		},
//		{
//			"duplicate liquidstaking name",
//			func(genState *types.GenesisState) {
//				genState.Params = types.DefaultParams()
//				genState.Params.BiquidStakings = []types.BiquidStaking{
//					{
//						Name:               "liquidStaking1",
//						Rate:               sdk.NewDecWithPrec(5, 2), // 5%
//						SourceAddress:      sdk.AccAddress(crypto.AddressHash([]byte("SourceAddress"))).String(),
//						DestinationAddress: sdk.AccAddress(crypto.AddressHash([]byte("DestinationAddress"))).String(),
//						StartTime:          startTime,
//						EndTime:            endTime,
//					},
//					{
//						Name:               "liquidStaking1",
//						Rate:               sdk.NewDecWithPrec(5, 2), // 5%
//						SourceAddress:      sdk.AccAddress(crypto.AddressHash([]byte("SourceAddress"))).String(),
//						DestinationAddress: sdk.AccAddress(crypto.AddressHash([]byte("DestinationAddress"))).String(),
//						StartTime:          startTime,
//						EndTime:            endTime,
//					},
//				}
//			},
//			"liquidStaking1: duplicate liquidstaking name",
//		},
//		{
//			"invalid liquidstaking name case",
//			func(genState *types.GenesisState) {
//				genState.Params = types.DefaultParams()
//				genState.BiquidStakingRecords = []types.BiquidStakingRecord{
//					{
//						Name:                "invalid name",
//						TotalCollectedCoins: nil,
//					},
//				}
//			},
//			"invalid name: liquidstaking name only allows letters, digits, and dash(-) without spaces and the maximum length is 50",
//		},
//		{
//			"invalid total_collected_coin case",
//			func(genState *types.GenesisState) {
//				genState.Params = types.DefaultParams()
//				genState.BiquidStakingRecords = []types.BiquidStakingRecord{
//					{
//						Name:                "liquidStaking1",
//						TotalCollectedCoins: sdk.Coins{sdk.NewCoin("stake", sdk.ZeroInt())},
//					},
//				}
//			},
//			"invalid total collected coins 0stake: coin 0stake amount is not positive: invalid coins",
//		},
//	}
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			genState := types.DefaultGenesisState()
//			tc.configure(genState)
//
//			err := types.ValidateGenesis(*genState)
//			if tc.expectedErr == "" {
//				require.Nil(t, err)
//			} else {
//				require.EqualError(t, err, tc.expectedErr)
//			}
//		})
//	}
//}
