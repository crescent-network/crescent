package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ParamChange struct {
	MakerFeeRate     *sdk.Dec
	TakerFeeRate     *sdk.Dec
	TickSpacing      *uint32
	MinOrderQuantity *sdk.Dec
	MinOrderQuote    *sdk.Dec
}

var ParamChanges = map[uint64]ParamChange{} // pairId => ParamChange

var (
	feeRateChanges = []struct {
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
	tickSpacingChanges = []struct {
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
	minOrderQtyChanges = []struct {
		MinOrderQuantity sdk.Dec
		TargetPairIds    []uint64
	}{
		{
			MinOrderQuantity: sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 16)),
			TargetPairIds: []uint64{
				8,  // WETH.grv/bCRE
				12, // WETH.grv/USDC.grv
				14, // WETH.axl/bCRE
				15, // WETH.axl/USDC.axl
				17, // WETH.axl/WETH.grv
				26, // INJ/bCRE
				36, // EVMOS/bCRE
				40, // OKT/bCRE
				41, // stEVMOS/bCRE
				42, // stEVMOS/EVMOS
				45, // CANTO/bCRE
			},
		},
		{
			MinOrderQuantity: sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 6)),
			TargetPairIds: []uint64{
				39, // WBTC.grv/USDC.grv
				51, // CRO/bCRE
			},
		},
	}
	minOrderQuoteChanges = []struct {
		MinOrderQuote sdk.Dec
		TargetPairIds []uint64
	}{
		{
			MinOrderQuote: sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 16)),
			TargetPairIds: []uint64{
				7,  // GRAV/WETH.grv
				11, // USDC.grv/WETH.grv
				17, // WETH.axl/WETH.grv
				42, // stEVMOS/EVMOS
			},
		},
	}
)

func init() {
	for _, feeRateChange := range feeRateChanges {
		feeRateChange := feeRateChange // copy
		for _, pairId := range feeRateChange.TargetPairIds {
			change := ParamChanges[pairId]
			change.MakerFeeRate = &feeRateChange.MakerFeeRate
			change.TakerFeeRate = &feeRateChange.TakerFeeRate
			ParamChanges[pairId] = change
		}
	}
	for _, tickSpacingChange := range tickSpacingChanges {
		tickSpacingChange := tickSpacingChange
		for _, pairId := range tickSpacingChange.TargetPairIds {
			change := ParamChanges[pairId]
			change.TickSpacing = &tickSpacingChange.TickSpacing
			ParamChanges[pairId] = change
		}
	}
	for _, minOrderQtyChange := range minOrderQtyChanges {
		minOrderQtyChange := minOrderQtyChange
		for _, pairId := range minOrderQtyChange.TargetPairIds {
			change := ParamChanges[pairId]
			change.MinOrderQuantity = &minOrderQtyChange.MinOrderQuantity
			ParamChanges[pairId] = change
		}
	}
	for _, minOrderQuoteChange := range minOrderQuoteChanges {
		minOrderQuoteChange := minOrderQuoteChange
		for _, pairId := range minOrderQuoteChange.TargetPairIds {
			change := ParamChanges[pairId]
			change.MinOrderQuote = &minOrderQuoteChange.MinOrderQuote
			ParamChanges[pairId] = change
		}
	}
}
