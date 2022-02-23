<!-- order: 7 -->

# Hooks

## TallyLiquidStakingGov

Calculate the corresponding voting power of the voter who owns bToken by the following method

- balance of bToken
- balance of PoolCoins including bToken
- farming position of bToken
- farming position of PoolCoins including bToken

This calculation is dependent on modules `x/liquidity` and `x/farming`, the farming position includes staking and queued staking.

the calculated voting power is added, deducted, overwritten as `otherVotes` on tally of `cosmos-sdk/x/gov` by calling `govHooks.GetOtherVotes` 

each voting power of `otherVotes` is distributed to liquid validators by current weight of **bonded** liquidTokens each liquid validators has **bonded** status of `cosmos-sdk/x/staking` module states     
