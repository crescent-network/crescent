package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterLegacyAminoCodec registers the necessary x/exchange interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateMarket{}, "exchange/MsgCreateMarket", nil)
	cdc.RegisterConcrete(&MsgPlaceLimitOrder{}, "exchange/MsgPlaceLimitOrder", nil)
	cdc.RegisterConcrete(&MsgPlaceMarketOrder{}, "exchange/MsgPlaceMarketOrder", nil)
	cdc.RegisterConcrete(&MsgPlaceMMLimitOrder{}, "exchange/MsgPlaceMMLimitOrder", nil)
	cdc.RegisterConcrete(&MsgCancelOrder{}, "exchange/MsgCancelOrder", nil)
	cdc.RegisterConcrete(&MsgSwapExactAmountIn{}, "exchange/MsgSwapExactAmountIn", nil)
	cdc.RegisterConcrete(&MarketParameterChangeProposal{}, "exchange/MarketParameterChangeProposal", nil)
}

// RegisterInterfaces registers the x/exchange interfaces types with the
// interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateMarket{},
		&MsgPlaceLimitOrder{},
		&MsgPlaceMMLimitOrder{},
		&MsgPlaceMarketOrder{},
		&MsgCancelOrder{},
		&MsgSwapExactAmountIn{},
	)
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&MarketParameterChangeProposal{},
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
