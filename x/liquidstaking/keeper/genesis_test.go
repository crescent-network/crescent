package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

func (s *KeeperTestSuite) TestInitGenesis() {
	genState := *types.DefaultGenesisState()
	s.keeper.InitGenesis(s.ctx, genState)
	got := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genState, *got)
}

func (s *KeeperTestSuite) TestImportExportGenesis() {
	k, ctx := s.keeper, s.ctx
	_, valOpers, _ := s.CreateValidators([]int64{1000000, 1000000, 1000000})
	params := k.GetParams(ctx)

	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
	}
	k.SetParams(ctx, params)
	k.UpdateLiquidValidatorSet(ctx)

	stakingAmt := sdk.NewInt(100000000)
	s.Require().NoError(s.liquidStaking(s.delAddrs[0], stakingAmt))
	lvs := k.GetAllLiquidValidators(ctx)
	s.Require().Len(lvs, 2)

	lvStates := k.GetAllLiquidValidatorStates(ctx)
	genState := k.ExportGenesis(ctx)

	bz := s.app.AppCodec().MustMarshalJSON(genState)

	var genState2 types.GenesisState
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	k.InitGenesis(ctx, genState2)
	genState3 := k.ExportGenesis(ctx)

	s.Require().Equal(*genState, genState2)
	s.Require().Equal(genState2, *genState3)

	lvs = k.GetAllLiquidValidators(ctx)
	s.Require().Len(lvs, 2)

	lvStates3 := k.GetAllLiquidValidatorStates(ctx)
	s.Require().EqualValues(lvStates, lvStates3)
}

func (s *KeeperTestSuite) TestImportExportGenesisEmpty() {
	k, ctx := s.keeper, s.ctx
	k.SetParams(ctx, types.DefaultParams())
	k.UpdateLiquidValidatorSet(ctx)
	genState := k.ExportGenesis(ctx)

	var genState2 types.GenesisState
	bz := s.app.AppCodec().MustMarshalJSON(genState)
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	k.InitGenesis(ctx, genState2)

	genState3 := k.ExportGenesis(ctx)
	s.Require().Equal(*genState, genState2)
	s.Require().Equal(genState2, *genState3)
}
