package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s Staking) String() string {
	out, _ := s.MarshalYAML()
	return out.(string)
}

func (s Staking) MarshalYAML() (interface{}, error) {
	bz, err := codec.MarshalYAML(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), &s)
	if err != nil {
		return nil, err
	}
	return string(bz), err
}

func (s Staking) GetFarmer() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(s.Farmer)
	return addr
}

func (s Staking) StakingCoinDenoms() (denoms []string) {
	denomSet := make(map[string]struct{})
	for _, coin := range append(s.StakedCoins, s.QueuedCoins...) {
		if _, ok := denomSet[coin.Denom]; !ok {
			denomSet[coin.Denom] = struct{}{}
			denoms = append(denoms, coin.Denom)
		}
	}
	return
}
