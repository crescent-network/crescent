<!-- order: 6 -->

# Events

The `liquidity` module emits the following events:

## Handlers

### MsgCreatePair

| Type        | Attribute Key    | Attribute Value  |
|-------------|------------------|------------------|
| create_pair | creator          | {creator}        |
| create_pair | base_coin_denom  | {baseCoinDenom}  |
| create_pair | quote_coin_denom | {quoteCoinDenom} |
| create_pair | pair_id          | {pairId}         |
| create_pair | escrow_address   | {escrowAddress}  |
| message     | module           | liquidity        |
| message     | action           | create_pair      |
| message     | sender           | {senderAddress}  |


### MsgCreatePool

| Type        | Attribute Key    | Attribute Value  |
|-------------|------------------|------------------|
| create_pool | creator          | {creator}        |
| create_pool | pair_id          | {pairId}         |
| create_pool | deposit_coins    | {depositCoins}   |
| create_pool | pool_id          | {poolId}         |
| create_pool | reserve_address  | {reserveAddress} |
| create_pool | minted_pool_coin | {poolCoin}       |
| message     | module           | liquidity        |
| message     | action           | create_pool      |
| message     | sender           | {senderAddress}  |

### MsgDeposit

| Type      | Attribute Key | Attribute Value |
|-----------|---------------|-----------------|
| deposit   | depositor     | {creator}       |
| deposit   | pool_id       | {poolId}        |
| deposit   | deposit_coins | {depositCoins}  |
| deposit   | request_id    | {reqId}         |
| message   | module        | liquidity       |
| message   | action        | deposit         |
| message   | sender        | {senderAddress} |

### MsgWithdraw

| Type      | Attribute Key | Attribute Value |
|-----------|---------------|-----------------|
| withdraw  | withdrawer    | {creator}       |
| withdraw  | pool_id       | {poolId}        |
| withdraw  | pool_coin     | {poolCoin}      |
| withdraw  | request_id    | {reqId}         |
| message   | module        | liquidity       |
| message   | action        | withdraw        |
| message   | sender        | {senderAddress} |

### MsgLimitOrder

| Type        | Attribute Key     | Attribute Value   |
|-------------|-------------------|-------------------|
| limit_order | orderer           | {orderer}         |
| limit_order | pair_id           | {pairId}          |
| limit_order | swap_direction    | {direction}       |
| limit_order | offer_coin        | {offerCoin}       |
| limit_order | demand_coin_denom | {demandCoinDenom} |
| limit_order | price             | {price}           |
| limit_order | amount            | {amount}          |
| limit_order | request_id        | {reqId}           |
| limit_order | batch_id          | {batchId}         |
| limit_order | expire_at         | {expireAt}        |
| limit_order | refunded_coin     | {refundedCoin}    |
| message     | module            | liquidity         |
| message     | action            | limit_order       |
| message     | sender            | {senderAddress}   |

### MsgMarketOrder

| Type         | Attribute Key     | Attribute Value   |
|--------------|-------------------|-------------------|
| market_order | request_id        | {reqId}           |
| market_order | orderer           | {orderer}         |
| market_order | pair_id           | {pairId}          |
| market_order | swap_direction    | {direction}       |
| market_order | offer_coin        | {offerCoin}       |
| market_order | demand_coin_denom | {demandCoinDenom} |
| market_order | price             | {price}           |
| market_order | amount            | {amount}          |
| market_order | batch_id          | {batchId}         |
| market_order | expire_at         | {expireAt}        |
| message      | module            | liquidity         |
| message      | action            | market_order      |
| message      | sender            | {senderAddress}   |

### MsgCancelOrder

| Type         | Attribute Key    | Attribute Value |
|--------------|------------------|-----------------|
| cancel_order | orderer          | {orderer}       |
| cancel_order | pair_id          | {pairId}        |
| cancel_order | swap_request_id  | {swapRequestId} |
| message      | module           | liquidity       |
| message      | action           | cancel_order    |
| message      | sender           | {senderAddress} |