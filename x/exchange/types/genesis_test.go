package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestNumMMOrdersRecord_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(record *types.NumMMOrdersRecord)
		expectedErr string
	}{
		{
			"happy case",
			func(record *types.NumMMOrdersRecord) {
			},
			"",
		},
		{
			"invalid orderer",
			func(record *types.NumMMOrdersRecord) {
				record.Orderer = "invalid"
			},
			"invalid orderer: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"invalid market id",
			func(record *types.NumMMOrdersRecord) {
				record.MarketId = 0
			},
			"market id must not be 0",
		},
		{
			"invalid num mm orders",
			func(record *types.NumMMOrdersRecord) {
				record.NumMMOrders = 0
			},
			"num mm orders must not be 0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := types.NumMMOrdersRecord{
				Orderer:     utils.TestAddress(1).String(),
				MarketId:    1,
				NumMMOrders: 10,
			}
			tc.malleate(&record)
			err := record.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
