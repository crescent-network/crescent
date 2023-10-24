package keeper_test

import (
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func (s *KeeperTestSuite) TestProposalHandler() {
	handler := liquidamm.NewProposalHandler(s.keeper)

	market := s.CreateMarket("ucre", "uusd")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))

	ctx := s.Ctx

	// Pool not found
	createProposal := types.NewPublicPositionCreateProposal(
		"Title", "Description", 3,
		utils.ParseDec("4.9"), utils.ParseDec("5.1"), utils.ParseDec("0.003"))
	s.Require().NoError(createProposal.ValidateBasic())
	s.Ctx, _ = ctx.CacheContext()
	s.Require().EqualError(handler(s.Ctx, createProposal), "pool not found: not found")

	// Not respecting tick spacing
	createProposal = types.NewPublicPositionCreateProposal(
		"Title", "Description", pool.Id,
		utils.ParseDec("4.9999"), utils.ParseDec("5.0001"), utils.ParseDec("0.003"))
	s.Ctx, _ = ctx.CacheContext()
	s.Require().EqualError(
		handler(s.Ctx, createProposal), "lower tick 39999 must be multiple of tick spacing 50: invalid request")

	// Successfully created a public position.
	createProposal = types.NewPublicPositionCreateProposal(
		"Title", "Description", pool.Id,
		utils.ParseDec("4.9"), utils.ParseDec("5.1"), utils.ParseDec("0.003"))
	s.Ctx = ctx
	s.Require().NoError(handler(s.Ctx, createProposal))

	// Public position not found
	paramChangeProposal := types.NewPublicPositionParameterChangeProposal(
		"Title", "Description", []types.PublicPositionParameterChange{
			types.NewPublicPositionParameterChange(3, utils.ParseDec("0.002")),
		})
	s.Require().NoError(paramChangeProposal.ValidateBasic())
	s.Ctx, _ = ctx.CacheContext()
	s.Require().EqualError(
		handler(s.Ctx, paramChangeProposal), "public position 3 not found: not found")

	// Public position parameter changed.
	paramChangeProposal = types.NewPublicPositionParameterChangeProposal(
		"Title", "Description", []types.PublicPositionParameterChange{
			types.NewPublicPositionParameterChange(1, utils.ParseDec("0.002")),
		})
	s.Require().NoError(paramChangeProposal.ValidateBasic())
	s.Ctx = ctx
	s.Require().NoError(handler(s.Ctx, paramChangeProposal))

	publicPosition, found := s.keeper.GetPublicPosition(s.Ctx, 1)
	s.Require().True(found)
	s.AssertEqual(utils.ParseDec("0.002"), publicPosition.FeeRate)
}
