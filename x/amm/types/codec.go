package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/amm interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreatePool{}, "amm/MsgCreatePool", nil)
	cdc.RegisterConcrete(&MsgAddLiquidity{}, "amm/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(&MsgRemoveLiquidity{}, "amm/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(&MsgCollect{}, "amm/MsgCollect", nil)
	cdc.RegisterConcrete(&MsgCreatePrivateFarmingPlan{}, "amm/MsgCreatePrivateFarmingPlan", nil)
	cdc.RegisterConcrete(&MsgHarvest{}, "amm/MsgHarvest", nil)
}

// RegisterInterfaces registers the x/amm interfaces types with the
// interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreatePool{},
		&MsgAddLiquidity{},
		&MsgRemoveLiquidity{},
		&MsgCollect{},
		&MsgCreatePrivateFarmingPlan{},
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
