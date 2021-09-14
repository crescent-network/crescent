package params

const (
	// farming module simulation operation weights for messages
	DefaultWeightMsgCreateFixedAmountPlan int = 10
	DefaultWeightMsgCreateRatioPlan       int = 10
	DefaultWeightMsgStake                 int = 85
	DefaultWeightMsgUnstake               int = 30
	DefaultWeightMsgHarvest               int = 30

	DefaultWeightAddPublicPlanProposal    int = 10
	DefaultWeightUpdatePublicPlanProposal int = 5
	DefaultWeightDeletePublicPlanProposal int = 5
)
