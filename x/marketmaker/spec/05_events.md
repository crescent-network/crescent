<!-- order: 5 -->

# Events

The `marketmaker` module emits the following events:

## Handlers

### MsgIntentMarketMaker

| Type               | Attribute Key | Attribute Value    |
|--------------------|---------------|--------------------|
| apply_market_maker | address       | {mmAddress}        |
| apply_market_maker | pair_ids      | []{pairId}         |
| message            | module        | marketmaker        |
| message            | action        | apply_market_maker |
| message            | sender        | {senderAddress}    |

### MsgClaimIncentives

| Type             | Attribute Key | Attribute Value  |
|------------------|---------------|------------------|
| claim_incentives | address       | {mmAddress}      |
| message          | module        | marketmaker      |
| message          | action        | claim_incentives |
| message          | sender        | {senderAddress}  |

## Proposals

### IncludeMarketMaker

| Type                 | Attribute Key | Attribute Value |
|----------------------|---------------|-----------------|
| include_market_maker | address       | {mmAddress}     |
| include_market_maker | pair_id       | {pairId}        |


### ExcludeMarketMaker

| Type                 | Attribute Key | Attribute Value |
|----------------------|---------------|-----------------|
| exclude_market_maker | address       | {mmAddress}     |
| exclude_market_maker | pair_id       | {pairId}        |


### RejectMarketMaker

| Type                | Attribute Key | Attribute Value |
|---------------------|---------------|-----------------|
| reject_market_maker | address       | {mmAddress}     |
| reject_market_maker | pair_id       | {pairId}        |


### DistributeIncentives

| Type                  | Attribute Key    | Attribute Value        |
|-----------------------|------------------|------------------------|
| distribute_incentives | budget_address   | {budgetAddress}        |
| distribute_incentives | total_incentives | {totalIncentivesCoins} |

