<!-- order: 3 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps `claim` module messages from transactions.

## MsgClaim

```go
// MsgClaim defines a message for claiming claimable amount.
type MsgClaim struct {
	Requestor  string	
	ActionType ActionType
}

// ActionType defines the type of action that a recipient
// must execute in order to receive their claimable amount.
enum ActionType {
	ACTION_TYPE_UNSPECIFIED
	ACTION_TYPE_DEPOSIT
	ACTION_TYPE_SWAP
	ACTION_TYPE_STAKE
}
```

