# Concepts

In `liquidfarming` module, the rewards received to the corresponding pool is auto-compounded, i.e., the rewards are converted to the pool coin and farmed.
This convertion of the rewards to the pool coin is done by auction process.
The bidders for the auction of a given `liquidfarm` place bids with the pool coins as bidding coin in order to get the rewards to be received at the end of the auction period.
The bidder placed with the highest bidding amount becomes the winning bidder of the auction and gets all the rewards accumulated to the `liquidfarm` from the last auction.
Then, the `liquidfarm` farms the winning bidding amount via `lpfarm` module, which is the role of the auto-compounding of the rewards.

# Auction Bidders

Auction bidders need to be aware of the following parameters to join the auctions.

## Whether the target pool is registered in governance parameter as liquidfarm

`Liquidfarm` functionalities can be provided after a target pool to the governance parameters.
The `id` of the `liquidfarm` becomes the same as the `id` of the corresponding pool.
Once `liquidfarm` is added to the governance parameters, the auction process, liquidfarm, liquidunfarm and other functions of the `liquidfarm` can be used.

## Auction period of the liquidfarm

The period of the auction of a `liquidfarm` is set when the `liquidfarm` is added to the governance parameters.
When a `liquidfarm` is added to the governance parameters, the first auction is started at 00:00UTC in the next day.
In every auction period, one auction is going on for a `liquidfarm`.
If no bid is placed for the auction, the auction is marked as skipped.
If the `liquidfarm` is removed to the governance parameters, all the bids to the active auction is refunded and the auction is finished.

## Rewards to be received to the liquidfarm

The exact rewards to be received cannot be known at the point of the time of bidding.
This is because the farming rewards to be received for the `liquidfarm` depend on the amount of farming coin of the `liquidfarm` and the total farming amount of the target pool. This farming amounts vary in every blocks.
A bidder can query the accumulated rewards to the `liquidfarm` and estimates based on the results, which might not be an exact value.
So, bidders should consider the estimate of the rewards and consider prices of the rewards and pool coin to determine its bidding amount.

## Bidding process

A bidder pays pool coins to place a bid, which is the bidding coin.
The bidding amount should be higher than the current highest bidding amount.
A bidder can place a single bid for the same auction of the `liquidfarm`.
When a bidder places another bid with higher bidding amount, then the previous bid of the bidder is refunded.

## Fee

The `liquidfarming` module collects fee from the received rewards when the auction is finished before giving the rewards to the winner.

# Liquidfarmers

Liquidfarmers need to be aware of the potential reduction of the effective rewards.
If there is only one bidder who places a bid with very low bidding amount compared to the received rewards, then the effective APR can be reduced.
The `liquidfarming` module expects that the multiple bidders compete each other and the `liquidfarm` can get the proper rewards via auction process.
