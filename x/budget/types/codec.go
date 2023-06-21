package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterLegacyAminoCodec registers the necessary x/budget interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
}

// RegisterInterfaces registers the x/budget interfaces types with the interface registry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/budget module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding as Amino
	// is still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/budget and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)
