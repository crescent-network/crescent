package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetAccAddress(address string) sdk.AccAddress {
	if address == "" {
		return nil
	}
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return addr
}

func (mm MarketMaker) GetAccAddress() sdk.AccAddress {
	return GetAccAddress(mm.Address)
}

func (mm MarketMaker) Validate() error {
	return ValidateMarketMaker(mm.Address, mm.PairId)
}

func (i Incentive) GetAccAddress() sdk.AccAddress {
	return GetAccAddress(i.Address)
}

func (i Incentive) Validate() error {
	_, err := sdk.AccAddressFromBech32(i.Address)
	if err != nil {
		return err
	}
	return i.Claimable.Validate()
}

func ValidateMarketMaker(address string, pairId uint64) error {
	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	if pairId == uint64(0) {
		return ErrInvalidPairId
	}
	return nil
}

func (mm MarketMakerHandle) Validate() error {
	return ValidateMarketMaker(mm.Address, mm.PairId)
}

func (mm MarketMakerHandle) GetAccAddress() sdk.AccAddress {
	return GetAccAddress(mm.Address)
}

func (id IncentiveDistribution) Validate() error {
	if err := ValidateMarketMaker(id.Address, id.PairId); err != nil {
		return err
	}
	if len(id.Amount) == 0 {
		return fmt.Errorf("incentive distribution amount should be not empty")
	}
	return id.Amount.Validate()
}

func (id IncentiveDistribution) GetAccAddress() sdk.AccAddress {
	return GetAccAddress(id.Address)
}

func UnmarshalMarketMaker(cdc codec.BinaryCodec, value []byte) (mm MarketMaker, err error) {
	err = cdc.Unmarshal(value, &mm)
	return mm, err
}

func (idr DepositRecord) Validate() error {
	if err := ValidateMarketMaker(idr.Address, idr.PairId); err != nil {
		return err
	}
	return idr.Amount.Validate()
}

func (idr DepositRecord) GetAccAddress() sdk.AccAddress {
	return GetAccAddress(idr.Address)
}
