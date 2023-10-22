package keeper_test

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestPoolOrdersMatching_FindEdgecase() {
	// NOTE: not broken yet

	r := rand.New(rand.NewSource(1))

	for i := 0; i < 5; i++ { // For 5 different random seeds
		s.SetupTest()

		seed := r.Int63()
		r := rand.New(rand.NewSource(seed))

		var initialPoolPrice sdk.Dec
		f := r.Float64()
		switch {
		case f < 0.3:
			initialPoolPrice = randDec(r, utils.ParseDec("0.01"), utils.ParseDec("0.1"))
		case f < 0.7:
			initialPoolPrice = randDec(r, utils.ParseDec("0.9"), utils.ParseDec("1.1"))
		default:
			initialPoolPrice = randDec(r, utils.ParseDec("10"), utils.ParseDec("100"))
		}

		market, pool := s.CreateMarketAndPool("ucre", "uusd", initialPoolPrice)

		lpAddr := s.FundedAccount(1, enoughCoins)
		ordererAddr := s.FundedAccount(2, enoughCoins)

		s.MakeLastPrice(
			market.Id, ordererAddr,
			exchangetypes.PriceAtTick(exchangetypes.TickAtPrice(initialPoolPrice)))

		basicLiquidity := randInt(r, sdk.NewIntWithDecimal(1, 6), sdk.NewIntWithDecimal(1, 8))
		s.AddLiquidityByLiquidity(lpAddr, pool.Id, types.MinPrice, types.MaxPrice, basicLiquidity)
		for j := 0; j < 10; j++ { // Create 10 random positions
			var basePrice sdk.Dec
			f := r.Float64()
			switch {
			case f < 0.3:
				basePrice = randDec(
					r, initialPoolPrice.Mul(utils.ParseDec("0.3")),
					initialPoolPrice.Mul(utils.ParseDec("0.5")))
			case f < 0.7:
				basePrice = randDec(
					r, initialPoolPrice.Mul(utils.ParseDec("0.9")),
					initialPoolPrice.Mul(utils.ParseDec("1.1")))
			default:
				basePrice = randDec(
					r, initialPoolPrice.Mul(utils.ParseDec("1.5")),
					initialPoolPrice.Mul(utils.ParseDec("1.7")))
			}
			spreadFactor := randDec(r, utils.ParseDec("0.01"), utils.ParseDec("0.2"))
			lowerPrice := exchangetypes.PriceAtTick(types.AdjustPriceToTickSpacing(
				basePrice.Mul(utils.OneDec.Sub(spreadFactor)), pool.TickSpacing, false))
			upperPrice := exchangetypes.PriceAtTick(types.AdjustPriceToTickSpacing(
				basePrice.Mul(utils.OneDec.Add(spreadFactor)), pool.TickSpacing, true))
			liquidity := randInt(r, sdk.NewIntWithDecimal(1, 6), sdk.NewIntWithDecimal(1, 8))
			s.AddLiquidityByLiquidity(
				lpAddr, pool.Id, lowerPrice, upperPrice, liquidity)
		}

		for j := 0; j < 300; j++ { // Execute 300 random market orders
			for k := 0; k < 10; k++ { // Execute 10 random market orders per block
				isBuy := r.Float64() < 0.5
				qty := randInt(r, sdk.NewInt(10000), sdk.NewInt(5_000000))

				s.PlaceMarketOrder(market.Id, ordererAddr, isBuy, qty)
			}
			_, broken := keeper.CanRemoveLiquidityInvariant(s.keeper)(s.Ctx)
			s.Require().False(broken)

			s.NextBlock()

			for k := 0; k < 10; k++ { // Execute 10 random market orders per block
				isBuy := r.Float64() < 0.5

				marketState := s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id)
				lastPrice := *marketState.LastPrice
				price := lastPrice
				lower := sdk.MaxInt(sdk.NewInt(10000), sdk.NewInt(15000).ToDec().Quo(lastPrice).TruncateInt())
				qty := randInt(r, lower, sdk.NewInt(5_000000))

				if isBuy {
					price = lastPrice.Mul(utils.OneDec.Add(randDec(r, utils.ParseDec("0.001"), utils.ParseDec("0.09"))))
				} else {
					price = lastPrice.Mul(utils.OneDec.Sub(randDec(r, utils.ParseDec("0.001"), utils.ParseDec("0.09"))))
				}
				price = exchangetypes.RoundPrice(price)
				s.PlaceBatchLimitOrder(market.Id, ordererAddr, isBuy, price, qty, time.Hour)
			}

			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().NoError(s.App.ExchangeKeeper.RunBatchMatching(s.Ctx, market))
		} // End of 300 random market orders
	} // End of 5 different random seeds
}

func randInt(r *rand.Rand, min, max sdk.Int) sdk.Int {
	return min.Add(simtypes.RandomAmount(r, max.Sub(min)))
}

func randDec(r *rand.Rand, min, max sdk.Dec) sdk.Dec {
	return min.Add(simtypes.RandomDecAmount(r, max.Sub(min)))
}
