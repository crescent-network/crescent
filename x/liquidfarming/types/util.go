package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// DeriveBidReserveAddress creates the reserve address for bids
// with the given liquid farm id.
func DeriveBidReserveAddress(liquidFarmId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("BidReserveAddress/%d", liquidFarmId)))
}
