package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (record ClaimRecord) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(record.Address)
	if err != nil {
		panic(err)
	}
	return addr
}
