package types

import (
	"bytes"
)

func Key(comps ...[]byte) []byte {
	return bytes.Join(comps, nil)
}
