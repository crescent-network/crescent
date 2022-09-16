<!-- order: 3 -->

# State Transitions

This document describes the state transaction operations in the `liquidfarming` module.

## Parameter Change for Activated Liquid Farms

### Activation of a Liquid Farm

When a new `liquidFarm` with a given pool id is added to the parameter `LiquidFarms` by governance, the `liquidFarm` with the pool id becomes activated and added to the state `LiquidFarms`.

### Deactivation of a Liquid Farm

When a `liquidFarm` with a given pool id in the parameter `LiquidFarms` is removed by governance, the `liquidFarm` becomes deactivated and deleted in the state `LiquidFarms`.
When the `liquidFarm` becomes deactivated, the module unstakes all pool coins for the `liquidFarm`.

## Coin Escrow for Liquidfarming Module Messages

The following messages cause state transition on the `bank`, `liquidty`, and `farming` modules.

### MsgFarm

- Pool coins are sent to a reserve address of a liquid farm.
- The `liquidfarming` module stakes the pool coins to the `farming` module.
- LF coins are minted and sent to the farmer.

### MsgUnfarm

- LF coins are sent to the `liquidfarm` module account, and the LF coins are burnt.
- The `liquidfarming` module unstakes pool coins from the `farming` module. 
- The pool coins are sent from a reserve address of a liquid farm to a farmer.

### MsgUnfarmWithdraw

- LF coins are sent to the `liquidfarm` module account, and the LF coins are burnt.
- The `liquidfarming` module unstakes pool coins from the `farming` module. 
- The pool coins are sent from a reserve account of a liquid farm to a farmer.
- The pool coins are sent to a reserve account in `liquidity` module, and the corresponding coins are withdrawn to the farmer.

### MsgPlaceBid

- Bidding coins are sent to the `PayingReserveAddress` of an auction.

### MsgRefundBid

- Bidding coins are sent to a bidder account from the `PayingReserveAddress` of an auction.


## State transition by hooks from other module
The following events triggered by hooks cause state transition on the `bank`, `liquidty`, and `farming` modules.

### AfterAllocateRewards hook from `farming` module

When `AfterAllocateRewards` hook is delivered, the following operations are performed.
- If the auction currently going on exists, the current auction becomes finished. And, 
  - the winner is chosen,
  - the rewards is harvested and sent to the winner,
  - the pool coins from the winner in the paying reserve address is sent to the module account,
  - the module stakes the pool coins from the auction, the amount of these pool coins is saved to `CompoundingRewards`
  - the pool coins from the others not winner in the paying reserve address is refunded to each bidder’s account.
- A new auction is created.
