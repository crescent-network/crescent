package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/liquidfarming interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgFarm{}, "liquidfarming/MsgFarm", nil)
	cdc.RegisterConcrete(&MsgUnfarm{}, "liquidfarming/MsgUnfarm", nil)
	cdc.RegisterConcrete(&MsgUnfarmAndWithdraw{}, "liquidfarming/MsgUnfarmAndWithdraw", nil)
	cdc.RegisterConcrete(&MsgPlaceBid{}, "liquidfarming/MsgPlaceBid", nil)
	cdc.RegisterConcrete(&MsgRefundBid{}, "liquidfarming/MsgRefundBid", nil)
}

// RegisterInterfaces registers the x/liquidfarming interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFarm{},
		&MsgUnfarm{},
		&MsgUnfarmAndWithdraw{},
		&MsgPlaceBid{},
		&MsgRefundBid{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
