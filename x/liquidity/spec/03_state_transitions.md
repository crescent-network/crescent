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

### MsgCreateRangedPool

Create a ranged liquidity pool in existing pair.

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

## Matching Process

Read more about matching process in the [Liquidity pool white paper](../../../docs/whitepapers/liquidity/matching.md).

## Change states of orders with expired lifespan

After batch execution, status of all remaining orders with `ExpireAt` higher than
current block time are changed to `OrderStatusExpired`

## Refund escrowed coins

Refunds are issued for escrowed coins for cancelled swap order and failed create pool, deposit, and withdraw messages.
