package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (r Reward) String() string {
	out, _ := r.MarshalYAML()
	return out.(string)
}

func (r Reward) MarshalYAML() (interface{}, error) {
	bz, err := codec.MarshalYAML(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), &r)
	if err != nil {
		return nil, err
	}
	return string(bz), err
}

func (r Reward) GetFarmer() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(r.Farmer)
	return addr
}

func (r RewardCoins) String() string {
	out, _ := r.MarshalYAML()
	return out.(string)
}

func (r RewardCoins) MarshalYAML() (interface{}, error) {
	bz, err := codec.MarshalYAML(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), &r)
	if err != nil {
		return nil, err
	}
	return string(bz), err
}
