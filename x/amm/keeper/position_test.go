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

	position, liquidity, amt := s.AddLiquidity(
		senderAddr, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		utils.ParseCoins("1000000ucre,1000000uusd"))
	fmt.Println(position, liquidity, amt)

	_, amt = s.RemoveLiquidity(
		senderAddr, position.Id, sdk.NewInt(9472135))
	fmt.Println(amt)
}
