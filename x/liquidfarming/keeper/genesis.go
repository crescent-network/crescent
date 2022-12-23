package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/liquidfarming/types"
)

// InitGenesis initializes the capability module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	for _, liquidFarm := range genState.LiquidFarms {
		k.SetLiquidFarm(ctx, liquidFarm)
	}

	for _, record := range genState.LastRewardsAuctionIdRecord {
		k.SetLastRewardsAuctionId(ctx, record.AuctionId, record.PoolId)
	}

	for _, auction := range genState.RewardsAuctions {
		k.SetRewardsAuction(ctx, auction)
	}

	for _, bid := range genState.Bids {
		k.SetBid(ctx, bid)
	}

	for _, record := range genState.WinningBidRecords {
		k.SetWinningBid(ctx, record.AuctionId, record.WinningBid)
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	// Initialize objects to prevent from having nil slice
	if params.LiquidFarms == nil {
		params.LiquidFarms = []types.LiquidFarm{}
	}

	poolIds := []uint64{}
	for _, liquidFarm := range params.LiquidFarms {
		poolIds = append(poolIds, liquidFarm.PoolId)
	}

	lastRewardsAuctionIdRecords := []types.LastRewardsAuctionIdRecord{}
	bids := []types.Bid{}
	winningBidRecords := []types.WinningBidRecord{}
	for _, poolId := range poolIds {
		lastRewardsAuctionIdRecords = append(lastRewardsAuctionIdRecords, types.LastRewardsAuctionIdRecord{
			PoolId:    poolId,
			AuctionId: k.GetLastRewardsAuctionId(ctx, poolId),
		})

		bids = append(bids, k.GetBidsByPoolId(ctx, poolId)...)

		auctionId := k.GetLastRewardsAuctionId(ctx, poolId)
		winningBid, found := k.GetWinningBid(ctx, auctionId, poolId)
		if found {
			winningBidRecords = append(winningBidRecords, types.WinningBidRecord{
				AuctionId:  auctionId,
				WinningBid: winningBid,
			})
		}
	}

	var endTime *time.Time
	tempEndTime, found := k.GetLastRewardsAuctionEndTime(ctx)
	if found {
		endTime = &tempEndTime
	}

	return &types.GenesisState{
		Params:                     params,
		LastRewardsAuctionIdRecord: lastRewardsAuctionIdRecords,
		LiquidFarms:                k.GetLiquidFarmsInStore(ctx),
		RewardsAuctions:            k.GetAllRewardsAuctions(ctx),
		Bids:                       bids,
		WinningBidRecords:          winningBidRecords,
		LastRewardsAuctionEndTime:  endTime,
	}
}
