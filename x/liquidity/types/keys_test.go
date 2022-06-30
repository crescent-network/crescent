package types_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

type keysTestSuite struct {
	suite.Suite
}

func TestKeysTestSuite(t *testing.T) {
	suite.Run(t, new(keysTestSuite))
}

func (s *keysTestSuite) TestGetPairKey() {
	s.Require().Equal([]byte{0xa5, 0, 0, 0, 0, 0, 0, 0, 0}, types.GetPairKey(0))
	s.Require().Equal([]byte{0xa5, 0, 0, 0, 0, 0, 0, 0, 0x9}, types.GetPairKey(9))
	s.Require().Equal([]byte{0xa5, 0, 0, 0, 0, 0, 0, 0, 0xa}, types.GetPairKey(10))
}

func (s *keysTestSuite) TestGetPairIndexKey() {
	s.Require().Equal([]byte{0xa6, 0x6, 0x64, 0x65, 0x6e, 0x6f, 0x6d, 0x31, 0x6, 0x64, 0x65, 0x6e, 0x6f, 0x6d, 0x32}, types.GetPairIndexKey("denom1", "denom2"))
	s.Require().Equal([]byte{0xa6, 0x6, 0x64, 0x65, 0x6e, 0x6f, 0x6d, 0x33, 0x6, 0x64, 0x65, 0x6e, 0x6f, 0x6d, 0x34}, types.GetPairIndexKey("denom3", "denom4"))
}

func (s *keysTestSuite) TestPairsByDenomsIndexKey() {
	testCases := []struct {
		denomA   string
		denomB   string
		pairId   uint64
		expected []byte
	}{
		{
			"denomA",
			"denomB",
			1,
			[]byte{0xa7, 0x6, 0x64, 0x65, 0x6e, 0x6f, 0x6d, 0x41, 0x6, 0x64,
				0x65, 0x6e, 0x6f, 0x6d, 0x42, 0, 0, 0, 0, 0, 0, 0, 0x1},
		},
		{
			"denomC",
			"denomD",
			20,
			[]byte{0xa7, 0x6, 0x64, 0x65, 0x6e, 0x6f, 0x6d, 0x43, 0x6, 0x64,
				0x65, 0x6e, 0x6f, 0x6d, 0x44, 0, 0, 0, 0, 0, 0, 0, 0x14},
		},
		{
			"denomE",
			"denomF",
			13,
			[]byte{0xa7, 0x6, 0x64, 0x65, 0x6e, 0x6f, 0x6d, 0x45, 0x6, 0x64,
				0x65, 0x6e, 0x6f, 0x6d, 0x46, 0, 0, 0, 0, 0, 0, 0, 0xd},
		},
	}

	for _, tc := range testCases {
		key := types.GetPairsByDenomsIndexKey(tc.denomA, tc.denomB, tc.pairId)
		s.Require().Equal(tc.expected, key)

		s.Require().True(bytes.HasPrefix(key, types.GetPairsByDenomsIndexKeyPrefix(tc.denomA, tc.denomB)))

		denomA, denomB, pairId := types.ParsePairsByDenomsIndexKey(key)
		s.Require().Equal(tc.denomA, denomA)
		s.Require().Equal(tc.denomB, denomB)
		s.Require().Equal(tc.pairId, pairId)
	}
}

func (s *keysTestSuite) TestGetPoolKey() {
	s.Require().Equal([]byte{0xab, 0, 0, 0, 0, 0, 0, 0, 0x1}, types.GetPoolKey(1))
	s.Require().Equal([]byte{0xab, 0, 0, 0, 0, 0, 0, 0, 0x5}, types.GetPoolKey(5))
	s.Require().Equal([]byte{0xab, 0, 0, 0, 0, 0, 0, 0, 0xa}, types.GetPoolKey(10))
}

func (s *keysTestSuite) TestGetPoolByReserveAddressIndexKey() {
	reserveAddr1 := types.PoolReserveAddress(1)
	reserveAddr2 := types.PoolReserveAddress(2)
	reserveAddr3 := types.PoolReserveAddress(3)
	s.Require().Equal([]byte{0xac, 0x20, 0x8d, 0x23, 0xde, 0x40, 0x5e, 0x99, 0xfa, 0x38, 0x11,
		0x3a, 0x68, 0x5f, 0xb0, 0x79, 0xc, 0x95, 0x46, 0x45, 0x61, 0x57, 0x5a, 0x8f, 0x5b, 0x8,
		0x63, 0x4a, 0xd5, 0xb3, 0x78, 0x6d, 0x62, 0x67}, types.GetPoolByReserveAddressIndexKey(reserveAddr1))
	s.Require().Equal([]byte{0xac, 0x20, 0xe9, 0xfb, 0x4b, 0x2f, 0xa8, 0x8, 0xe3, 0x41, 0x46,
		0x11, 0x9d, 0x87, 0x62, 0x49, 0x92, 0x96, 0x69, 0x65, 0xc0, 0x9c, 0xbd, 0x41, 0x8, 0x24,
		0xb2, 0x26, 0xf3, 0x2d, 0x4e, 0xf4, 0x3b, 0x5c}, types.GetPoolByReserveAddressIndexKey(reserveAddr2))
	s.Require().Equal([]byte{0xac, 0x20, 0xb9, 0xaa, 0x33, 0x5a, 0xe2, 0x97, 0x9a, 0x24, 0x7c,
		0xa2, 0xbc, 0xde, 0xb0, 0x19, 0x44, 0x5f, 0x24, 0x5f, 0xd3, 0x40, 0x99, 0x92, 0x6a, 0x96,
		0xb0, 0x42, 0x8f, 0x2e, 0x76, 0xe5, 0x3c, 0x11}, types.GetPoolByReserveAddressIndexKey(reserveAddr3))
}

func (s *keysTestSuite) TestPoolsByPairIndexKey() {
	testCases := []struct {
		pairId   uint64
		poolId   uint64
		expected []byte
	}{
		{
			5,
			10,
			[]byte{0xad, 0, 0, 0, 0, 0, 0, 0, 0x5, 0, 0, 0, 0, 0, 0, 0, 0xa},
		},
		{
			2,
			7,
			[]byte{0xad, 0, 0, 0, 0, 0, 0, 0, 0x2, 0, 0, 0, 0, 0, 0, 0, 0x7},
		},
		{
			3,
			5,
			[]byte{0xad, 0, 0, 0, 0, 0, 0, 0, 0x3, 0, 0, 0, 0, 0, 0, 0, 0x5},
		},
	}

	for _, tc := range testCases {
		key := types.GetPoolsByPairIndexKey(tc.pairId, tc.poolId)
		s.Require().Equal(tc.expected, key)

		s.Require().True(bytes.HasPrefix(key, types.GetPoolsByPairIndexKeyPrefix(tc.pairId)))

		poolId := types.ParsePoolsByPairIndexKey(key)
		s.Require().Equal(tc.poolId, poolId)
	}
}

func (s *keysTestSuite) TestGetDepositRequestKey() {
	s.Require().Equal([]byte{0xb0, 0, 0, 0, 0, 0, 0, 0, 0x1, 0, 0,
		0, 0, 0, 0, 0, 0x1}, types.GetDepositRequestKey(1, 1))
	s.Require().Equal([]byte{0xb0, 0, 0, 0, 0, 0, 0, 0x3, 0xe8, 0,
		0, 0, 0, 0, 0, 0x3, 0xe9}, types.GetDepositRequestKey(1000, 1001))
}

func (s *keysTestSuite) TestDepositRequestIndexKey() {
	depositor := sdk.AccAddress(crypto.AddressHash([]byte("depositor")))
	key := types.GetDepositRequestIndexKey(depositor, 1, 2)
	s.Require().Equal([]byte{0xb4, 0x14, 0x9a, 0x69, 0x97, 0x1f, 0x1d, 0xb2, 0xe1, 0xd8, 0x77,
		0x73, 0x6f, 0x7d, 0x36, 0x96, 0x90, 0xa3, 0xbf, 0x57, 0xcf, 0x22, 0, 0, 0, 0,
		0, 0, 0, 0x1, 0, 0, 0, 0, 0, 0, 0, 0x2}, key)
	s.Require().True(bytes.HasPrefix(key, types.GetDepositRequestIndexKeyPrefix(depositor)))
	depositor2, poolId, reqId := types.ParseDepositRequestIndexKey(key)
	s.Require().Equal(depositor, depositor2)
	s.Require().Equal(uint64(1), poolId)
	s.Require().Equal(uint64(2), reqId)
}

func (s *keysTestSuite) TestGetWithdrawRequestKey() {
	s.Require().Equal([]byte{0xb1, 0, 0, 0, 0, 0, 0, 0, 0x1, 0, 0,
		0, 0, 0, 0, 0, 0x1}, types.GetWithdrawRequestKey(1, 1))
	s.Require().Equal([]byte{0xb1, 0, 0, 0, 0, 0, 0, 0x3, 0xe8, 0,
		0, 0, 0, 0, 0, 0x3, 0xe9}, types.GetWithdrawRequestKey(1000, 1001))
}

func (s *keysTestSuite) TestWithdrawRequestIndexKey() {
	withdrawer := sdk.AccAddress(crypto.AddressHash([]byte("withdrawer")))
	key := types.GetWithdrawRequestIndexKey(withdrawer, 1, 2)
	s.Require().Equal([]byte{0xb5, 0x14, 0x19, 0xcd, 0x70, 0x1f, 0x44, 0xf1, 0xed, 0xe, 0x3,
		0xa7, 0xf3, 0xf8, 0x7c, 0xff, 0x84, 0x79, 0x58, 0xc6, 0x56, 0xc2, 0, 0, 0, 0,
		0, 0, 0, 0x1, 0, 0, 0, 0, 0, 0, 0, 0x2}, key)
	s.Require().True(bytes.HasPrefix(key, types.GetWithdrawRequestIndexKeyPrefix(withdrawer)))
	withdrawer2, poolId, reqId := types.ParseWithdrawRequestIndexKey(key)
	s.Require().Equal(withdrawer, withdrawer2)
	s.Require().Equal(uint64(1), poolId)
	s.Require().Equal(uint64(2), reqId)
}

func (s *keysTestSuite) TestGetOrderKey() {
	s.Require().Equal([]byte{0xb2, 0, 0, 0, 0, 0, 0, 0, 0x1, 0, 0,
		0, 0, 0, 0, 0, 0x1}, types.GetOrderKey(1, 1))
	s.Require().Equal([]byte{0xb2, 0, 0, 0, 0, 0, 0, 0x3, 0xe8, 0,
		0, 0, 0, 0, 0, 0x3, 0xe9}, types.GetOrderKey(1000, 1001))
}

func (s *keysTestSuite) TestGetOrdersByPairKeyPrefix() {
	s.Require().Equal([]byte{0xb2, 0, 0, 0, 0, 0, 0, 0, 0x1}, types.GetOrdersByPairKeyPrefix(1))
	s.Require().Equal([]byte{0xb2, 0, 0, 0, 0, 0, 0, 0x3, 0xe8}, types.GetOrdersByPairKeyPrefix(1000))
}

func (s *keysTestSuite) TestOrderIndexKey() {
	orderer := sdk.AccAddress(crypto.AddressHash([]byte("orderer")))
	key := types.GetOrderIndexKey(orderer, 1, 1)
	s.Require().Equal([]byte{0xb3, 0x14, 0x54, 0x7e, 0xfe, 0x47, 0x8f, 0xc9, 0xf9, 0x52, 0xb2,
		0x5c, 0xbc, 0x50, 0xf2, 0x85, 0xf7, 0x7d, 0xff, 0x52, 0x9f, 0x25, 0, 0, 0, 0,
		0, 0, 0, 0x1, 0, 0, 0, 0, 0, 0, 0, 0x1}, key)
	s.Require().True(bytes.HasPrefix(key, types.GetOrderIndexKeyPrefix(orderer)))
	orderer2, pairId, orderId := types.ParseOrderIndexKey(key)
	s.Require().Equal(orderer, orderer2)
	s.Require().Equal(uint64(1), pairId)
	s.Require().Equal(uint64(1), orderId)
}
