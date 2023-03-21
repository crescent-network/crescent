package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/exchange interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateSpotMarket{}, "exchange/MsgCreateSpotMarket", nil)
	cdc.RegisterConcrete(&MsgPlaceSpotLimitOrder{}, "exchange/MsgPlaceSpotLimitOrder", nil)
	cdc.RegisterConcrete(&MsgPlaceSpotMarketOrder{}, "exchange/MsgPlaceSpotMarketOrder", nil)
}

// RegisterInterfaces registers the x/exchange interfaces types with the
// interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateSpotMarket{},
		&MsgPlaceSpotLimitOrder{},
		&MsgPlaceSpotMarketOrder{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
