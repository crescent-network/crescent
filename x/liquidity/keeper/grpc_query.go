package keeper

import (
	"github.com/tendermint/farming/x/liquidity/types"
)

var _ types.QueryServer = Keeper{}
