package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterLegacyAminoCodec registers the necessary x/liquidfarming interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgMintShare{}, "liquidfarming/MsgMintShare", nil)
	cdc.RegisterConcrete(&MsgBurnShare{}, "liquidfarming/MsgBurnShare", nil)
	cdc.RegisterConcrete(&MsgPlaceBid{}, "liquidfarming/MsgPlaceBid", nil)
	cdc.RegisterConcrete(&LiquidFarmCreateProposal{}, "liquidfarming/LiquidFarmCreateProposal", nil)
	cdc.RegisterConcrete(&LiquidFarmParameterChangeProposal{}, "liquidfarming/LiquidFarmParameterChangeProposal", nil)
}

// RegisterInterfaces registers the x/liquidfarming interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgMintShare{},
		&MsgBurnShare{},
		&MsgPlaceBid{},
	)
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&LiquidFarmCreateProposal{},
		&LiquidFarmParameterChangeProposal{},
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
