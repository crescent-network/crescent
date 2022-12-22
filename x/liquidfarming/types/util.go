package types

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingtypes "github.com/crescent-network/crescent/v4/x/farming/types"
)

const (
	PayingReserveAddressPrefix           string = "PayingReserveAddress"
	WithdrawnRewardsReserveAddressPrefix string = "WithdrawnRewardsReserveAddress"
	ModuleAddressNameSplitter            string = "|"

	// The module uses the address type of 32 bytes length, but it can always be changed depending on Cosmos SDK's direction.
	ReserveAddressType = farmingtypes.AddressType32Bytes
)

// PayingReserveAddress creates the paying reserve address in the form of sdk.AccAddress
// with the given pool id.
func PayingReserveAddress(poolId uint64) sdk.AccAddress {
	return farmingtypes.DeriveAddress(
		ReserveAddressType,
		ModuleName,
		strings.Join([]string{PayingReserveAddressPrefix, strconv.FormatUint(poolId, 10)}, ModuleAddressNameSplitter),
	)
}

// WithdrawnRewardsReserveAddress creates the withdrawn rewards reserve address in the form of sdk.AccAddress
// with the given pool id.
func WithdrawnRewardsReserveAddress(poolId uint64) sdk.AccAddress {
	return farmingtypes.DeriveAddress(
		ReserveAddressType,
		ModuleName,
		strings.Join([]string{WithdrawnRewardsReserveAddressPrefix, strconv.FormatUint(poolId, 10)}, ModuleAddressNameSplitter),
	)
}
