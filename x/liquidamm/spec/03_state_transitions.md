<!-- order: 3 -->

# State Transitions

This document describes the state transaction operations in the `liquidfarming` module.

## Parameter Change for Activated Liquid Farms

### Activation of a Liquid Farm

When a new `liquidFarm` with a given pool id is registered to the parameter `LiquidFarms` by governance, the `liquidFarm` with the pool id is added to the state `LiquidFarms`.

### Deactivation of a Liquid Farm

When a `liquidFarm` with a given pool id in the parameter `LiquidFarms` is removed by governance, the `liquidFarm` is deleted in the state `LiquidFarms`.
When the `liquidFarm` is deleted in the state `LiquidFarms`, the module unstakes all pool coins for the `liquidFarm`.

## Coin Escrow for Liquidfarming Module Messages

The following messages cause state transition on the `bank`, `liquidty`, and `lpfarm` modules.

### MsgLiquidFarm

- A farmer farms in the `liquidfarming` module with their pool coins.
- The module sends that the pool coin to the reserve account.
- The reserve account farms their pool coin to the `lpfarm` module and start generating farming rewards.
- The `LFCoin` is minted and sends it to the farmer.

### MsgLiquidUnfarm

- A farmer unfarms in the `liquidfarming` module with their liquid farming coin (LFCoin).
- The module calculates the corresponding pool coin by the burn rate and sends the pool coin back to the farmer.
- The module burns the unfarmed `LFCoin`.

### MsgLiquidUnfarmWithdraw

- A farmer unfarms in the `liquidfarming` module with their liquid farming coin (LFCoin).
- The module calculates the corresponding pool coin by the burn rate and calls to withdraw the pool coin in the `liquidity` module.
- The module burns the unfarmed `LFCoin`.
- The corresponding two kinds of withdrawn coins are sent back to the farmer.

### MsgPlaceBid

- Bidding coins are sent to the `PayingReserveAddress` of an auction.
- If a bidder who already placed a bid in the liquidfarm places another bid with higher bidding amount, then the previous bidding coins is refunded and the new bidding coins are sent to the `PayingReserveAddress` of an auction.

### MsgRefundBid

- Bidding coins are sent to a bidder account from the `PayingReserveAddress` of an auction.
