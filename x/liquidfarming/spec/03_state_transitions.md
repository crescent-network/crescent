<!-- order: 3 -->

# State Transitions

This document describes the state transaction operations in the `liquidfarming` module.

## Parameter Change for Activated Liquid Farms

### Activation of a Liquid Farm

When a new `liquidFarm` with a given pool id is registered to the parameter `LiquidFarms` by governance, the `liquidFarm` with the pool id becomes activated and added to the state `LiquidFarms`.

### Deactivation of a Liquid Farm

When a `liquidFarm` with a given pool id in the parameter `LiquidFarms` is removed by governance, the `liquidFarm` becomes deactivated and deleted in the state `LiquidFarms`.
When the `liquidFarm` becomes deactivated, the module unstakes all pool coins for the `liquidFarm`.

## Coin Escrow for Liquidfarming Module Messages

The following messages cause state transition on the `bank`, `liquidty`, and `farming` modules.

### MsgLiquidFarm

- A farmer farms in the `liquidfarming` module with their farming coin (pool coin)
- The module sends that farming coin to dynamically generated reserve account
- The reserve account farms their farming coin to the `farm` module and start generating farming rewards
- The `LFCoin` is minted and sends it to the farmer.

### MsgLiquidUnfarm

- A farmer unfarms in the `liquidfarming` module with their liquid farming coin (LFCoin)
- The module calculates the corresponding pool coin by the burn rate and releases the pool coin back to the farmer
- The module burns the `LFCoin`

### MsgLiquidUnfarmWithdraw

- A farmer unfarms in the `liquidfarming` module with their liquid farming coin (LFCoin)
- The module calculates the corresponding pool coin by the burn rate and calls to withdraw the pool coin in the `liquidity` module
- The corresponding two coins are back to the farmer
- The module burns the `LFCoin`

### MsgPlaceBid

- Bidding coins are sent to the `PayingReserveAddress` of an auction.

### MsgRefundBid

- Bidding coins are sent to a bidder account from the `PayingReserveAddress` of an auction.
