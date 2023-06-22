<!-- order: 6 -->

# Parameters

The budget module contains the following parameters:


| Key         | Type     | Example                                                                                                                                                                                                                                                                                                                 |
|-------------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| EpochBlocks | uint32   | {"epoch_blocks":1}                                                                                                                                                                                                                                                                                                      |
| Budgets     | []Budget | {"budgets":[{"name":"liquidity-farming-20213Q-20221Q","rate":"0.300000000000000000","source_address":"cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta","destination_address":"cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky","start_time":"2021-10-01T00:00:00Z","end_time":"2022-04-01T00:00:00Z"}]} |

## EpochBlocks

The universal epoch length in number of blocks.

Every process for budget collecting is executed with the `epoch_blocks` frequency.

- The default value is 1. 
- All budget collections are disabled if the value is 0. 

Budget collection logic is executed with the following condition. 

```
params.EpochBlocks > 0 && Current Block Height % params.EpochBlocks == 0
```

## Budgets

The budget structure is described in [State](02_state.md).

Parameter of a budget can be added, modified, and deleted through parameter change governance proposal.

### Validity Checks

- Budget name: 

  - Supports valid characters are letters (`A-Z, a-z`), digits(`0-9`), and `-`. 

  - Must not include spaces. 

  - Has a maximum length of 50. 
  
  - Must be unique among existing budget names.

- Validate `DestinationAddress` address.

- Validate `SourceAddress` address.

- EndTime must not be earlier than StartTime.

- The total rate of budgets with the same `SourceAddress` value must not exceed 1 (100%).
