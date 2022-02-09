package types

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetShareValue multiplies with truncation by receiving int amount and decimal ratio and returns int result.
func GetShareValue(amount sdk.Int, ratio sdk.Dec) sdk.Int {
	return amount.ToDec().MulTruncate(ratio).TruncateInt()
}

type StrIntMap map[string]sdk.Int

// AddOrSet Set when the key not existed on the map or add existed value of the key.
func (m StrIntMap) AddOrSet(key string, value sdk.Int) {
	if _, ok := m[key]; !ok {
		m[key] = value
	} else {
		m[key] = m[key].Add(value)
	}
}

func PP(data interface{}) {
	var p []byte
	p, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}

// MustParseRFC3339 parses string time to time in RFC3339 format.
func MustParseRFC3339(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// DateRangesOverlap returns true if two date ranges overlap each other.
// End time is exclusive and start time is inclusive.
func DateRangesOverlap(startTimeA, endTimeA, startTimeB, endTimeB time.Time) bool {
	return startTimeA.Before(endTimeB) && endTimeA.After(startTimeB)
}
