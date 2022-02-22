<!-- order: 7 -->

# Hooks

## TallyLiquidGov

Calculate the corresponding voting power of the voter who owns bToken by the following method

- balance of bToken
- balance of PoolCoins including bToken
- farming position of bToken
- farming position of PoolCoins including bToken

This calculation is dependent on modules `x/liquidity` and `x/farming`, the farming position includes staking and queued staking.