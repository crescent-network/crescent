package simulation_test

//func TestDecodeBiquidStakingStore(t *testing.T) {
//
//	cdc := simapp.MakeTestEncodingConfig()
//	dec := simulation.NewDecodeStore(cdc.Marshaler)
//
//	tc := types.TotalCollectedCoins{
//		TotalCollectedCoins: sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(1000000))),
//	}
//
//	kvPairs := kv.Pairs{
//		Pairs: []kv.Pair{
//			{Key: types.TotalCollectedCoinsKeyPrefix, Value: cdc.Marshaler.MustMarshal(&tc)},
//			{Key: []byte{0x99}, Value: []byte{0x99}},
//		},
//	}
//
//	tests := []struct {
//		name        string
//		expectedLog string
//	}{
//		{"totalCollectedCoins", fmt.Sprintf("%v\n%v", tc, tc)},
//		{"other", ""},
//	}
//	for i, tt := range tests {
//		i, tt := i, tt
//		t.Run(tt.name, func(t *testing.T) {
//			switch i {
//			case len(tests) - 1:
//				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
//			default:
//				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
//			}
//		})
//	}
//}
