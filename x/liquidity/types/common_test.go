package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

var testAddr = sdk.AccAddress(crypto.AddressHash([]byte("test")))

func newInt(i int64) sdk.Int {
	return sdk.NewInt(i)
}
