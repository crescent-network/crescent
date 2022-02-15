package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (r ClaimRecord) GetRecipient() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.Recipient)
	if err != nil {
		panic(err)
	}
	return addr
}
