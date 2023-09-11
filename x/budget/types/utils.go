package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/tendermint/tendermint/crypto"
)

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

// DateRangeIncludes returns true if the target date included on the start, end time range.
// End time is exclusive and start time is inclusive.
func DateRangeIncludes(startTime, endTime, targetTime time.Time) bool {
	return endTime.After(targetTime) && !startTime.After(targetTime)
}

// DeriveAddress derives an address with the given address length type, module name, and
// address derivation name. It is used to derive source or destination address.
func DeriveAddress(addressType AddressType, moduleName, name string) sdk.AccAddress {
	switch addressType {
	case AddressType32Bytes:
		return sdk.AccAddress(address.Module(moduleName, []byte(name)))
	case AddressType20Bytes:
		return sdk.AccAddress(crypto.AddressHash([]byte(moduleName + name)))
	default:
		return sdk.AccAddress{}
	}
}
