<!-- order: 2 -->

# State

### Airdrop

```go

// Airdrop defines airdrop information.
type Airdrop struct {
	AirdropId          uint64    // airdrop_id specifies index of the airdrop
	SourceAddress      string    // source_address defines the bech32-encoded source address
	SourceCoins        sdk.Coins // source_coins specifies the airdrop coins
	TerminationAddress string    // termination_address defines the bech32-encoded termination address
	StartTime          time.Time // start_time specifies the start time of the airdrop
	EndTime            time.Time // end_time specifies the start time of the airdrop
}
```

### Claim Record

```go
// ClaimRecord defines claim record that corresponds to the airdrop.
type ClaimRecord struct {
	AirdropId             uint64    // airdrop_id specifies airdrop id
	Recipient             string    // recipient specifies the bech32-encoded address that is eligible to claim airdrop
	InitialClaimableCoins sdk.Coins // initial_claimable_coins specifies the initial claimable coins
	ClaimableCoins        sdk.Coins // claimable_coins specifies the unclaimed claimable coins
	Actions               []Action  // actions specifies a list of actions
}
```

### Action Type

```go

// Action defines an action type and its claimed status.
type Action struct {
	ActionType ActionType // action_type specifies the action type
	Claimed    bool       // claimed specifies the status of an action
}

// ActionType defines the type of action that a recipient must execute in order to receive a claimable amount.
type ActionType int32

const (
	// ACTION_TYPE_UNSPECIFIED specifies an unknown action type
	ActionTypeUnspecified ActionType = 0
	// ACTION_TYPE_DEPOSIT specifies deposit action type
	ActionTypeDeposit ActionType = 1
	// ACTION_TYPE_SWAP specifies swap action type
	ActionTypeSwap ActionType = 2
	// ACTION_TYPE_FARMING specifies farming (stake) action type
	ActionTypeFarming ActionType = 3
)
```

### Parameters

- ModuleName: `claim`
- RouterKey: `claim`
- StoreKey: `claim`
- QuerierRoute: `claim`


### Stores

Stores are KVStores in the multi-store. The key to find the store is the first parameter in the list.

- `LastAirdropIdKey: 0xd0 -> uint64`
- `StartTimeKey: 0xd5 | AirdropId -> ProtocolBuffer(Timestamp)`
- `EndTimeKey: 0xd6 | AirdropId -> ProtocolBuffer(Timestamp)`
- `AirdropKey: 0xd7 | AirdropId -> ProtocolBuffer(Airdrop)`
- `ClaimRecordKey: 0xd8 | AirdropId -> ProtocolBuffer(ClaimRecord)`
- `ClaimRecordByRecipientKey: 0xd9 | AirdropId | RecipientAddrLen (1 byte) | RecipientAddr -> ProtocolBuffer(ClaimRecord)`
