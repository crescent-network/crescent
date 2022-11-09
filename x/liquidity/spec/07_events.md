<!-- order: 7 -->

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

### MsgCreateRangedPool

| Type               | Attribute Key    | Attribute Value    |
|--------------------|------------------|--------------------|
| create_ranged_pool | creator          | {creator}          |
| create_ranged_pool | pair_id          | {pairId}           |
| create_ranged_pool | deposit_coins    | {depositCoins}     |
| create_ranged_pool | pool_id          | {poolId}           |
| create_ranged_pool | reserve_address  | {reserveAddress}   |
| create_ranged_pool | minted_pool_coin | {poolCoin}         |
| message            | module           | liquidity          |
| message            | action           | create_ranged_pool |
| message            | sender           | {senderAddress}    |

### MsgDeposit

| Type      | Attribute Key | Attribute Value |
|-----------|---------------|-----------------|
| deposit   | depositor     | {depositor}     |
| deposit   | pool_id       | {poolId}        |
| deposit   | deposit_coins | {depositCoins}  |
| deposit   | request_id    | {reqId}         |
| message   | module        | liquidity       |
| message   | action        | deposit         |
| message   | sender        | {senderAddress} |

### MsgWithdraw

| Type      | Attribute Key | Attribute Value |
|-----------|---------------|-----------------|
| withdraw  | withdrawer    | {withdrawer}    |
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
| limit_order | order_direction   | {direction}       |
| limit_order | offer_coin        | {offerCoin}       |
| limit_order | demand_coin_denom | {demandCoinDenom} |
| limit_order | price             | {price}           |
| limit_order | amount            | {amount}          |
| limit_order | order_id          | {orderId}         |
| limit_order | batch_id          | {batchId}         |
| limit_order | expire_at         | {expireAt}        |
| limit_order | refunded_coins    | {refundedCoins}   |
| message     | module            | liquidity         |
| message     | action            | limit_order       |
| message     | sender            | {senderAddress}   |

### MsgMarketOrder

| Type         | Attribute Key     | Attribute Value   |
|--------------|-------------------|-------------------|
| market_order | orderer           | {orderer}         |
| market_order | pair_id           | {pairId}          |
| market_order | order_direction   | {direction}       |
| market_order | offer_coin        | {offerCoin}       |
| market_order | demand_coin_denom | {demandCoinDenom} |
| market_order | price             | {price}           |
| market_order | amount            | {amount}          |
| market_order | order_id          | {orderId}         |
| market_order | batch_id          | {batchId}         |
| market_order | expire_at         | {expireAt}        |
| market_order | refunded_coins    | {refundedCoins}   |
| message      | module            | liquidity         |
| message      | action            | market_order      |
| message      | sender            | {senderAddress}   |

### MsgMMOrder

| Type     | Attribute Key      | Attribute Value |
|----------|--------------------|-----------------|
| mm_order | orderer            | {orderer}       |
| mm_order | pair_id            | {pairId}        |
| mm_order | batch_id           | {batchId}       |
| mm_order | order_ids          | {orderIds}      |
| mm_order | canceled_order_ids | {orderIds}      |
| message  | module             | liquidity       |
| message  | action             | mm_order        |
| message  | sender             | {senderAddress} |

### MsgCancelOrder

| Type         | Attribute Key | Attribute Value |
|--------------|---------------|-----------------|
| cancel_order | orderer       | {orderer}       |
| cancel_order | pair_id       | {pairId}        |
| cancel_order | order_id      | {orderId}       |
| message      | module        | liquidity       |
| message      | action        | cancel_order    |
| message      | sender        | {senderAddress} |

### MsgCancelAllOrders

| Type              | Attribute Key      | Attribute Value   |
|-------------------|--------------------|-------------------|
| cancel_all_orders | orderer            | {orderer}         |
| cancel_all_orders | pair_ids           | {pairIds}         |
| cancel_all_orders | canceled_order_ids | {orderIds}        |
| message           | module             | liquidity         |
| message           | action             | cancel_all_orders |
| message           | sender             | {senderAddress}   |

### MsgCancelMMOrder

| Type            | Attribute Key      | Attribute Value |
|-----------------|--------------------|-----------------|
| cancel_mm_order | orderer            | {orderer}       |
| cancel_mm_order | pair_id            | {pairId}        |
| cancel_mm_order | canceled_order_ids | {orderIds}      |
| message         | module             | liquidity       |
| message         | action             | cancel_mm_order |
| message         | sender             | {senderAddress} |

## EndBlocker

### Batch Result for MsgDeposit

| Type           | Attribute Key    | Attribute Value  |
|----------------|------------------|------------------|
| deposit_result | request_id       | {reqId}          |
| deposit_result | depositor        | {depositor}      |
| deposit_result | pool_id          | {poolId}         |
| deposit_result | deposit_coins    | {depositCoins}   |
| deposit_result | accepted_coins   | {acceptedCoins}  |
| deposit_result | refunded_coins   | {refundedCoins}  |
| deposit_result | minted_pool_coin | {mintedPoolCoin} |
| deposit_result | status           | {status}         |

### Batch Result for MsgWithdraw

| Type              | Attribute Key    | Attribute Value  |
|-------------------|------------------|------------------|
| withdrawal_result | request_id       | {reqId}          |
| withdrawal_result | withdrawer       | {withdrawer}     |
| withdrawal_result | pool_id          | {poolId}         |
| withdrawal_result | pool_coin        | {poolCoin}       |
| withdrawal_result | refunded_coins   | {refundedCoins}  |
| withdrawal_result | withdrawn_coins  | {withdrawnCoins} |
| withdrawal_result | status           | {status}         |

### Batch Result for MsgLimitOrder, MsgMarketOrder

| Type               | Attribute Key        | Attribute Value      |
|--------------------|----------------------|----------------------|
| order_result       | order_direction      | {direction}          |
| order_result       | orderer              | {orderer}            |
| order_result       | pair_id              | {pairId}             |
| order_result       | order_id             | {orderId}            |
| order_result       | amount               | {amount}             |
| order_result       | open_amount          | {openAmount}         |
| order_result       | offer_coin           | {offerCoin}          |
| order_result       | remaining_offer_coin | {remainingOfferCoin} |
| order_result       | received_coin        | {receivedCoin}       |
| order_result       | status               | {status}             |
| user_order_matched | order_direction      | {orderDirection}     |
| user_order_matched | orderer              | {orderer}            |
| user_order_matched | pair_id              | {pairId}             |
| user_order_matched | order_id             | {orderId}            |
| user_order_matched | matched_amount       | {matchedAmount}      |
| user_order_matched | paid_coin            | {paidCoin}           |
| user_order_matched | received_coin        | {receivedCoin}       |
| pool_order_matched | order_direction      | {orderDirection}     |
| pool_order_matched | pair_id              | {pairId}             |
| pool_order_matched | pool_id              | {poolId}             |
| pool_order_matched | matched_amount       | {matchedAmount}      |
| pool_order_matched | paid_coin            | {paidCoin}           |
| pool_order_matched | received_coin        | {receivedCoin}       |