package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestSimulation() {
	minPrice, maxPrice := utils.ParseDec("0.000001"), utils.ParseDec("100000")
	enoughCoins := utils.ParseCoins("1000000000000000000000000000000ucre,1000000000000000000000000000000uusd")
	for seed := int64(1); seed <= 5; seed++ {
		s.SetupTest()

		r := rand.New(rand.NewSource(seed))
		var initialPrice sdk.Dec
		v := r.Float64()
		switch {
		case v <= 0.3:
			initialPrice = utils.RandomDec(r, minPrice, utils.ParseDec("0.000001"))
		case v <= 0.7:
			initialPrice = utils.RandomDec(r, utils.ParseDec("0.05"), utils.ParseDec("500"))
		default:
			initialPrice = utils.RandomDec(r, utils.ParseDec("100000"), maxPrice)
		}
		market, pool := s.CreateMarketAndPool("ucre", "uusd", initialPrice)

		lpAddrs := make([]sdk.AccAddress, 10)
		for i := range lpAddrs {
			lpAddrs[i] = s.FundedAccount(1+i, enoughCoins)
		}
		ordererAddr := s.FundedAccount(100, enoughCoins)

		for i := 0; i < 100; i++ {
			poolState := s.keeper.MustGetPoolState(s.Ctx, pool.Id)

			// Randomly add liquidity
			for _, lpAddr := range lpAddrs {
				if r.Float64() <= 0.3 {
					var lowerPrice, upperPrice sdk.Dec
					var desiredAmt sdk.Coins
					v := r.Float64()
					switch {
					case v <= 0.3: // lowerPrice < upperPrice <= poolPrice
						lowerPrice = utils.RandomDec(r, minPrice, poolState.CurrentPrice.Mul(utils.ParseDec("0.8")))
						upperPrice = utils.RandomDec(r, lowerPrice.Mul(utils.ParseDec("1.001")), poolState.CurrentPrice)
					case v <= 0.7: // lowerPrice < poolPrice <= upperPrice
						lowerPrice = utils.RandomDec(r, minPrice, poolState.CurrentPrice)
						upperPrice = utils.RandomDec(r, poolState.CurrentPrice, maxPrice)
					default: // poolPrice <= lowerPrice < upperPrice
						lowerPrice = utils.RandomDec(r, poolState.CurrentPrice, maxPrice.Mul(utils.ParseDec("0.8")))
						upperPrice = utils.RandomDec(r, lowerPrice.Mul(utils.ParseDec("1.001")), maxPrice)
					}
					lowerPrice = types.AdjustPriceToTickSpacing(lowerPrice, pool.TickSpacing, false)
					upperPrice = types.AdjustPriceToTickSpacing(upperPrice, pool.TickSpacing, true)
					if upperPrice.LTE(poolState.CurrentPrice) {
						desiredAmt = sdk.NewCoins(sdk.NewCoin("uusd", utils.RandomInt(r, sdk.NewInt(10000), sdk.NewInt(1000_000000))))
					} else if lowerPrice.GTE(poolState.CurrentPrice) {
						desiredAmt = sdk.NewCoins(sdk.NewCoin("ucre", utils.RandomInt(r, sdk.NewInt(10000), sdk.NewInt(1000_000000))))
					} else {
						desiredAmt = utils.ParseCoins("1000_000000ucre,1000_000000uusd")
					}
					s.AddLiquidity(lpAddr, pool.Id, lowerPrice, upperPrice, desiredAmt)
				}
			}

			// Randomly remove liquidity
			for _, lpAddr := range lpAddrs {
				if r.Float64() <= 0.2 {
					var positions []types.Position
					s.keeper.IteratePositionsByOwner(s.Ctx, lpAddr, func(position types.Position) (stop bool) {
						positions = append(positions, position)
						return false
					})
					if len(positions) == 0 {
						continue
					}
					position := positions[r.Intn(len(positions))]
					liquidity := utils.RandomInt(r, utils.ZeroInt, position.Liquidity).Add(sdk.NewInt(1))
					s.RemoveLiquidity(lpAddr, position.Id, liquidity)
				}
			}

			// Randomly place market orders
			isBuy := r.Float64() <= 0.5
			qty := utils.RandomInt(r, sdk.NewInt(100), sdk.NewInt(10_000000))
			s.PlaceMarketOrder(market.Id, ordererAddr, isBuy, qty)
		}
	}
}
