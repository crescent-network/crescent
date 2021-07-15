package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	proto "github.com/gogo/protobuf/proto"
)

const (
	ProposalTypePublicPlan string = "PublicPlan"
)

// Implements Proposal Interface
var _ gov.Content = &PublicPlanProposal{}

func init() {
	gov.RegisterProposalType(ProposalTypePublicPlan)
	gov.RegisterProposalTypeCodec(&PublicPlanProposal{}, "cosmos-sdk/PublicPlanProposal")
}

func NewPublicPlanProposal(title, description string, plans []PlanI) (gov.Content, error) {
	plansAny, err := PackPlans(plans)
	if err != nil {
		panic(err)
	}

	return &PublicPlanProposal{
		Title:       title,
		Description: description,
		Plans:       plansAny,
	}, nil
}

func (p *PublicPlanProposal) GetTitle() string { return p.Title }

func (p *PublicPlanProposal) GetDescription() string { return p.Description }

func (p *PublicPlanProposal) ProposalRoute() string { return RouterKey }

func (p *PublicPlanProposal) ProposalType() string { return ProposalTypePublicPlan }

func (p *PublicPlanProposal) ValidateBasic() error {
	for _, plan := range p.Plans {
		_, ok := plan.GetCachedValue().(PlanI)
		if !ok {
			return fmt.Errorf("expected planI")
		}
		// TODO: PlanI needs ValidateBasic()?
		// if err := p.ValidateBasic(); err != nil {
		// 	return err
		// }
	}
	return gov.ValidateAbstract(p)
}

func (p PublicPlanProposal) String() string {
	return fmt.Sprintf(`Create FixedAmountPlan Proposal:
  Title:       %s
  Description: %s
  Plans: 	   %s
`, p.Title, p.Description, p.Plans)
}

// PackPlans converts PlanIs to Any slice.
func PackPlans(plans []PlanI) ([]*types.Any, error) {
	plansAny := make([]*types.Any, len(plans))
	for i, plan := range plans {
		msg, ok := plan.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("cannot proto marshal %T", plan)
		}
		any, err := types.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
		plansAny[i] = any
	}

	return plansAny, nil
}

// UnpackPlans converts Any slice to PlanIs.
func UnpackPlans(plansAny []*types.Any) ([]PlanI, error) {
	plans := make([]PlanI, len(plansAny))
	for i, any := range plansAny {
		p, ok := any.GetCachedValue().(PlanI)
		if !ok {
			return nil, fmt.Errorf("expected planI")
		}
		plans[i] = p
	}

	return plans, nil
}

// UnpackPlan converts Any slice to PlanI.
func UnpackPlan(any *types.Any) (PlanI, error) {
	p, ok := any.GetCachedValue().(PlanI)
	if !ok {
		return nil, fmt.Errorf("expected planI")
	}

	return p, nil
}
