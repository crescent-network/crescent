<!-- order: 7 -->

# Proposal

## MarketMakerProposal

```go
type MarketMakerProposal struct {
    // title specifies the title of the proposal
    Title string 
    // description specifies the description of the proposal
    Description string
    // set the market makers to eligible, refund deposit
    Inclusions []MarketMakerHandle
    // delete existing eligible market makers
    Exclusions []MarketMakerHandle
    // delete the not eligible market makers, refund deposit
    Rejections []MarketMakerHandle
    // distribute claimable incentive to eligible market makers
    Distributions []IncentiveDistribution
}

type MarketMakerHandle struct {
    // registered market maker address
    Address string
    PairId uint64
}

type IncentiveDistribution struct {
    // registered market maker address
    Address string
    PairId uint64
    Amount sdk.Coins
}

```

- MarketMakerProposal is passed through the gov module in the following order
  - Inclusions
  - Distributions
  - Exclusions
  - Rejections

- The same market maker cannot be duplicated with inclusion, exclusion and rejection
- inclusion
    - include only not eligible market maker
- exclusion
    - exclude only for existing eligible market maker
- rejection
    - reject only not eligible market maker
- distribution
    - distribute only for eligible market makers
    - sufficient balance of `IncentiveBudgetAcc`