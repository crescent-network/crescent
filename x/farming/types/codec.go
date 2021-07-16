package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterLegacyAminoCodec registers the necessary x/farming interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateFixedAmountPlan{}, "cosmos-sdk/MsgCreateFixedAmountPlan", nil)
	cdc.RegisterConcrete(&MsgCreateRatioPlan{}, "cosmos-sdk/MsgCreateRatioPlan", nil)
	cdc.RegisterConcrete(&MsgStake{}, "cosmos-sdk/MsgStake", nil)
	cdc.RegisterConcrete(&MsgUnstake{}, "cosmos-sdk/MsgUnstake", nil)
	cdc.RegisterConcrete(&MsgHarvest{}, "cosmos-sdk/MsgHarvest", nil)
}

// RegisterInterfaces registers the x/farming interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateFixedAmountPlan{},
		&MsgCreateRatioPlan{},
		&MsgStake{},
		&MsgUnstake{},
		&MsgHarvest{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&PublicPlanProposal{},
	)

	registry.RegisterInterface(
		"cosmos.farming.v1beta1.PlanI",
		(*PlanI)(nil),
		&FixedAmountPlan{},
		&RatioPlan{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/farming module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding as Amino
	// is still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/farming and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
