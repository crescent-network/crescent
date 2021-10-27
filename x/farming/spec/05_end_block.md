<!-- order: 6 -->
# End-Block

++ https://github.com/tendermint/farming/blob/69db071ce3/x/farming/abci.go#L13-L46

At the end of each block, it terminates plans that their end time has passed over the current block time. It sends all remaining coins in the plan's farming pool account `FarmingPoolAddress` to the termination address `TerminationAddress` and mark the plan as terminated by making `Terminated` true. A global parameter `NextEpochDays` is there but the farming module uses an internal state `CurrentEpochDays` to prevent from a case where it may affect the rewards allocation. Suppose `NextEpochDays` is 7 and it is proposed to change the value to 1 through governance proposal. Although the proposal is passed, rewards allocation should continue to proceed with 7, not 1. The [test code](https://github.com/tendermint/farming/blob/69db071ce3/x/farming/abci_test.go#L12-L64) is available for you if you want to understand it in more detail. Then it allocates farming rewards, processes `QueueStaking` to be staked, and sets `LastEpochTime` to keep in track in case of the chain upgrade.
