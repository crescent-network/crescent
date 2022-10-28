package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterLegacyAminoCodec registers the necessary x/lpfarm interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreatePrivatePlan{}, "lpfarm/MsgCreatePrivatePlan", nil)
	cdc.RegisterConcrete(&MsgFarm{}, "lpfarm/MsgFarm", nil)
	cdc.RegisterConcrete(&MsgUnfarm{}, "lpfarm/MsgUnfarm", nil)
	cdc.RegisterConcrete(&MsgHarvest{}, "lpfarm/MsgHarvest", nil)
	cdc.RegisterConcrete(&FarmingPlanProposal{}, "lpfarm/FarmingPlanProposal", nil)
}

// RegisterInterfaces registers the x/lpfarm interfaces types with the
// interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreatePrivatePlan{},
		&MsgFarm{},
		&MsgUnfarm{},
		&MsgHarvest{},
	)
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&FarmingPlanProposal{},
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
