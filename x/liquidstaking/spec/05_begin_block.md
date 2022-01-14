<!-- order: 5 -->

# Begin-Block

## Commission Parameter Check

In order for the liquid staking module to operate smoothly, it is necessary to keep the commission rate the same for all whitelisted validators. Because commission rate can be changed by each validator via staking module `MsgEditValidator`, in the Begin-Block, it should be preceded by checking that the global parameters `params.CommissionRate` and the rates of each validator are kept the same.

- for all whitelisted validators, their commission rate is compared with the `params.CommissionRate`
- if a validator has different commission rate, the validator will be temporarily removed from the whitelist
- The whitelist of validators is confirmed and the whitelist becomes the standard within the block.
