<!-- order: 2 -->

# State

### Airdrop

```go

// Airdrop defines airdrop information.
type Airdrop struct {
	Id                 uint64          // the airdrop id
	SourceAddress      string          // the bech32-encoded source address
	Conditions         []ConditionType // the list of conditions
	StartTime          time.Time       // the start time of the airdrop
	EndTime            time.Time       // the end time of the airdrop
}
```

### Claim Record

```go
// ClaimRecord defines claim record that corresponds to the airdrop.
type ClaimRecord struct {
	AirdropId             uint64    // airdrop id
	Recipient             string    // the bech32-encoded address that is eligible to claim airdrop
	InitialClaimableCoins sdk.Coins // the initial claimable coins
	ClaimableCoins        sdk.Coins // the unclaimed claimable coins
	ClaimedConditions     []bool    // the list of condition statuses
}
```

### Condition Type

```go
// ConditionType defines the type of condition that a recipient must execute in order to receive a claimable amount.
type ConditionType int32

const (
	// CONDITION_TYPE_UNSPECIFIED specifies an unknown condition type
	ConditionTypeUnspecified ConditionType = 0
	// CONDITION_TYPE_DEPOSIT specifies deposit condition type
	ConditionTypeDeposit ConditionType = 1
	// CONDITION_TYPE_SWAP specifies swap condition type
	ConditionTypeSwap ConditionType = 2
	// CONDITION_TYPE_LIQUIDSTAKE specifies liquid stake condition type
	ConditionTypeLiquidStake ConditionType = 3
	// CONDITION_TYPE_VOTE specifies governance vote condition type
	ConditionTypeVote ConditionType = 4
)
```

### Parameters

- ModuleName: `claim`
- RouterKey: `claim`
- StoreKey: `claim`
- QuerierRoute: `claim`


### Stores

Stores are KVStores in the multi-store. The key to find the store is the first parameter in the list.

- `AirdropKey: 0xd5 | AirdropId -> ProtocolBuffer(Airdrop)`
- `ClaimRecordKey: 0xd6 | AirdropId | RecipientAddrLen (1 byte) | RecipientAddr -> ProtocolBuffer(ClaimRecord)`
