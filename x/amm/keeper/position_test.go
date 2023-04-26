package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (s *KeeperTestSuite) TestAddLiquidity() {
	senderAddr := utils.TestAddress(1)

	s.FundAccount(senderAddr, utils.ParseCoins("10000000ucre,10000000uusd"))

	market := s.CreateMarket(senderAddr, "ucre", "uusd", true)
	pool := s.CreatePool(senderAddr, market.Id, sdk.NewDec(1), true)
	fmt.Println(pool)

	position, liquidity, amt0, amt1 := s.AddLiquidity(
		senderAddr, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		sdk.NewInt(1000000), sdk.NewInt(1000000), sdk.NewInt(10000), sdk.NewInt(10000))
	fmt.Println(position, liquidity, amt0, amt1)

	_, amt0, amt1 = s.RemoveLiquidity(
		senderAddr, position.Id, sdk.NewDec(9472135), sdk.ZeroInt(), sdk.ZeroInt())
	fmt.Println(amt0, amt1)
}
