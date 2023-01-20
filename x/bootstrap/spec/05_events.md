<!-- order: 5 -->

# Events

The `bootstrap` module emits the following events:

## Handlers

### MsgIntentBootstrap

| Type               | Attribute Key | Attribute Value    |
|--------------------|---------------|--------------------|
| apply_market_maker | address       | {mmAddress}        |
| apply_market_maker | pair_ids      | []{pairId}         |
| message            | module        | bootstrap        |
| message            | action        | apply_market_maker |
| message            | sender        | {senderAddress}    |

### MsgClaimIncentives

| Type             | Attribute Key | Attribute Value  |
|------------------|---------------|------------------|
| claim_incentives | address       | {mmAddress}      |
| message          | module        | bootstrap      |
| message          | action        | claim_incentives |
| message          | sender        | {senderAddress}  |

## Proposals

### IncludeBootstrap

| Type                 | Attribute Key | Attribute Value |
|----------------------|---------------|-----------------|
| include_market_maker | address       | {mmAddress}     |
| include_market_maker | pair_id       | {pairId}        |


### ExcludeBootstrap

| Type                 | Attribute Key | Attribute Value |
|----------------------|---------------|-----------------|
| exclude_market_maker | address       | {mmAddress}     |
| exclude_market_maker | pair_id       | {pairId}        |


### RejectBootstrap

| Type                | Attribute Key | Attribute Value |
|---------------------|---------------|-----------------|
| reject_market_maker | address       | {mmAddress}     |
| reject_market_maker | pair_id       | {pairId}        |


### DistributeIncentives

| Type                  | Attribute Key    | Attribute Value        |
|-----------------------|------------------|------------------------|
| distribute_incentives | budget_address   | {budgetAddress}        |
| distribute_incentives | total_incentives | {totalIncentivesCoins} |

