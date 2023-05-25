package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestTickBytes(t *testing.T) {
	for tick := int32(-100); tick <= 100; tick++ {
		bz := types.TickToBytes(tick)
		tick2 := types.BytesToTick(bz)
		require.Equal(t, tick, tick2)
	}
}
