package app

// DONTCOVER

import (
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/std"

	"github.com/crescent-network/crescent/v5/app/params"

	enccodec "github.com/evmos/ethermint/encoding/codec"
	ethermint "github.com/evmos/ethermint/types"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeTestEncodingConfig()
	//std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// ethermint
	cryptocodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ethermint.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	enccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)

	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
