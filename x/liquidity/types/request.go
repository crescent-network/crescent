package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (req DepositRequest) GetDepositor() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		panic(err)
	}
	return addr
}

func (req WithdrawRequest) GetWithdrawer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.Withdrawer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (req SwapRequest) GetOrderer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(req.Orderer)
	if err != nil {
		panic(err)
	}
	return addr
}

func (req SwapRequest) Order() *Order {
	return &Order{
		RequestId:       &req.Id,
		Orderer:         req.GetOrderer(),
		Direction:       req.Direction,
		Price:           req.Price,
		OrderAmount:     req.RemainingAmount, // TODO: introduce new OrderAmount field in SwapRequest?
		RemainingAmount: req.RemainingAmount,
		ReceivedAmount:  sdk.ZeroInt(),
	}
}
