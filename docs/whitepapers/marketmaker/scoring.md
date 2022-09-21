# Abstract

The scoring system is for off-chain evaluation of market maker performance from on-chain data stored in Crescent Network

# Concepts

## Spread

Spread(or Bid-ask spread) measures the distance of 2-sided liquidity. It represents the minimum cost of liquidity consumption. Spread is a price difference between the lowest ask and highest bid.

## Width

Width(ask or bid) indicates the difference between each side of the order’s maximum and minimum price. The market price does not stand still, so the proper width of maker orders helps stay stable.

## Depth

Depth(ask or bid) is the total order amount on each side. If market depth is insufficient enough to handle large trade amounts, market liquidity is low despite its narrow spread. 

## Uptime

Uptime measures the availability of liquidity. High uptime ensures that users can trade at any time they want. Uptime is calculated as a percentage of time a market maker provides orders. 

## Market Maker Order Type

Market makers are provided a particular order type to place and cancel multiple market-making orders by a single transaction. This specific order type allows market makers atomic placement of a group of limit orders, minimizing broadcasting failures and burdens.  


# Scoring

## Methodology

- The evaluation objective is to reward market makers who consistently provide tight and deep liquidity for users.
- Market makers earn points according to liquidity contribution, which means there is a score cap per block.
- The calculation is done off-chain based on on-chain blockchain data. The codebase is publicly available for any third-party verification.

## Market Maker Eligibility

- Market maker eligibility is decided by governance every month.
- Only registered market makers can earn incentives based on their scores. Market makers who fail to meet the uptime requirements are also given their month incentives.
- New applicants are requested to submit ApplyMarketMaker transaction to be included in the uptime evaluation. Those who meet the uptime requirement will be included in eligible market makers through governance.
- The governance can exclude existing market makers who fail to meet the requirements for 2 consecutive months or 3 months within the last 5 months.

## Market Making Orders Requirements

- Orders satisfying following conditions can be recognized and evaluated:
    - Use MMOrder from eligible MMAddress 
    - `Spread` narrower than `MaxSpread`
    - `AskWidth`/`BidWidth`wider than `MinWidth`
    - `AskDepth`/`BidDepth` larger than `MinDepth`
- Parameters (bottom of the document) are assigned differently for each pair depending on the market characteristics and can be adjusted by governance.

## Uptime Requirement

- Uptime requirement is measured in 2 stages:
    - `LiveHour` is added as market maker provide valid orders for an hour. Following 1 out of 2 conditions, it fails:
        - No valid orders longer than `MaxDowntime` in a row
        - No valid orders longer than `MaxTotalDowntime` total in an hour
    - `LiveDay` is added as `LiveHour` is larger than `MinHours` in a day
- Those who earn`LiveDay` larger than `MinDays` satisfy the uptime requirement.
- Parameters (bottom of the document) are decided and adjusted by governance.

## Formula

- Following formula is used to compute how much incentives should be rewarded to each market maker per month. The amount of CRE earned is determined by the relative share of each market maker’s score.
 1. First step (within a block)
    - Calculate liquidity point on bid/ask side and take minimum value

      $P = min \lbrack{ {AskQ1}\over{AskD1^2} }+ { {AskQ2}\over{AskD2^2} } + ... , { {BidQ1}\over{BidD1^2} }+ { {BidQ2}\over{BidD2^2} } + ... \rbrack$
       
 2. Second step (within a block)
    - Calculate contribution score of market maker $P \over \sum_i P_i$

 3. Third step (within a month)    
    - Calculate `Uptime` by dividing `LiveHour` with total hours in a month

       U = Total `LiveHour` / Total hours in a month
    
 4. Forth step (Final score)
    - Final score of market maker 'm' is as following :
   
    $$
    S_m = {U_m} ^3 \sum_{t=1}^{B} \lbrack { P_{mt} \over \sum_i P_{it} }\rbrack
    $$
    
         U_m : % of Market maker m’s Uptime in the month
         B  : total number of blocks for the month
         P_mt : liquidity point of market maker 'm' at block t

## Processing Order data

- All variables are used as final values after the trade is made.
- Even if the order of m-th tick is partially filled, the tick can still be a reference point(whether the highest bid or lowest ask) of measures when one of the following conditions is met. If not, the following tick becomes the reference point.
- `AskQ(m)`/`BidQ(m)` is larger than `MinOpenRatio` of the original order amount of the tick
- `AskQ(m)`/`BidQ(m)` is larger than `MinOpenDepthRatio` of `MinDepth`

![1](1.png)

## Measures

- Breadth
    - `MidPrice`  :  Average price of lowest ask and highest bid
    - `Spread`   : Price difference between lowest ask and highest bid divide by `MidPrice`
    - `AskWidth` : Price difference between highest ask and lowest ask divided by `MidPrice`, `BidWidth` is same with bid prices
    - `AskD(n)`  : price difference between `MidPrice` and n-th ask price divided by `MidPrice`, `BidD(n)` is same with bid prices
- Depth
    - `AskQ(n)`/`BidQ(n)` : The number of token remaining in n-th tick
    - `AskDepth`/`BidDepth` : Total number of token remaining in ask/bid side

# Example

  ![2](2.png)

  - Assume `MaxSpread` 0.012, `MinWidth` 0.002, `MinDepth` 100, `MinOpenRatio` 0.5,  `MinOpenDepthRatio` 0.1
- Measure
    - Block 1
    
    |  | Market Maker A | Market Maker B |
    | --- | --- | --- |
    | MidPrice | 9.945 = (9.96 + 9.93) / 2 | 9.945 = (9.97 + 9.92) / 2 |
    | Spread | 0.00301659.. = (9.96 - 9.93) / 9.945 | 0.00502765.. = (9.97 - 9.92) / 9.945 |
    | AskWidth | 0.00301659.. = (9.99-9.96) / 9.945  | 0.00201106.. = (9.99 - 9.97) / 9.945 |
    | BidWidth | 0.00301659.. = (9.93-9.90) / 9.945 | 0.00201106.. = (9.92 - 9.90) / 9.945 |
    | AskD(1) | 0.00150829.. = (9.96-9.945) / 9.945 | 0.00251382.. = (9.97 - 9.945) / 9.945 |
    | AskD(2) | 0.00251382.. = (9.97-9.945) / 9.945 | 0.00351935.. = (9.98 - 9.945) / 9.945 |
    | AskD(3) | 0.00351935.. = (9.98-9.945) / 9.945 | 0.00452488.. = (9.99 - 9.945) / 9.945 |
    | AskD(4) | 0.00452488.. = (9.99-9.945) / 9.945 | - |
    | BidD(1) | 0.00150829.. = (9.945-9.93) / 9.945 | 0.00251382.. = (9.945 - 9.92) / 9.945 |
    | BidD(2) | 0.00251382.. = (9.945-9.92) / 9.945 | 0.00351935.. = (9.945 - 9.91) / 9.945 |
    | BidD(3) | 0.00351935.. = (9.945-9.91) / 9.945 | 0.00452488.. = (9.945 - 9.90) / 9.945 |
    | BidD(4) | 0.00452488.. = (9.945-9.90) / 9.945 | - |
    | AskDepth | 200 | 225 |
    | BidDepth | 160 | 240 |
    | AskQ(1) | 50 | 75 |
    | AskQ(2) | 50 | 75 |
    | AskQ(3) | 50 | 75 |
    | AskQ(4) | 50 | - |
    | BidQ(1) | 40 | 80 |
    | BidQ(2) | 40 | 80 |
    | BidQ(3) | 40 | 80 |
    | BidQ(4) | 40 | - |
    - Block 2
    
    |  | Market Maker A | Market Maker B |
    | --- | --- | --- |
    | MidPrice | 9.935 = (9.96 + 9.91) / 2 | 9.945 = (9.97 + 9.92) / 2 |
    | Spread | 0.00503271.. = (9.96 - 9.91) / 9.935 | 0.00502765.. = (9.97 - 9.92) / 9.945 |
    | AskWidth | 0.00301962.. = (9.99-9.96) / 9.935  | 0.00201106.. = (9.99 - 9.97) / 9.945 |
    | BidWidth | 0.00100654.. = (9.91-9.90) / 9.935 | 0.00201106.. = (9.92 - 9.90) / 9.945 |
    | AskD(1) | 0.00251635.. = (9.96-9.935) / 9.935 | 0.00251382.. = (9.97 - 9.945) / 9.945 |
    | AskD(2) | 0.00352289.. = (9.97-9.935) / 9.935 | 0.00351935.. = (9.98 - 9.945) / 9.945 |
    | AskD(3) | 0.00452944.. = (9.98-9.935) / 9.935 | 0.00452488.. = (9.99 - 9.945) / 9.945 |
    | AskD(4) | 0.00553598.. = (9.99-9.935) / 9.935 | - |
    | BidD(1) | 0.00050327.. = (9.935-9.93) / 9.935 | 0.00251382.. = (9.945 - 9.92) / 9.945 |
    | BidD(2) | 0.00150981.. = (9.935-9.92) / 9.935 | 0.00351935.. = (9.945 - 9.91) / 9.945 |
    | BidD(3) | 0.00251635.. = (9.935-9.91) / 9.935 | 0.00452488.. = (9.945 - 9.90) / 9.945 |
    | BidD(4) | 0.00352289.. = (9.935-9.90) / 9.935 | - |
    | AskDepth | 190 | 225 |
    | BidDepth | 85 | 180 |
    | AskQ(1) | 40 | 75 |
    | AskQ(2) | 50 | 75 |
    | AskQ(3) | 50 | 75 |
    | AskQ(4) | 50 | - |
    | BidQ(1) | 0 | 20 |
    | BidQ(2) | 5 | 80 |
    | BidQ(3) | 40 | 80 |
    | BidQ(4) | 40 | - |
    - In case of market maker A, `MidPrice` has changed because reference point(bid-side) moved from 9.93 to 9.91. `BidQ(2)` does not meet the both conditions of `MinOpenRatio` of original order amount (5/40 < 0.5) and `MinOpenDepthRatio` ( 100(`MinDepth`) * 0.1 = 10 ).
    - For market maker B’s case, reference point haven’t changed. `BidQ(1)` could not meet the `MinOpenRatio` (20/80 < 50)condition, but larger than `MinOpenDepthRatio`.
  
- Liquidity Point
    - Block 1
        - Both market maker A and B meet the condition
            - Spread(A), Spread(B)  < `MaxSpread` (0.012)
            - AskWidth(A), BidWidth(A), AskWidth(B), BidWidth(B) > `MinWidth` (0.002)
            - AskDepth(A), BidDepth(A), AskDepth(B), BidDepth(B) > `MinDepth` (100)
        - Each point on Block 1 is as follows,
            
            $P_{A1} = min \lbrack{ {50}\over{0.0015..^2} }+ {{50}\over{0.0025..^2} } + {{50}\over{0.0035..^2} }+ {{50}\over{0.0045..^2} }, {{40}\over{0.0015..^2} }+ {{40}\over{0.0025..^2} } + {{40}\over{0.0035..^2}}+ {{40}\over{0.0035..^2} }\rbrack$
            
            $P_{A1} = min \lbrack36369600, 29095680\rbrack = 29,095,680$
            
            $P_{B1} = min \lbrack{ {75}\over{0.0025..^2} }+ {{75}\over{0.0035..^2} } + {{75}\over{0.0045..^2} }, {{80}\over{0.0025..^2} }+ {{80}\over{0.0035..^2} } + {{80}\over{0.0045..^2}}\rbrack$
                        
            $P_{B1} = min \lbrack 21586725, 2302580\rbrack = 21,586,725$
            
    - Block 2
        - Market maker A failed to meet the condition,
            - Spread(A), Spread(B) < `MaxSpread` (0.012)
            - AskWidth(A), AskWidth(B), BidWidth(B) > `MinWidth` (0.002)
            - `BidWidth(B) < MinWidth`
            - AskDepth(A), AskDepth(B), BidDepth(B) > `MinDepth` (100)
            - `BidDepth(A) < MinDepth`
            
            $P_{A2} = min \lbrack 14414430,0\rbrack = 0$
            
            $P_{B2} = min \lbrack{ {75}\over{0.0025..^2} }+ {{75}\over{0.0035..^2} } + {{75}\over{0.0045..^2} }, {{20}\over{0.0025..^2} }+ {{80}\over{0.0035..^2} } + {{80}\over{0.0045..^2}}\rbrack$
            
            $P_{B2} = min \lbrack 21586725, 13531150\rbrack = 13,531,150$

- Contribution Score
    - Block 1
        
        $C_{A1} = \frac{29095680}{29095680+21586725} = 0.574078..$
        
        $C_{B1} = \frac{21586725}{29095680+21586725} = 0.425921..$
        
    - Block 2
        
        $C_{A2} = 0$
        
        $C_{B2} = 1$


## Parameters

Scoring system references the following parameters of the market maker module :

### Common Parameters

| Key               | Definition                                          | Example      |
|-------------------|-----------------------------------------------------|--------------|
| MinOpenRatio      | Minimum ratio to maintain the tick order            | 0.5          |
| MinOpenDepthRatio | Minimum ratio of open amount to MinDepth            | 0.1          |
| MaxDowntime       | Maximum allowable consecutive blocks of outage      | 20 (blocks)  |
| MaxTotalDowntime  | Maximum allowable sum of blocks in an hour          | 100 (blocks) |
| MinHours          | Minimum value of LiveHour to achieve LiveDay        | 16           |
| MinDays           | Minimum value of LiveDay to maintain MM eligibility | 22           |

### Parameters for each pair

| Key             | Definition                                                                | Example                                                        |
|-----------------|---------------------------------------------------------------------------|----------------------------------------------------------------|
| PairId          | Pair id of liquidity module                                               | 20                                                             |
| UpdateTime      | Time the pair variables start to be applied to the scoring system         | 2022-12-01T00:00:00Z                                           |
| IncentiveWeight | Incentive weights for each pair                                           | 0.1                                                            |
| MaxSpread       | Maximum allowable spread between bid and ask                              | 0.006 (ETH-USDC pair), 0.012 (ATOM-USDC pair)                  |
| MinWidth        | Minimum allowable price difference of high and low on both side of orders | 0.001 (ETH-USDC pair), 0.002 (ATOM-USDC pair)                  |
| MinDepth        | Minimum allowable order depth on each side                                | 600000000000000000 (ETH-USDC pair), 100000000 (ATOM-USDC pair) |
