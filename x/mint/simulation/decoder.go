package simulation

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding mint type.
//func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
//	return func(kvA, kvB kv.Pair) string {
//		switch {
//		case bytes.Equal(kvA.Key, types.MinterKey):
//			return fmt.Sprintf("%v\n%v", minterA, minterB)
//		default:
//			panic(fmt.Sprintf("invalid mint key %X", kvA.Key))
//		}
//	}
//}
