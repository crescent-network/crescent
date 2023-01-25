package types

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
