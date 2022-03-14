<!-- order: 6 -->

## Before-End-Block

These operations occur before the end-block operations for the liquidity module.

### Store requests from messages

After successful message verification and coin `escrow` process, the incoming
`MsgDeposit`, `MsgWithdraw`, `MsgLimitOrder` and `MsgMarketOrder` messages
are converted to requests and stored.

## End-Block

End-block operations for the liquidity module.

### Execute Requests

If there are `{*action}Request` and `Order` that have not yet executed in the batch,
the batch is executed.
This batch contains one or more `Deposit`, `Withdraw`, and swap processes.

- **Transact and refund for each request**

  A liquidity module escrow account holds coins temporarily and releases them when state changes.
  Refunds from the escrow account are made for cancellations, expiration, and failed requests.

- **Set states for each request according to the results**

  After transacting and refunding transactions occurred for each request,
  update the state of each `{*action}Request` or `Order` according to the results.

  Even if the request is completed or expired:

    1. Set the status as `ShouldBeDeleted` instead of deleting the request directly from the `end-block`
    2. Delete the request that have `ShouldBeDeleted` state from the begin-block in the next block
       so that each request with result state in the block can be stored to kvstore.

  This process allows searching for past requests that have this result state.
  Searching is supported when the kvstore is not pruning.