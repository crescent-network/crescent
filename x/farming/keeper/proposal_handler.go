package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/farming/x/farming/types"
)

// HandlePublicPlanProposal is a handler for executing a fixed amount plan creation proposal.
func HandlePublicPlanProposal(ctx sdk.Context, k Keeper, plansAny []*codectypes.Any) error {
	plans, err := types.UnpackPlans(plansAny)
	if err != nil {
		return err
	}

	for _, plan := range plans {
		switch p := plan.(type) {
		case *types.FixedAmountPlan:
			msg := types.NewMsgCreateFixedAmountPlan(
				p.GetFarmingPoolAddress(),
				p.GetStakingCoinWeights(),
				p.GetStartTime(),
				p.GetEndTime(),
				p.GetEpochDays(),
				p.EpochAmount,
			)

			fixedPlan := k.CreateFixedAmountPlan(ctx, msg, types.PlanTypePublic)

			logger := k.Logger(ctx)
			logger.Info("created public fixed amount plan", "fixed_amount_plan", fixedPlan)

		case *types.RatioPlan:
			msg := types.NewMsgCreateRatioPlan(
				p.GetFarmingPoolAddress(),
				p.GetStakingCoinWeights(),
				p.GetStartTime(),
				p.GetEndTime(),
				p.GetEpochDays(),
				p.EpochRatio,
			)

			ratioPlan := k.CreateRatioPlan(ctx, msg, types.PlanTypePublic)

			logger := k.Logger(ctx)
			logger.Info("created public fixed amount plan", "ratio_plan", ratioPlan)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized farming proposal plan type: %T", p)
		}
	}

	return nil
}
