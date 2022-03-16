<!-- order: 5 -->
# End-Block

At each end-block call, the `farming` module operations are specified to execute.

++ `/x/farming/abci.go`

At the end of each block:

- Terminates plans if their end time has passed over the current block time. 

  - Sends all remaining coins in the plan's farming pool account `FarmingPoolAddress` to the termination address `TerminationAddress`.
  - Marks the plan as terminated by making `Terminated` true. 
  - Allocates farming rewards.
  - Processes `QueueStaking` to be staked.
  - Sets `LastEpochTime` to track in case of chain upgrade.

## Internal state CurrentEpochDays

Although a global parameter `NextEpochDays` exists, the farming module uses an internal state `CurrentEpochDays` to prevent impacting rewards allocation. 

Suppose `NextEpochDays` is 7 and it is proposed to change the value to 1 through governance proposal. Although the proposal is passed, rewards allocation must continue to proceed with 7, not 1. 

To explore internal state `CurrentEpochDays` in more detail, see the test code on `/x/farming/abci_test.go` 
