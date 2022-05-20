package types

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

const (
	MaxNameLength                   int    = 140
	PrivatePlanFarmingPoolAccPrefix string = "PrivatePlan"
	StakingReserveAccPrefix         string = "StakingReserveAcc"
	AccNameSplitter                 string = "|"
)

var (
	_ PlanI = (*FixedAmountPlan)(nil)
	_ PlanI = (*RatioPlan)(nil)

	planNameRegexp = regexp.MustCompile(`^[[:print:]]+$`)
)

// NewBasePlan creates a new BasePlan object
//nolint:interfacer
func NewBasePlan(id uint64, name string, typ PlanType, farmingPoolAddr, terminationAddr string, coinWeights sdk.DecCoins, startTime, endTime time.Time) *BasePlan {
	basePlan := &BasePlan{
		Id:                   id,
		Name:                 name,
		Type:                 typ,
		FarmingPoolAddress:   farmingPoolAddr,
		TerminationAddress:   terminationAddr,
		StakingCoinWeights:   coinWeights,
		StartTime:            startTime,
		EndTime:              endTime,
		Terminated:           false,
		LastDistributionTime: nil,
		DistributedCoins:     sdk.NewCoins(),
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

func (plan BasePlan) GetName() string { //nolint:golint
	return plan.Name
}

func (plan *BasePlan) SetName(name string) error { //nolint:golint
	plan.Name = name
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
	addr, err := sdk.AccAddressFromBech32(plan.FarmingPoolAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (plan *BasePlan) SetFarmingPoolAddress(addr sdk.AccAddress) error {
	plan.FarmingPoolAddress = addr.String()
	return nil
}

func (plan BasePlan) GetTerminationAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(plan.TerminationAddress)
	if err != nil {
		panic(err)
	}
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

func (plan *BasePlan) IsTerminated() bool {
	return plan.Terminated
}

func (plan *BasePlan) SetTerminated(terminated bool) error {
	plan.Terminated = terminated
	return nil
}

func (plan *BasePlan) GetLastDistributionTime() *time.Time {
	return plan.LastDistributionTime
}

func (plan *BasePlan) SetLastDistributionTime(t *time.Time) error {
	plan.LastDistributionTime = t
	return nil
}

func (plan *BasePlan) GetDistributedCoins() sdk.Coins {
	return plan.DistributedCoins
}

func (plan *BasePlan) SetDistributedCoins(distributedCoins sdk.Coins) error {
	plan.DistributedCoins = distributedCoins
	return nil
}

func (plan BasePlan) GetBasePlan() *BasePlan {
	return &BasePlan{
		Id:                   plan.GetId(),
		Name:                 plan.GetName(),
		Type:                 plan.GetType(),
		FarmingPoolAddress:   plan.GetFarmingPoolAddress().String(),
		TerminationAddress:   plan.GetTerminationAddress().String(),
		StakingCoinWeights:   plan.GetStakingCoinWeights(),
		StartTime:            plan.GetStartTime(),
		EndTime:              plan.GetEndTime(),
		Terminated:           plan.IsTerminated(),
		LastDistributionTime: plan.GetLastDistributionTime(),
		DistributedCoins:     plan.GetDistributedCoins(),
	}
}

// Validate checks for errors on the Plan fields
func (plan BasePlan) Validate() error {
	if plan.Type != PlanTypePrivate && plan.Type != PlanTypePublic {
		return sdkerrors.Wrapf(ErrInvalidPlanType, "unknown plan type: %s", plan.Type)
	}
	if _, err := sdk.AccAddressFromBech32(plan.FarmingPoolAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", plan.FarmingPoolAddress, err)
	}
	if _, err := sdk.AccAddressFromBech32(plan.TerminationAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid termination address %q: %v", plan.TerminationAddress, err)
	}
	if err := ValidatePlanName(plan.Name); err != nil {
		return sdkerrors.Wrap(ErrInvalidPlanName, err.Error())
	}
	if err := ValidateStakingCoinTotalWeights(plan.StakingCoinWeights); err != nil {
		return err
	}
	if !plan.EndTime.After(plan.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", plan.EndTime, plan.StartTime)
	}
	if err := plan.DistributedCoins.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid distributed coins: %v", err)
	}
	return nil
}

// NewFixedAmountPlan returns a new fixed amount plan.
func NewFixedAmountPlan(basePlan *BasePlan, epochAmount sdk.Coins) *FixedAmountPlan {
	return &FixedAmountPlan{
		BasePlan:    basePlan,
		EpochAmount: epochAmount,
	}
}

// NewRatioPlan returns a new ratio plan.
func NewRatioPlan(basePlan *BasePlan, epochRatio sdk.Dec) *RatioPlan {
	return &RatioPlan{
		BasePlan:   basePlan,
		EpochRatio: epochRatio,
	}
}

// PlanI represents a farming plan.
type PlanI interface {
	proto.Message

	GetId() uint64
	SetId(uint64) error

	GetName() string
	SetName(string) error

	GetType() PlanType
	SetType(PlanType) error

	GetFarmingPoolAddress() sdk.AccAddress
	SetFarmingPoolAddress(sdk.AccAddress) error

	GetTerminationAddress() sdk.AccAddress
	SetTerminationAddress(sdk.AccAddress) error

	GetStakingCoinWeights() sdk.DecCoins
	SetStakingCoinWeights(sdk.DecCoins) error

	GetStartTime() time.Time
	SetStartTime(time.Time) error

	GetEndTime() time.Time
	SetEndTime(time.Time) error

	IsTerminated() bool
	SetTerminated(bool) error

	GetLastDistributionTime() *time.Time
	SetLastDistributionTime(*time.Time) error

	GetDistributedCoins() sdk.Coins
	SetDistributedCoins(sdk.Coins) error

	GetBasePlan() *BasePlan

	Validate() error
}

// ValidateTotalEpochRatio validates a farmer's total epoch ratio that must be equal to 1.
func ValidateTotalEpochRatio(plans []PlanI) error {
	for i, plan := range plans {
		plan, ok := plan.(*RatioPlan)
		if !ok {
			continue
		}

		totalRatio := plan.EpochRatio
		for j, otherPlan := range plans {
			if i == j {
				continue
			}
			otherPlan, ok := otherPlan.(*RatioPlan)
			if !ok {
				continue
			}
			if otherPlan.FarmingPoolAddress == plan.FarmingPoolAddress &&
				DateRangeIncludes(
					otherPlan.GetStartTime(), otherPlan.GetEndTime(), plan.GetStartTime()) {
				totalRatio = totalRatio.Add(otherPlan.EpochRatio)
			}
		}
		if totalRatio.GT(sdk.OneDec()) {
			return sdkerrors.Wrap(ErrInvalidTotalEpochRatio, "total epoch ratio must be lower than 1")
		}
	}

	return nil
}

// ValidateEpochRatio validate a epoch ratio that must be positive and less than 1.
func ValidateEpochRatio(epochRatio sdk.Dec) error {
	if !epochRatio.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "epoch ratio must be positive: %s", epochRatio)
	}
	if epochRatio.GT(sdk.OneDec()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "epoch ratio must be less than 1: %s", epochRatio)
	}
	return nil
}

// ValidateEpochAmount validate a epoch amount that must be valid coins.
func ValidateEpochAmount(epochAmount sdk.Coins) error {
	if err := epochAmount.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid epoch amount: %v", err)
	}
	return nil
}

// PackPlan converts PlanI to Any
func PackPlan(plan PlanI) (*codectypes.Any, error) {
	any, err := codectypes.NewAnyWithValue(plan)
	if err != nil {
		return nil, err
	}
	return any, nil
}

// UnpackPlan converts Any to PlanI.
func UnpackPlan(any *codectypes.Any) (PlanI, error) {
	if any == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot unpack nil")
	}
	if any.TypeUrl == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "empty type url")
	}
	var plan PlanI
	v := any.GetCachedValue()
	if v == nil {
		registry := codectypes.NewInterfaceRegistry()
		RegisterInterfaces(registry)
		if err := registry.UnpackAny(any, &plan); err != nil {
			return nil, err
		}
		return plan, nil
	}
	plan, ok := v.(PlanI)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot unpack Plan from %T", v)
	}
	return plan, nil
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
func ValidateStakingCoinTotalWeights(weights sdk.DecCoins) error {
	if weights.Empty() {
		return sdkerrors.Wrap(ErrInvalidStakingCoinWeights, "staking coin weights must not be empty")
	}
	if err := weights.Validate(); err != nil {
		return sdkerrors.Wrapf(ErrInvalidStakingCoinWeights, "invalid staking coin weights: %v", err)
	}
	totalWeight := sdk.ZeroDec()
	for _, w := range weights {
		totalWeight = totalWeight.Add(w.Amount)
	}
	if !totalWeight.Equal(sdk.OneDec()) {
		return sdkerrors.Wrap(ErrInvalidStakingCoinWeights, "total weight must be 1")
	}
	return nil
}

// ValidatePlanName validates a plan name.
func ValidatePlanName(name string) error {
	if strings.TrimSpace(name) != name {
		return fmt.Errorf("extra whitespaces around name")
	}
	if name == "" {
		return fmt.Errorf("plan name must not be empty")
	}
	if !planNameRegexp.MatchString(name) {
		return fmt.Errorf("name contains invalid characters: %s", name)
	}
	if strings.Contains(name, AccNameSplitter) {
		return fmt.Errorf("plan name cannot contain %s", AccNameSplitter)
	}
	if len(name) > MaxNameLength {
		return fmt.Errorf("plan name cannot be longer than max length of %d", MaxNameLength)
	}
	return nil
}

// IsPlanActiveAt returns if the plan is active at given time t.
func IsPlanActiveAt(plan PlanI, t time.Time) bool {
	return !plan.GetStartTime().After(t) && plan.GetEndTime().After(t)
}

// PrivatePlanFarmingPoolAcc returns a unique farming pool address for a newly created plan.
func PrivatePlanFarmingPoolAcc(name string, planId uint64) sdk.AccAddress {
	poolAccName := strings.Join([]string{PrivatePlanFarmingPoolAccPrefix, fmt.Sprint(planId), name}, AccNameSplitter)
	return DeriveAddress(ReserveAddressType, ModuleName, poolAccName)
}

// StakingReserveAcc returns module account for the staking reserve pool account by staking coin denom and type.
func StakingReserveAcc(stakingCoinDenom string) sdk.AccAddress {
	return DeriveAddress(ReserveAddressType, ModuleName, StakingReserveAccPrefix+AccNameSplitter+stakingCoinDenom)
}
