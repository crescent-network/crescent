package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v2/x/mint/types"
)

//NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
//Value to the corresponding mint type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key, types.LastBlockTimeKey):
			ts1, err := sdk.ParseTimeBytes(kvA.Value)
			if err != nil {
				panic(err)
			}
			ts2, err := sdk.ParseTimeBytes(kvA.Value)
			if err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", ts1, ts2)
		default:
			panic(fmt.Sprintf("invalid last block time key %X", kvA.Key))
		}
	}
}
