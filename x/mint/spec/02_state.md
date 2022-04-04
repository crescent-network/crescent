<!--
order: 2
-->

# State
## LastBlockTime

LastBlockTime defines block time of the last block. It is used to calculate inflation.

- LastBlockTimeKey: `0x90 -> sdk.FormatTimeBytes(time.Time)`

## Params

Minting params are held in the global params store.

- Params: `mint/params -> ProtocolBuffer(params)`
