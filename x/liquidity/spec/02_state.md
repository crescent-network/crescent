<!-- order: 2 -->

# State

The `liquidity` module keeps track of ...

## Pair

## Pool

## DepositRequest

## WithdrawRequest

## SwapRequest

## SwapDirection

## SwapRequestStaus

# Parameter

- ModuleName: `liquidity`
- RouterKey: `liquidity`
- StoreKey: `liquidity`
- QuerierRoute: `liquidity`

# Store

Stores are KVStores in the `multistore`. The key to find the store is the first parameter in the list.

### The key for the latest pair id

- LastPairIdKey: `[]byte{0xa0} -> ProtocolBuffer(uint64)`

### The key for the latest pool id

- LastPoolIdKey: `[]byte{0xa1} -> ProtocolBuffer(uint64)`

### The key to get the pair object 

- PairKey: `[]byte{0xa5} | PairId -> ProtocolBuffer(Pair)`

### The index key to get the pair object by base and quote denoms

- PairIndexKey: `[]byte{0xa6} | BaseCoinDenomLen (1 byte) | BaseCoinDenom | QuoteCoinDenomLen (1 byte) | QuoteCoinDenom -> ProtocolBuffer(uint64)`

### The index key to lookup pairs with the given denom

- PairIndexKey: `[]byte{0xa7} | BaseCoinDenomLen (1 byte) | BaseCoinDenom | QuoteCoinDenomLen (1 byte) | QuoteCoinDenom | PairId -> nil`

### The key to get the pool object

- PoolKey: `[]byte{0xab} | PoolId -> ProtocolBuffer(Pool)`

### The index key to get the pool object from the reserve address

- PoolByReserveAddressIndexKey: `[]byte{0xac} | ReserveAddressLen (1 byte) | ReserveAddress -> ProtocolBuffer(uint64)`

### The index key to lookup pools by pair id

- PoolsByPairIndexKey: `[]byte{0xad} | PairId | PoolId -> nil`

### The key to get the deposit request by pool id and deposit request id

- DepositRequestKey: `[]byte{0xb0} | PoolId | DepositRequestId -> ProtocolBuffer(DepositRequest)`

### The key to get the withdraw request by pool id and withdraw request id

- WithdrawRequestKey: `[]byte{0xb1} | PoolId | WithdrawRequestId -> ProtocolBuffer(WithdrawRequest)`

### The key to get the swap request by pool id and swap request id

- SwapRequestKey: `[]byte{0xb2} | PoolId | SwapRequestId -> ProtocolBuffer(SwapRequest)`
