package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v4/x/marketmaker/types"
)

func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.MarketMakerKeyPrefix):
			var mA, mB types.MarketMaker
			cdc.MustUnmarshal(kvA.Value, &mA)
			cdc.MustUnmarshal(kvB.Value, &mB)
			return fmt.Sprintf("%v\n%v", mA, mB)

		case bytes.Equal(kvA.Key[:1], types.DepositKeyPrefix):
			var dA, dB types.Deposit
			cdc.MustUnmarshal(kvA.Value, &dA)
			cdc.MustUnmarshal(kvB.Value, &dB)
			return fmt.Sprintf("%v\n%v", dA, dB)

		case bytes.Equal(kvA.Key[:1], types.IncentiveKeyPrefix):
			var iA, iB types.Incentive
			cdc.MustUnmarshal(kvA.Value, &iA)
			cdc.MustUnmarshal(kvB.Value, &iB)
			return fmt.Sprintf("%v\n%v", iA, iB)

		default:
			panic(fmt.Sprintf("invalid marketmaker key prefix %X", kvA.Key[:1]))
		}
	}
}
