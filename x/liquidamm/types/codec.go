package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterLegacyAminoCodec registers the necessary x/liquidamm interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgMintShare{}, "liquidamm/MsgMintShare", nil)
	cdc.RegisterConcrete(&MsgBurnShare{}, "liquidamm/MsgBurnShare", nil)
	cdc.RegisterConcrete(&MsgPlaceBid{}, "liquidamm/MsgPlaceBid", nil)
	cdc.RegisterConcrete(&PublicPositionCreateProposal{}, "liquidamm/PublicPositionCreateProposal", nil)
	cdc.RegisterConcrete(&PublicPositionParameterChangeProposal{}, "liquidamm/PublicPositionParameterChangeProposal", nil)
}

// RegisterInterfaces registers the x/liquidamm interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgMintShare{},
		&MsgBurnShare{},
		&MsgPlaceBid{},
	)
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&PublicPositionCreateProposal{},
		&PublicPositionParameterChangeProposal{},
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
