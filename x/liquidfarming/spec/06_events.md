<!-- order: 6 -->

# Events

The `liquidfarming` module emits the following events:

## Handlers

### MsgFarm

| Type       | Attribute Key      | Attribute Value        |
| ---------- | ------------------ | ---------------------- |
| farm       | pool_id            | {poolId}               |
| farm       | farmer             | {farmer}               |
| farm       | farming_coin       | {farmingCoin}          |
| farm       | farm_coin          | {farmCoin}             |
| message    | module             | liquidfarming          |
| message    | action             | farm                   |
| message    | farmer             | {farmerAddress}        |

### MsgUnfarm

| Type       | Attribute Key      | Attribute Value        |
| ---------- | ------------------ | ---------------------- |
| unfarm     | pool_id            | {poolId}               |
| unfarm     | farmer             | {farmer}               |
| unfarm     | unfarming_coin     | {unfarmingCoin}        |
| unfarm     | unfarm_coin        | {unfarmCoin}           |
| message    | module             | liquidfarming          |
| message    | action             | unfarm                 |
| message    | farmer             | {farmerAddress}        |

### MsgUnfarmAndWithdraw

| Type       | Attribute Key      | Attribute Value        |
| ---------- | ------------------ | ---------------------- |
| unfarm     | pool_id            | {poolId}               |
| unfarm     | farmer             | {farmer}               |
| unfarm     | unfarming_coin     | {unfarmingCoin}        |
| message    | module             | liquidfarming          |
| message    | action             | unfarmandwithdraw      |
| message    | farmer             | {farmerAddress}        |

### MsgPlaceBid

| Type       | Attribute Key      | Attribute Value        |
| ---------- | ------------------ | ---------------------- |
| place_bid  | auction_id         | {auctionId}            |
| place_bid  | bidder             | {bidder}               |
| place_bid  | amount             | {amount}               |
| message    | module             | liquidfarming          |
| message    | action             | deposit                |
| message    | bidder             | {bidderAddress}        |

### MsgRefundBid 

| Type       | Attribute Key      | Attribute Value        |
| ---------- | ------------------ | ---------------------- |
