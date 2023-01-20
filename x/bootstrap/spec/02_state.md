<!-- order: 2 -->

# State

The `bootstrap`
module keeps track of the market maker states.

## Bootstrap

Market maker object created by applying, if included through `BootstrapProposal`, eligible becomes true and is deleted if rejected or excluded

```go
type Bootstrap struct {
    Address    string
    PairId     uint64
    Eligible   bool
}
```

## Incentive

Store the total amount of incentives distributed through `BootstrapProposal`, and it can be claimed at once through `MsgClaimIncentives`

```go
type Incentive struct {
    Address   string
    Claimable sdk.Coins
}
```

## Deposit

stores apply deposit amount for a future refund

```go
type Deposit struct {
    Amount sdk.Coins
}
```

# Parameter

- ModuleName: `bootstrap`
- RouterKey: `bootstrap`
- StoreKey: `bootstrap`
- QuerierRoute: `bootstrap`

# Store

Stores are KVStores in the `multistore`. The key to find the store is the first parameter in the list.

### **The key to get the market maker object by address and pair id**

- BootstrapKey: `[]byte{0xc0} | AddressLen (1 byte) | Address | PairId -> ProtocalBuffer(Bootstrap)`

### **The index key to get the market maker object by pair id and address**

- BootstrapIndexByPairIdKey: `[]byte{0xc1} | PairId | Address -> nil`

### **The key to get the deposit object by address and pair id**

- DepositKey: `[]byte{0xc2} | AddressLen (1 byte) | Address | PairId -> ProtocalBuffer(Deposit)`

### **The key to get the incentive object**

- IncentiveKey: `[]byte{0xc5} | Address -> ProtocalBuffer(Incentive)`