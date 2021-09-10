package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
