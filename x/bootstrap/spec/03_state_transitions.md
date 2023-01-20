<!-- order: 3 -->
These messages in the market maker module trigger state transitions.

## Apply market maker

apply through `MsgApplyBootstrap` for registered as incentive pairs on params

- deposit apply deposit amount * number of pairs
- create apply deposit object for refund when inclusion or rejection
- create market maker object for each pair

### Inclusion

- Set the `Bootstrap.Eligible` value to true
- refund `Deposit` amount and delete `Deposit`

### Reject

- Delete the `Bootstrap`
- refund `Deposit` amount and delete `Deposit`

### Exclusion

- Delete existing eligible `Bootstrap`

## Incentive Distribution

send from `params.IncentiveBudgetAddress` to `ClaimableIncentiveReserveAcc` as much as the input amount for the existing eligible market maker, and create or update Incentive object with claimable amount

### Claim

When distribution occurs through `BootstrapProposal.Distributions` and there is claimable incentive, the whole amount can be claim through `MsgClaimIncentives`

- Send all claimable Incentives to the market maker
- Delete the Incentive object