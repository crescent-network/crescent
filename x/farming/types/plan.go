package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	MaxNameLength int = 140
)

var (
	_ PlanI = (*FixedAmountPlan)(nil)
	_ PlanI = (*RatioPlan)(nil)
)

// NewBasePlan creates a new BasePlan object
//nolint:interfacer
func NewBasePlan(id uint64, name string, typ PlanType, farmingPoolAddr, terminationAddr string, coinWeights sdk.DecCoins, startTime, endTime time.Time) *BasePlan {
	basePlan := &BasePlan{
		Id:                 id,
		Name:               name,
		Type:               typ,
		FarmingPoolAddress: farmingPoolAddr,
		RewardPoolAddress:  GenerateRewardPoolAcc(PlanUniqueKey(id, typ, farmingPoolAddr)).String(),
		TerminationAddress: terminationAddr,
		StakingCoinWeights: coinWeights,
		StartTime:          startTime,
		EndTime:            endTime,
	}
	return basePlan
}

func (plan BasePlan) GetId() uint64 { //nolint:golint
	return plan.Id
}

func (plan *BasePlan) SetId(id uint64) error { //nolint:golint
	plan.Id = id
	return nil
}

func (plan BasePlan) GetType() PlanType {
	return plan.Type
}

func (plan *BasePlan) SetType(typ PlanType) error {
	plan.Type = typ
	return nil
}

func (plan BasePlan) GetFarmingPoolAddress() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(plan.FarmingPoolAddress)
	return addr
}

func (plan *BasePlan) SetFarmingPoolAddress(addr sdk.AccAddress) error {
	plan.FarmingPoolAddress = addr.String()
	return nil
}

func (plan BasePlan) GetRewardPoolAddress() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(plan.RewardPoolAddress)
	return addr
}

func (plan *BasePlan) SetRewardPoolAddress(addr sdk.AccAddress) error {
	plan.RewardPoolAddress = addr.String()
	return nil
}

func (plan BasePlan) GetTerminationAddress() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(plan.TerminationAddress)
	return addr
}

func (plan *BasePlan) SetTerminationAddress(addr sdk.AccAddress) error {
	plan.TerminationAddress = addr.String()
	return nil
}

func (plan BasePlan) GetStakingCoinWeights() sdk.DecCoins {
	return plan.StakingCoinWeights
}

func (plan *BasePlan) SetStakingCoinWeights(coinWeights sdk.DecCoins) error {
	plan.StakingCoinWeights = coinWeights
	return nil
}

func (plan BasePlan) GetStartTime() time.Time {
	return plan.StartTime
}

func (plan *BasePlan) SetStartTime(t time.Time) error {
	plan.StartTime = t
	return nil
}

func (plan BasePlan) GetEndTime() time.Time {
	return plan.EndTime
}

func (plan *BasePlan) SetEndTime(t time.Time) error {
	plan.EndTime = t
	return nil
}

// Validate checks for errors on the Plan fields
func (plan BasePlan) Validate() error {
	if plan.Type != PlanTypePrivate && plan.Type != PlanTypePublic {
		return sdkerrors.Wrapf(ErrInvalidPlanType, "unknown plan type: %s", plan.Type)
	}
	if _, err := sdk.AccAddressFromBech32(plan.FarmingPoolAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", plan.FarmingPoolAddress, err)
	}
	if _, err := sdk.AccAddressFromBech32(plan.RewardPoolAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid reward pool address %q: %v", plan.RewardPoolAddress, err)
	}
	if _, err := sdk.AccAddressFromBech32(plan.TerminationAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid termination address %q: %v", plan.TerminationAddress, err)
	}
	if len(plan.Name) > MaxNameLength {
		return sdkerrors.Wrapf(ErrInvalidNameLength, "plan name cannot be longer than max length of %d", MaxNameLength)
	}
	if plan.StakingCoinWeights.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking coin weights must not be empty")
	}
	if err := plan.StakingCoinWeights.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid staking coin weights: %v", err)
	}
	if ok := ValidateStakingCoinTotalWeights(plan.StakingCoinWeights); !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total weight must be 1")
	}
	if !plan.EndTime.After(plan.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", plan.EndTime, plan.StartTime)
	}
	return nil
}

func (plan BasePlan) String() string {
	out, _ := plan.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of an Plan.
func (plan BasePlan) MarshalYAML() (interface{}, error) {
	bz, err := codec.MarshalYAML(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), &plan)
	if err != nil {
		return nil, err
	}
	return string(bz), err
}

func NewFixedAmountPlan(basePlan *BasePlan, epochAmount sdk.Coins) *FixedAmountPlan {
	return &FixedAmountPlan{
		BasePlan:    basePlan,
		EpochAmount: epochAmount,
	}
}

func NewRatioPlan(basePlan *BasePlan, epochRatio sdk.Dec) *RatioPlan {
	return &RatioPlan{
		BasePlan:   basePlan,
		EpochRatio: epochRatio,
	}
}

// PlanUniqueKey returns unique name of the plan consists of given Id, Type and FarmingPoolAddress.
func PlanUniqueKey(id uint64, typ PlanType, farmingPoolAddr string) string {
	poolNameObjects := make([]string, 3)
	poolNameObjects[0] = strconv.FormatUint(id, 10)
	poolNameObjects[1] = strconv.FormatInt(int64(typ), 10)
	poolNameObjects[2] = farmingPoolAddr
	return strings.Join(poolNameObjects, "/")
}

// GenerateRewardPoolAcc returns deterministically generated reward pool account for the given plan name
func GenerateRewardPoolAcc(name string) sdk.AccAddress {
	return address.Module(ModuleName, []byte(strings.Join([]string{RewardPoolAccKeyPrefix, name}, "/")))
}

type PlanI interface {
	proto.Message

	GetId() uint64
	SetId(uint64) error

	GetType() PlanType
	SetType(PlanType) error

	GetFarmingPoolAddress() sdk.AccAddress
	SetFarmingPoolAddress(sdk.AccAddress) error

	GetRewardPoolAddress() sdk.AccAddress
	SetRewardPoolAddress(sdk.AccAddress) error

	GetTerminationAddress() sdk.AccAddress
	SetTerminationAddress(sdk.AccAddress) error

	GetStakingCoinWeights() sdk.DecCoins
	SetStakingCoinWeights(sdk.DecCoins) error

	GetStartTime() time.Time
	SetStartTime(time.Time) error

	GetEndTime() time.Time
	SetEndTime(time.Time) error

	String() string
}

// ValidateRatioPlans validates a farmer's total epoch ratio and plan name.
// A total epoch ratio cannot be higher than 1 and plan name must not be duplicate.
func ValidateRatioPlans(i interface{}) error {
	plans, ok := i.([]PlanI)
	if !ok {
		return sdkerrors.Wrapf(ErrInvalidPlanType, "invalid plan type %T", i)
	}

	totalEpochRatio := make(map[string]sdk.Dec)
	names := make(map[string]bool)

	for _, plan := range plans {
		farmingPoolAddr := plan.GetFarmingPoolAddress().String()

		if plan, ok := plan.(*RatioPlan); ok {
			if err := plan.Validate(); err != nil {
				return err
			}

			if epochRatio, ok := totalEpochRatio[farmingPoolAddr]; ok {
				totalEpochRatio[farmingPoolAddr] = epochRatio.Add(plan.EpochRatio)
			} else {
				totalEpochRatio[farmingPoolAddr] = plan.EpochRatio
			}

			if _, ok := names[plan.Name]; ok {
				return sdkerrors.Wrap(ErrDuplicatePlanName, plan.Name)
			}
			names[plan.Name] = true
		}
	}

	for _, farmerRatio := range totalEpochRatio {
		if farmerRatio.GT(sdk.NewDec(1)) {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total epoch ratio must be lower than 1")
		}
	}

	return nil
}

// UnpackPlan converts Any to PlanI.
func UnpackPlan(any *codectypes.Any) (PlanI, error) {
	v := any.GetCachedValue()
	p, ok := v.(PlanI)
	if !ok {
		return nil, fmt.Errorf("expected PlanI, got %T", v)
	}
	return p, nil
}

// UnpackPlans converts Any slice to PlanIs.
func UnpackPlans(plansAny []*codectypes.Any) ([]PlanI, error) {
	plans := make([]PlanI, len(plansAny))
	for i, any := range plansAny {
		p, err := UnpackPlan(any)
		if err != nil {
			return nil, err
		}
		plans[i] = p
	}
	return plans, nil
}

// ValidateStakingCoinTotalWeights validates the total staking coin weights must be equal to 1.
func ValidateStakingCoinTotalWeights(weights sdk.DecCoins) bool {
	totalWeight := sdk.ZeroDec()
	for _, w := range weights {
		totalWeight = totalWeight.Add(w.Amount)
	}
	return totalWeight.Equal(sdk.NewDec(1))
}
