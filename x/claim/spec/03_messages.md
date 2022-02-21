<!-- order: 3 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps `claim` module messages from transactions.

## MsgClaim

```go
// MsgClaim defines a message for claiming claimable amount.
type MsgClaim struct {
	AirdropId     uint64
	Requestor     string	
	ConditionType ConditionType
}
```

