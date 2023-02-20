package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	utils "github.com/crescent-network/crescent/v4/types"
	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

func (s *KeeperTestSuite) TestBootstrapProposal() {
	ctx := s.ctx
	//k := s.keeper
	//params := k.GetParams(ctx)

	proposal := types.BootstrapProposal{
		Title:           "test",
		Description:     "test",
		ProposerAddress: s.addrs[1].String(),
		OfferCoins:      sdk.Coins{sdk.NewCoin(denom1, sdk.NewInt(100_000_000))},
		BaseCoinDenom:   denom1,
		QuoteCoinDenom:  denom2,
		MinPrice:        nil,
		MaxPrice:        nil,
		PairId:          s.pairs[0].Id,
		InitialOrders: []types.InitialOrder{
			{
				OfferCoin: sdk.NewCoin(denom1, sdk.NewInt(50_000_000)),
				Price:     sdk.OneDec(),
				Direction: types.OrderDirectionSell,
			},
			{
				OfferCoin: sdk.NewCoin(denom1, sdk.NewInt(50_000_000)),
				Price:     sdk.OneDec(),
				Direction: types.OrderDirectionSell,
			},
		},
		StartTime:     utils.ParseTime("2023-02-14T00:00:00Z"),
		NumOfStages:   5,
		StageDuration: 24 * time.Hour,
	}
	s.Require().NoError(proposal.ValidateBasic())
	fmt.Println(s.app.BankKeeper.GetAllBalances(ctx, s.addrs[1]))
	s.handleProposal(&proposal)

	bp, found := s.keeper.GetBootstrapPool(ctx, 1)
	s.Require().True(found)

	fmt.Println(bp.Stages)

	fmt.Println(s.app.BankKeeper.GetAllBalances(ctx, bp.GetEscrowAddress()))
	fmt.Println(s.app.BankKeeper.GetAllBalances(ctx, bp.GetFeeCollector()))
	fmt.Println(s.app.BankKeeper.GetAllBalances(ctx, s.addrs[1]))

	fmt.Println(s.keeper.GetAllOrders(ctx))

	// TODO: assert orders
	// TODO: assert stage

	// TODO: add test cases using malleate
	//for _, tc := range []struct {
	//	name        string
	//	malleate    func(*types.PublicPlanProposal)
	//	expectedErr string
	//}{
	//	{
	//		"happy case",
	//		func(proposal *types.PublicPlanProposal) {},
	//		"",
	//	},
	//	{
	//		"empty proposals",
	//		func(proposal *types.PublicPlanProposal) {
	//			proposal.AddPlanRequests = []types.AddPlanRequest{}
	//			proposal.ModifyPlanRequests = []types.ModifyPlanRequest{}
	//			proposal.DeletePlanRequests = []types.DeletePlanRequest{}
	//		},
	//		"proposal request must not be empty: invalid request",
	//	},
	//	{
	//		"invalid add request proposal",
	//		func(proposal *types.PublicPlanProposal) {
	//			proposal.AddPlanRequests[0].Name = strings.Repeat("a", 256)
	//		},
	//		"plan name cannot be longer than max length of 140: invalid plan name",
	//	},
	//	{
	//		"invalid update request proposal",
	//		func(proposal *types.PublicPlanProposal) {
	//			proposal.ModifyPlanRequests[0].Name = strings.Repeat("a", 256)
	//		},
	//		"plan name cannot be longer than max length of 140: invalid plan name",
	//	},
	//	{
	//		"invalid delete request proposal",
	//		func(proposal *types.PublicPlanProposal) {
	//			proposal.DeletePlanRequests[0].PlanId = 0
	//		},
	//		"invalid plan id: 0: invalid request",
	//	},
	//} {
	//	t.Run(tc.name, func(t *testing.T) {
	//		proposal := types.NewPublicPlanProposal(
	//			"title",
	//			"description",
	//			[]types.AddPlanRequest{
	//				{
	//					Name:               "name",
	//					FarmingPoolAddress: sdk.AccAddress(crypto.AddressHash([]byte("address1"))).String(),
	//					TerminationAddress: sdk.AccAddress(crypto.AddressHash([]byte("address1"))).String(),
	//					StakingCoinWeights: sdk.NewDecCoins(sdk.NewInt64DecCoin("stake1", 1)),
	//					StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
	//					EndTime:            types.ParseTime("9999-12-31T00:00:00Z"),
	//					EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("reward1", 10000000)),
	//				},
	//			},
	//			[]types.ModifyPlanRequest{
	//				{
	//					PlanId:             1,
	//					Name:               "new name",
	//					FarmingPoolAddress: sdk.AccAddress(crypto.AddressHash([]byte("address2"))).String(),
	//					TerminationAddress: sdk.AccAddress(crypto.AddressHash([]byte("address2"))).String(),
	//					StakingCoinWeights: sdk.NewDecCoins(sdk.NewInt64DecCoin("stake2", 1)),
	//					EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("reward2", 10000000)),
	//				},
	//			},
	//			[]types.DeletePlanRequest{
	//				{
	//					PlanId: 1,
	//				},
	//			},
	//		)
	//		tc.malleate(proposal)
	//		err := proposal.ValidateBasic()
	//		if tc.expectedErr == "" {
	//			require.NoError(t, err)
	//		} else {
	//			require.EqualError(t, err, tc.expectedErr)
	//		}
	//	})
	//}
}
