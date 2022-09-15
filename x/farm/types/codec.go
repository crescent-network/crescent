package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/farm interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreatePrivatePlan{}, "farm/MsgCreatePrivatePlan", nil)
	cdc.RegisterConcrete(&MsgFarm{}, "farm/MsgFarm", nil)
	cdc.RegisterConcrete(&MsgUnfarm{}, "farm/MsgUnfarm", nil)
	cdc.RegisterConcrete(&MsgHarvest{}, "farm/MsgHarvest", nil)
}

// RegisterInterfaces registers the x/farm interfaces types with the
// interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreatePrivatePlan{},
		&MsgFarm{},
		&MsgUnfarm{},
		&MsgHarvest{},
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
