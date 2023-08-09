package params

// Default simulation operation weights for messages and gov proposals.
const (
	// deprecated farming module
	DefaultWeightMsgCreateFixedAmountPlan int = 0
	DefaultWeightMsgCreateRatioPlan       int = 0
	DefaultWeightMsgStake                 int = 0
	DefaultWeightMsgUnstake               int = 0
	DefaultWeightMsgHarvest               int = 0
	DefaultWeightMsgRemovePlan            int = 0

	// deprecated liquidity module
	DefaultWeightMsgCreatePair       int = 0
	DefaultWeightMsgCreatePool       int = 0
	DefaultWeightMsgCreateRangedPool int = 0
	DefaultWeightMsgDeposit          int = 0
	DefaultWeightMsgWithdraw         int = 0
	DefaultWeightMsgLimitOrder       int = 0
	DefaultWeightMsgMarketOrder      int = 0
	DefaultWeightMsgMMOrder          int = 0
	DefaultWeightMsgCancelOrder      int = 0
	DefaultWeightMsgCancelAllOrders  int = 0
	DefaultWeightMsgCancelMMOrder    int = 0

	// Deprecated claim module
	DefaultWeightMsgClaim int = 0

	// Deprecated liquidfarming module
	DefaultWeightMsgLiquidFarm   int = 0
	DefaultWeightMsgLiquidUnfarm int = 0
	DefaultWeightMsgPlaceBid     int = 0
	DefaultWeightMsgRefundBid    int = 0

	DefaultWeightAddPublicPlanProposal    int = 5
	DefaultWeightUpdatePublicPlanProposal int = 5
	DefaultWeightDeletePublicPlanProposal int = 5

	DefaultWeightMsgLiquidStake   int = 80
	DefaultWeightMsgLiquidUnstake int = 30

	DefaultWeightAddWhitelistValidatorsProposal    int = 50
	DefaultWeightUpdateWhitelistValidatorsProposal int = 5
	DefaultWeightDeleteWhitelistValidatorsProposal int = 5
	DefaultWeightCompleteRedelegationUnbonding     int = 2
	DefaultWeightTallyWithLiquidStaking            int = 30

	DefaultWeightMsgApplyMarketMaker int = 20
	DefaultWeightMsgClaimIncentives  int = 10

	DefaultWeightMarketMakerProposal  int = 20
	DefaultWeightChangeIncentivePairs int = 5
	DefaultWeightChangeDepositAmount  int = 2

	DefaultWeightUpdateBudgetPlans int = 5
)
