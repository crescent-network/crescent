package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	farmingtypes "github.com/cosmosquad-labs/squad/x/farming/types"
)

var (
	SourceAddressPrefix       string = "SourceAddress"
	ModuleAddressNameSplitter string = "|"
)

func (a Airdrop) GetSourceAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(a.SourceAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (a Airdrop) GetTerminationAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(a.TerminationAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// SourceAddress returns an account for the source reserve address with the given airdrop id.
func SourceAddress(airdropId uint64) sdk.AccAddress {
	return farmingtypes.DeriveAddress(farmingtypes.ReserveAddressType, ModuleName, SourceAddressPrefix+ModuleAddressNameSplitter+fmt.Sprint(airdropId))
}
