package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers the necessary x/liquidstaking interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLiquidStake{}, "liquidstaking/MsgLiquidStake", nil)
	cdc.RegisterConcrete(&MsgLiquidUnstake{}, "liquidstaking/MsgLiquidUnstake", nil)
}

// RegisterInterfaces registers the x/liquidstaking interfaces types with the interface registry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgLiquidStake{},
		&MsgLiquidUnstake{},
	)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/liquidstaking module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding as Amino
	// is still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/liquidstaking and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
