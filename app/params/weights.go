package params

// Default simulation operation weights for messages and gov proposals.
const (
	DefaultWeightMsgCreateFixedAmountPlan int = 10
	DefaultWeightMsgCreateRatioPlan       int = 10
	DefaultWeightMsgStake                 int = 85
	DefaultWeightMsgUnstake               int = 30
	DefaultWeightMsgHarvest               int = 30

	DefaultWeightMsgCreatePair      int = 5
	DefaultWeightMsgCreatePool      int = 10
	DefaultWeightMsgDeposit         int = 15
	DefaultWeightMsgWithdraw        int = 15
	DefaultWeightMsgLimitOrder      int = 80
	DefaultWeightMsgMarketOrder     int = 60
	DefaultWeightMsgCancelOrder     int = 20
	DefaultWeightMsgCancelAllOrders int = 20

	DefaultWeightAddPublicPlanProposal    int = 5
	DefaultWeightUpdatePublicPlanProposal int = 5
	DefaultWeightDeletePublicPlanProposal int = 5
)
