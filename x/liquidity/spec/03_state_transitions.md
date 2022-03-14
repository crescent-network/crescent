<!-- order: 3 -->

# State Transitions

These messages in the liquidity module trigger state transitions.

## Pair creation

### MsgCreatePair

Add a coin pair to the liquidity module so that users can create a pool
for that coin pair or request a swap order.

## Pool creation

### MsgCreatePool

Create a pool in existing pair so that users can deposit reserve coins or withdraw pool coin.

## Coin Escrow for Liquidity Module Messages

Transaction confirmation causes state transition on the bank module.
Some messages on the liquidity module require coin escrow before confirmation.

### MsgDeposit

To deposit coins into an existing `Pool`, the depositor must escrow `DepositCoins` into `GlobalEscrowAddr`.

### MsgWithdraw

To withdraw coins from a `Pool`, the withdrawer must escrow `PoolCoin` into `GlobalEscrowAddr`.

### MsgLimitOrder, MsgMarketOrder

To request a coin swap, the orderer must escrow `OfferCoin` into each pair’s `EscrowAddress`.

## Cancel Swap Order

### MsgCancelOrder

Cancel an order already stored in the store.
It is impossible to cancel a swap order message submitted in the same batch because
it can be canceled only by specifying order id.

### MsgCancelAllOrders

Cancel the user's all orders for specific pairs or for all pairs in the liquidity module.

## Batch Execution

Batch execution causes state transitions on the `bank` module.
The following categories describe state transition executed by each process in the batch execution.

### Coin Swap

After a successful coin swap, coins accumulated in the pair's `EscrowAddress` for swaps
are sent to other orderer or to the `Pool`.

### Deposit

After a successful deposit transaction, escrowed coins are sent to the `ReserveAddress`
of the targeted `Pool` and new pool coins are minted and sent to the depositor.

### Withdrawal

After a successful withdraw transaction, escrowed pool coins are burned and
corresponding amount of reserve coins are sent to the withdrawer from the liquidity `Pool`.

## Swap Process

### Matching Price Calculation

- CX(i): Sum of all buy amount of orders with tick equal to or higher than tick(i)
- CY(i): Sum of all sell amount of orders with tick equal to or lower than tick(i)
- Check matchability
    - for all tick i, if CX(i) == 0 or CY(i) == 0, there is no matchable order
    - If i < j, for highest tick i with CX(i) > 0 and lowest tick j with CY(j) > 0, there is no matchable order
- price discovery logic
    - find tick(i) with CX(i+1) ≤ CY(i) and CX(i) ≥ CY(i-1)
    - if there are two or more ticks that satisfies above condition
        - tick(l): lowest tick with both conditions hold
        - tick(h): highest tick with both conditions hold
        - result = PriceToRoundedTick((tick(l)+tick(h))/2)

### Matching Execution

- Priority
  1. price: buy order with high order price, sell order with low order price
  2. amount: large amount → small amount
  3. orderbook type: order from pool → order from user
  4. id: small number id → large number id(pool id, order id)
- Matching
  1. line up buy orders with price greater than or equal to matching price and
     sell orders with price less than or equal to matching price.
     both are sorted by above priority, respectively.
  2. Then match each other based on the shorter side.
     Orders at the end of each side may be partially matched.
     Sell orders which receive no quote coin due to decimal truncation will be dropped during this process.

## Change states of orders with expired lifespan

After batch execution, status of all remaining orders with `ExpireAt` higher than
current block time are changed to `OrderStatusExpired`

## Refund escrowed coins

Refunds are issued for escrowed coins for cancelled swap order and failed create pool, deposit, and withdraw messages.
