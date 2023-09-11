package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ParamChange struct {
	MakerFeeRate *sdk.Dec
	TakerFeeRate *sdk.Dec
	TickSpacing  *uint32
}

var ParamChanges = map[uint64]ParamChange{} // pairId => ParamChange

var (
	marketParamChanges = []struct {
		MakerFeeRate  sdk.Dec
		TakerFeeRate  sdk.Dec
		TargetPairIds []uint64
	}{
		{
			MakerFeeRate: sdk.NewDecWithPrec(5, 4),
			TakerFeeRate: sdk.NewDecWithPrec(1, 3),
			TargetPairIds: []uint64{
				18, // USDC.axl/USDC.grv
				22, // USDC.grv/IST
				23, // USDC.axl/IST
				24, // IST/USDC.grv
				25, // IST/USDC.axl
				28, // CMST/USDC.axl
				35, // CMST/USDC.grv
				52, // USDC.axl/USDT.grv
				53, // USDC.grv/USDT.grv
			},
		},
	}
	poolParamChanges = []struct {
		TickSpacing   uint32
		TargetPairIds []uint64
	}{
		{
			TickSpacing: 10,
			TargetPairIds: []uint64{
				11, // USDC.grv/WETH.grv
				12, // WETH.grv/USDC.grv
				13, // bCRE/USDC.grv
				15, // WETH.axl/USDC.axl
				16, // bCRE/USDC.axl
				19, // ATOM/USDC.axl
				20, // ATOM/USDC.grv
				27, // bCRE/CMST
				32, // stATOM/IST
				33, // bCRE/IST
				39, // WBTC.grv/USDC.grv
			},
		},
		{
			TickSpacing: 5,
			TargetPairIds: []uint64{
				18, // USDC.axl/USDC.grv
				22, // USDC.grv/IST
				23, // USDC.axl/IST
				24, // IST/USDC.grv
				25, // IST/USDC.axl
				28, // CMST/USDC.axl
				35, // CMST/USDC.grv
				37, // stkATOM/ATOM
				42, // stEVMOS/EVMOS
				43, // stATOM/ATOM
				52, // USDC.axl/USDT.grv
				53, // USDC.grv/USDT.grv
			},
		},
	}
)

func init() {
	for _, marketParamChange := range marketParamChanges {
		marketParamChange := marketParamChange // copy
		for _, pairId := range marketParamChange.TargetPairIds {
			change := ParamChanges[pairId]
			change.MakerFeeRate = &marketParamChange.MakerFeeRate
			change.TakerFeeRate = &marketParamChange.TakerFeeRate
			ParamChanges[pairId] = change
		}
	}
	for _, poolParamChange := range poolParamChanges {
		poolParamChange := poolParamChange
		for _, pairId := range poolParamChange.TargetPairIds {
			change := ParamChanges[pairId]
			change.TickSpacing = &poolParamChange.TickSpacing
			ParamChanges[pairId] = change
		}
	}
}
