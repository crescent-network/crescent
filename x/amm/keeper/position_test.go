package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (s *KeeperTestSuite) TestAddLiquidity() {
	senderAddr := utils.TestAddress(1)
	pool, err := s.k.CreatePool(s.ctx, senderAddr, "ucre", "uusd", 1)
	s.Require().NoError(err)
	fmt.Println(pool)
	position, liquidity, amt0, amt1, err := s.k.AddLiquidity(
		s.ctx, senderAddr, 1, -20000, 2500,
		sdk.NewInt(1000000), sdk.NewInt(1000000), sdk.NewInt(10000), sdk.NewInt(10000))
	s.Require().NoError(err)
	fmt.Println(position, liquidity, amt0, amt1)

	_, amt0, amt1, err = s.k.RemoveLiquidity(
		s.ctx, senderAddr, 1, sdk.NewInt(9472135), sdk.ZeroInt(), sdk.ZeroInt())
	s.Require().NoError(err)
	fmt.Println(amt0, amt1)
}
