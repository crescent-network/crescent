package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
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

// ParseDec is a shortcut for sdk.MustNewDecFromStr.
func ParseDec(s string) sdk.Dec {
	return sdk.MustNewDecFromStr(s)
}

// ParseCoin parses and returns sdk.Coin.
func ParseCoin(s string) sdk.Coin {
	coin, err := sdk.ParseCoinNormalized(s)
	if err != nil {
		panic(err)
	}
	return coin
}

// ParseCoins parses and returns sdk.Coins.
func ParseCoins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return coins
}

// ParseTime parses and returns time.Time in time.RFC3339 format.
func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// DecApproxEqual returns true if a and b are approximately equal,
// which means the diff ratio is equal or less than 0.1%.
func DecApproxEqual(a, b sdk.Dec) bool {
	if b.GT(a) {
		a, b = b, a
	}
	return a.Sub(b).Quo(a).LTE(sdk.NewDecWithPrec(1, 3))
}

// RandomInt returns a random integer in the half-open interval [min, max).
func RandomInt(r *rand.Rand, min, max sdk.Int) sdk.Int {
	return min.Add(sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.Sub(min).BigInt())))
}

// RandomDec returns a random decimal in the half-open interval [min, max).
func RandomDec(r *rand.Rand, min, max sdk.Dec) sdk.Dec {
	return min.Add(sdk.NewDecFromBigIntWithPrec(new(big.Int).Rand(r, max.Sub(min).BigInt()), sdk.Precision))
}
