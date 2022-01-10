package keeper

import (
	"github.com/crescent-network/crescent/x/liquidity/types"
)

var _ types.QueryServer = Keeper{}
