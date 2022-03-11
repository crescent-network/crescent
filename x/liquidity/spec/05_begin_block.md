<!-- order: 5 -->

# Begin-Block

Begin block operations for the liquidity module delete requests that were executed or ready to be deleted.

## **Delete batch messages**

- Delete `DepositRequest` and `WithdrawRequest` messages with status `RequestStatusSucceeded`
  or `RequestStatusFailed`
- Delete `Order` messages with status `OrderStatusCompleted`, `OrderStatusCanceled` or `OrderStatusExpired`