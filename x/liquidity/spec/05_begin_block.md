<!-- order: 5 -->

# Begin-Block

Begin block operations for the liquidity module delete requests that were executed or ready to be deleted.

## **Delete batch messages**

- Delete `DepositRequest` and `WithdrawRequest` messages with status `RequestStatusSucceeded`
  or `RequestStatusFailed`
- Delete `SwapRequest` messages with status `SwapRequestStatusCompleted`, `SwapRequestStatusCanceled` or `SwapRequestStatusExpired`