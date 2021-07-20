package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func (reward Reward) GetFarmerAddress() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(reward.Farmer)
	return addr
}
