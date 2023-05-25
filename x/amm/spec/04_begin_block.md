<!-- order: 4 -->

# Begin-Block

## Allocate Farming Rewards

The allocation of rewards is done by following procedure:

1. Calculate the block duration(the time between the last block and the current
    block) and clip the duration to its maximum value specified by the
    `MaxFarmingBlockTime` param.
2. Collect all active(non-terminated and its `StartTime` has past) plans.
3. For each active plan, iterate through its reward allocation entries and
    calculate how many rewards should be allocated to each pair for this block
    based on the block duration.
    Note that a pool can be rewarded by many farming plans.
4. Move rewards from each farming pool to the `RewardsPoolAddress` and increase
    farming rewards growth of the pool.
