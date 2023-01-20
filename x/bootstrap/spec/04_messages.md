<!-- order: 4 -->

# Messages

## MsgApplyBootstrap

Apply as a Market maker, if included through `BootstrapProposal`, eligible becomes true and is deleted if rejected or excluded, need to deposit `DepositAmount` * number of pairs

```go
type MsgApplyBootstrap struct {
    // address specifies the bech32-encoded address of the market maker will market making
    Address string
    PairIds []uint64
}
```

- Fail if the same market maker pair already exists
- Fail for pairs that are not registered as incentive pairs on params
- Fail if the balance is less than `ApplyDeposit` * `len(PairIds)` amount

## MsgClaimIncentives

Claim claimable amount of incentives distributed through `BootstrapProposal` at once

```go
type MsgClaimIncentives struct {
    // address specifies the bech32-encoded address of the market maker that claim incentives
    Address string
}
```