package types_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func TestPoolKey(t *testing.T) {
	require.Equal(t, []byte{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40}, types.GetPoolKey(1000000))
}

func TestPoolStateKey(t *testing.T) {
	require.Equal(t, []byte{0x43, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40}, types.GetPoolStateKey(1000000))
}

func TestPoolByReserveAddressIndexKey(t *testing.T) {
	reserveAddr := types.DerivePoolReserveAddress(1000000)
	require.Equal(t, []byte{
		0x44, 0x88, 0xb4, 0xd1, 0xe0, 0x9, 0x51, 0x86, 0x34, 0x32, 0x2a, 0x4f, 0xac, 0xdf, 0x83, 0x1e, 0x60, 0x7e, 0xda,
		0x92, 0x11, 0xfe, 0x3f, 0xeb, 0xc7, 0x3e, 0xc8, 0xb8, 0xb6, 0x5, 0x64, 0xf1, 0x7,
	}, types.GetPoolByReserveAddressIndexKey(reserveAddr))
}

func TestPoolByMarketIndexKey(t *testing.T) {
	require.Equal(t, []byte{0x45, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40}, types.GetPoolByMarketIndexKey(1000000))
}

func TestPositionKey(t *testing.T) {
	require.Equal(t, []byte{0x46, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40}, types.GetPositionKey(1000000))
}

func TestPositionByParamsIndexKey(t *testing.T) {
	ownerAddr := utils.TestAddress(10000000)
	key := types.GetPositionByParamsIndexKey(ownerAddr, 1000000, -10000, 20000)
	require.Equal(t, []byte{
		0x47, 0x14, 0x80, 0xda, 0xc4, 0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40, 0x0, 0xff, 0xff, 0xd8, 0xf0, 0x1, 0x0, 0x0, 0x4e, 0x20,
	}, key)
	prefix := types.GetPositionsByOwnerIteratorPrefix(ownerAddr)
	require.True(t, bytes.HasPrefix(key, prefix))
	prefix = types.GetPositionsByOwnerAndPoolIteratorPrefix(ownerAddr, 1000000)
	require.True(t, bytes.HasPrefix(key, prefix))
}

func TestPositionsByPoolIndexKey(t *testing.T) {
	key := types.GetPositionsByPoolIndexKey(1000000, 2000000)
	require.Equal(t, []byte{0x48, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1e, 0x84, 0x80}, key)
	poolId, positionId := types.ParsePositionsByPoolIndexKey(key)
	require.EqualValues(t, 1000000, poolId)
	require.EqualValues(t, 2000000, positionId)
	prefix := types.GetPositionsByPoolIteratorPrefix(poolId)
	require.True(t, bytes.HasPrefix(key, prefix))
}

func TestTickInfoKey(t *testing.T) {
	key1 := types.GetTickInfoKey(1000000, -50000)
	require.Equal(t, []byte{0x49, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40, 0x0, 0xff, 0xff, 0x3c, 0xb0}, key1)
	key2 := types.GetTickInfoKey(1000000, 50000)
	require.Equal(t, []byte{0x49, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40, 0x1, 0x0, 0x0, 0xc3, 0x50}, key2)
	prefix := types.GetTickInfosByPoolIteratorPrefix(1000000)
	require.True(t, bytes.HasPrefix(key1, prefix))
	require.True(t, bytes.HasPrefix(key2, prefix))
	poolId, tick := types.ParseTickInfoKey(key1)
	require.Equal(t, uint64(1000000), poolId)
	require.Equal(t, int32(-50000), tick)
	poolId, tick = types.ParseTickInfoKey(key2)
	require.Equal(t, uint64(1000000), poolId)
	require.Equal(t, int32(50000), tick)
}

func TestFarmingPlanKey(t *testing.T) {
	require.Equal(t, []byte{0x4b, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40}, types.GetFarmingPlanKey(1000000))
}

func TestTickBytes(t *testing.T) {
	for tick := int32(-100); tick <= 100; tick++ {
		bz := types.TickToBytes(tick)
		tick2 := types.BytesToTick(bz)
		require.Equal(t, tick, tick2)
	}
}
