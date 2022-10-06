<!-- order: 4 -->

# Begin-Block

## Rewards Allocation

The allocation of rewards is done by following procedure:

1. Calculate the block duration(the time between the last block and the current
    block) and clip the duration to its maximum value specified by the
    `MaxBlockDuration` param.
2. Collect all active(non-terminated and its `StartTime` has past) plans.
3. For each active plan, iterate through its reward allocation entries and
    calculate how many rewards should be allocated to each pair for this block
    based on the block duration.
    Note that a pair can be rewarded by many farming plans.
4. Iterate through all active plans again and calculate the amount of rewards
    for each pool coin denom based on the pool's *reward weight*.
5. Move rewards from each farming pool to the `RewardsPoolAddress` and increase
     `CurrentRewards` and `OutstandingRewards` for pool coins.
