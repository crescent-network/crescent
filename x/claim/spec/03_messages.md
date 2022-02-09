<!-- order: 3 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps `claim` module messages from transactions.

## MsgClaim

```go
type MsgClaim struct {
	Requestor  string	
	ActionType ActionType
}

enum ActionType {
	ACTION_TYPE_UNSPECIFIED
	ACTION_TYPE_DEPOSIT
	ACTION_TYPE_SWAP
	ACTION_TYPE_STAKE
}
```

