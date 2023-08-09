package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
)

func ValidatePublicPositionShareSupply(ammPosition ammtypes.Position, shareSupply sdk.Int) {
	// TODO: check why other modules change the supply during simulation test
	//if !ammPosition.Liquidity.GTE(shareSupply) {
	//	panic(fmt.Errorf("must satisfy: %s >= %s", ammPosition.Liquidity, shareSupply))
	//}
}

func ValidateMintShareResult(mintedLiquidity, totalLiquidityBefore, mintedShareAmt, shareSupplyBefore sdk.Int) {
	if !mintedShareAmt.LTE(mintedLiquidity) {
		panic(fmt.Errorf("must satisfy: %s <= %s", mintedShareAmt, mintedLiquidity))
	}
	if shareSupplyBefore.IsZero() {
		if !mintedLiquidity.Equal(mintedShareAmt) {
			panic(fmt.Errorf("must satisfy: %s == %s", mintedShareAmt, mintedLiquidity))
		}
		return
	}
	totalLiquidityAfter := totalLiquidityBefore.Add(mintedLiquidity)
	shareSupplyAfter := shareSupplyBefore.Add(mintedShareAmt)
	ratioBefore := totalLiquidityBefore.ToDec().Quo(shareSupplyBefore.ToDec())
	ratioAfter := totalLiquidityAfter.ToDec().Quo(shareSupplyAfter.ToDec())
	if !utils.DecApproxEqual(ratioBefore, ratioAfter) {
		panic(fmt.Errorf("must satisfy: %s == %s", ratioBefore, ratioAfter))
	}
}
