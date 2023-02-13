package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func DeriveBootstrapPoolEscrowAddress(id uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("BootstrapPoolEscrowAddress/%d", id)))
}

func DeriveBootstrapPoolFeeCollectorAddress(id uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("BootstrapPoolFeeCollectorAddress/%s", id)))
}

func NewBootstrapPool(id uint64, bootstrapCoinDenom, QuoteCoinDenom string, minPrice, maxPrice sdk.Dec, proposer sdk.AccAddress) BootstrapPool {
	return BootstrapPool{
		Id: id,
		//BootstrapCoin:       sdk.Coin{},
		BootstrapCoinDenom: bootstrapCoinDenom,
		QuoteCoinDenom:     QuoteCoinDenom,
		MinPrice:           &minPrice,
		MaxPrice:           &maxPrice,
		//StageSchedule:       nil,
		ProposerAddress:     proposer.String(),
		EscrowAddress:       DeriveBootstrapPoolEscrowAddress(id).String(),
		FeeCollectorAddress: DeriveBootstrapPoolFeeCollectorAddress(id).String(),
	}
}

func (m BootstrapPool) GetProposer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ProposerAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (m BootstrapPool) GetEscrowAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.EscrowAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (m BootstrapPool) GetFeeCollector() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.FeeCollectorAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// TODO: GetFeeCollector
// TODO: GetEscrowAddress
// TODO: GetProposer

//func GetAccAddress(address string) sdk.AccAddress {
//	if address == "" {
//		return nil
//	}
//	addr, err := sdk.AccAddressFromBech32(address)
//	if err != nil {
//		panic(err)
//	}
//	return addr
//}
//
//func (mm Bootstrap) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(mm.Address)
//}
//
//func (mm Bootstrap) Validate() error {
//	return ValidateBootstrap(mm.Address, mm.PairId)
//}
//
//func (i Incentive) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(i.Address)
//}
//
//func (i Incentive) Validate() error {
//	_, err := sdk.AccAddressFromBech32(i.Address)
//	if err != nil {
//		return err
//	}
//	return i.Claimable.Validate()
//}
//
//func ValidateBootstrap(address string, pairId uint64) error {
//	_, err := sdk.AccAddressFromBech32(address)
//	if err != nil {
//		return err
//	}
//
//	if pairId == uint64(0) {
//		return ErrInvalidPairId
//	}
//	return nil
//}
//
//func (mm BootstrapHandle) Validate() error {
//	return ValidateBootstrap(mm.Address, mm.PairId)
//}
//
//func (mm BootstrapHandle) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(mm.Address)
//}
//
//func (id IncentiveDistribution) Validate() error {
//	if err := ValidateBootstrap(id.Address, id.PairId); err != nil {
//		return err
//	}
//	if len(id.Amount) == 0 {
//		return fmt.Errorf("incentive distribution amount should be not empty")
//	}
//	return id.Amount.Validate()
//}
//
//func (id IncentiveDistribution) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(id.Address)
//}
//
//func UnmarshalBootstrap(cdc codec.BinaryCodec, value []byte) (mm Bootstrap, err error) {
//	err = cdc.Unmarshal(value, &mm)
//	return mm, err
//}
//
//func (idr DepositRecord) Validate() error {
//	if err := ValidateBootstrap(idr.Address, idr.PairId); err != nil {
//		return err
//	}
//	return idr.Amount.Validate()
//}
//
//func (idr DepositRecord) GetAccAddress() sdk.AccAddress {
//	return GetAccAddress(idr.Address)
//}
