package types

// Event types for the module.
const (
	EventTypeLiquidFarm              = "liquid_farm"
	EventTypeLiquidUnfarm            = "liquid_unfarm"
	EventTypeLiquidUnfarmAndWithdraw = "liquid_unfarm_and_withdraw"
	EventTypePlaceBid                = "place_bid"
	EventTypeRefundBid               = "refund_bid"

	AttributeKeyPoolId                   = "pool_id"
	AttributeKeyAuctionId                = "auction_id"
	AttributeKeyBidId                    = "bid_id"
	AttributeKeyFarmer                   = "farmer"
	AttributeKeyLiquidFarmReserveAddress = "liquid_farm_reserve_address"
	AttributeKeyBidder                   = "bidder"
	AttributeKeyFarmingCoin              = "farming_coin"
	AttributeKeyMintedCoin               = "minted_coin"
	AttributeKeyBiddingCoin              = "bidding_coin"
	AttributeKeyUnfarmingCoin            = "unfarming_coin"
	AttributeKeyUnfarmedCoin             = "unfarmed_coin"
	AttributeKeyRefundCoin               = "refund_coin"
)
