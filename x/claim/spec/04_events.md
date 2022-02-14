<!-- order: 4 -->

# Events

The `claim` module emits the following events:

## Handlers

### MsgClaim

| Type    | Attribute Key           | Attribute Value         |
| ------- | ----------------------- | ----------------------- |
| claim   | recipient               | {recipientAddress}      |
| claim   | initial_claimable_coins | {initialClaimableCoins} |
| claim   | claimable_coins         | {claimableCoins} |
| claim   | deposit_action_claimed  | {depositActionClaimed}  |
| claim   | swap_action_claimed     | {swapActionClaimed}     |
| claim   | farming_action_claimed  | {farmingActionClaimed}  |
| message | module                  | claim                   |
|         |                         |                         |