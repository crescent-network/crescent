<!-- order: 3 -->

# State Transitions

These messages in the liquidity module trigger state transitions.

## Pair creation

### MsgCreatePair
Add a coin pair to the liquidity module so that users can create a pool for that coin pair or request a swap order.

## Coin Escrow for Liquidity Module Messages

Transaction confirmation causes state transition on the Bank module. Some messages on the liquidity module require coin escrow before confirmation.

### MsgDeposit

To deposit coins into an existing `Pool`, the depositor must escrow `DepositCoins` into `GlobalEscrowAddr`.

### MsgWithdraw

To withdraw coins from a `Pool`, the withdrawer must escrow `PoolCoin` into `GlobalEscrowAddr`.

### MsgSwap

To request a coin swap, the swap requestor must escrow `OfferCoin` into each pair’s `EscrowAddress`.

## Batch Execution

Batch execution causes state transitions on the `Bank` module. The following categories describe state transition executed by each process in the batch execution.

### Coin Swap

After a successful coin swap, coins accumulated in the pair's `EscrowAddress` for swaps are sent to other swap requestors or to the `Pool`.

### Deposit

After a successful deposit transaction, escrowed coins are sent to the `ReserveAccount` of the targeted `Pool` and new pool coins are minted and sent to the depositor.

### Withdrawal

After a successful withdraw transaction, escrowed pool coins are burned and a corresponding amount of reserve coins are sent to the withdrawer from the liquidity `Pool`.

## Swap Process

### Find price direction
- get mid price
    - mid price = (HighestBuyPrice+LowestSellPrice)/2
- if one or more than one price does not exist → no matching
- expecting direction
    - definitions
        - BuyGteMidPrice (BGMP) : buy amount greater than or equal to mid price
        - SellLteMidPrice (SLMP) : sell amount less than or equal to mid price
    - price direction
        - increasing : if BGMP > SLMP
        - decreasing : if BGMP < SLMP
        - staying : otherwise

### Matching Price Calculation
- variables
  `CX(i)` : Sum of all buy amount of orders with tick equal or higher than this `tick(i)`
  `CY(i)` : Sum of all sell amount of orders with tick equal or less than this `tick(i)`
- for all tick i, if `CX(i) == 0` or `CY(i) == 0`, there is no matchable order
- if i < j, for highest tick i with `CX(i)` > 0 and lowest tick j with `CY(j)` > 0, there is no matchable order
- for staying case
    - if mid price is on tick, mid price is matching price
    - else : applying banker’s rounding for mid price to match the tick precision
- price discovery for increasing or decreasing case
    - initial price (start of iteration)
        - increasing case : tick with price less than or equal to mid price
        - decreasing case : tick with price greater than or equal to mid price
    - end of iteration
        - increasing case : highest tick i with `CX(i)` > 0
        - decreasing case : lowest tick j with `CY(j)` > 0
    - discovery logic
        - find first `tick(i)` with `CX(i+1)` ≤ `CY(i)` and `CX(i)` ≥ `CY(i-1)`
        - if there is no tick with upper condition, the last tick(from end of iteration) is the matching price

### Matching Execution
- Priority
    - price : buy order with high order price, sell order with low order price
    - amount : large amount → small amount
    - orderbook : order from user → order from pool
    - id : small number id → large number id (pool id, order id)
- Matching : buy orders with price greater than or equal to matching price and sell orders with price less than or equal to matching price are sorted by above priority, respectively. 
Then match each other based on the shorter side. Orders at the end of the longer side may be partially matched.

## Change states of swap requests with expired lifespan

After batch execution, status of all remaining swap requests with `ExpiredAt` higher than current block time are changed to `SwapRequestStatusExpired`

## Refund escrowed coins

Refunds are issued for escrowed coins for cancelled swap order and failed create pool, deposit, and withdraw messages.


