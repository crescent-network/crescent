package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (staking Staking) GetFarmerAddress() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(staking.Farmer)
	return addr
}

func (staking Staking) IdBytes() []byte {
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, staking.Id)
	return idBytes
}

func (staking Staking) Denoms() (denomList []string) {
	keys := make(map[string]bool)
	for _, coin := range staking.QueuedCoins {
		if _, value := keys[coin.Denom]; !value {
			keys[coin.Denom] = true
			denomList = append(denomList, coin.Denom)
		}
	}
	for _, coin := range staking.StakedCoins {
		if _, value := keys[coin.Denom]; !value {
			keys[coin.Denom] = true
			denomList = append(denomList, coin.Denom)
		}
	}
	return denomList
}
