<!-- order: 6 -->

# Events

The `liquidfarming` module emits the following events:

## Handlers

### MsgLiquidFarm

| Type        | Attribute Key | Attribute Value |
| ----------- | ------------- | --------------- |
| liquid_farm | pool_id       | {poolId}        |
| liquid_farm | farmer        | {farmer}        |
| liquid_farm | farming_coin  | {farmingCoin}   |
| liquid_farm | minted_coin   | {mintingCoin}   |
| message     | module        | liquidfarming   |
| message     | action        | farm            |
| message     | farmer        | {farmerAddress} |

### MsgLiquidUnfarm

| Type          | Attribute Key  | Attribute Value |
| ------------- | -------------- | --------------- |
| liquid_unfarm | pool_id        | {poolId}        |
| liquid_unfarm | farmer         | {farmer}        |
| liquid_unfarm | unfarming_coin | {unfarmingCoin} |
| liquid_unfarm | unfarmed_coin  | {unfarmedCoin}  |
| message       | module         | liquidfarming   |
| message       | action         | unfarm          |
| message       | farmer         | {farmerAddress} |

### MsgLiquidUnfarmAndWithdraw

| Type                       | Attribute Key  | Attribute Value   |
| -------------------------- | -------------- | ----------------- |
| liquid_unfarm_and_withdraw | pool_id        | {poolId}          |
| liquid_unfarm_and_withdraw | farmer         | {farmer}          |
| liquid_unfarm_and_withdraw | unfarming_coin | {unfarmingCoin}   |
| liquid_unfarm_and_withdraw | unfarmed_coin  | {unfarmedCoin}    |
| message                    | module         | liquidfarming     |
| message                    | action         | unfarmandwithdraw |
| message                    | farmer         | {farmerAddress}   |

### MsgPlaceBid

| Type      | Attribute Key | Attribute Value |
| --------- | ------------- | --------------- |
| place_bid | auction_id    | {auctionId}     |
| place_bid | bidder        | {bidder}        |
| place_bid | bidding_coin  | {biddingCoin}   |
| message   | module        | liquidfarming   |
| message   | action        | deposit         |
| message   | bidder        | {bidderAddress} |

### MsgRefundBid

| Type       | Attribute Key | Attribute Value |
| ---------- | ------------- | --------------- |
| refund_bid | auction_id    | {auctionId}     |
| refund_bid | bidder        | {bidder}        |
| refund_bid | refund_coin   | {bidAmount}     |
| message    | module        | liquidfarming   |
| message    | action        | deposit         |
| message    | bidder        | {bidderAddress} |
