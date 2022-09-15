package types

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
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

// ParseDec is a shortcut for sdk.MustNewDecFromStr.
func ParseDec(s string) sdk.Dec {
	return sdk.MustNewDecFromStr(s)
}

// ParseDecP is like ParseDec, but it returns a pointer to sdk.Dec.
func ParseDecP(s string) *sdk.Dec {
	d := ParseDec(s)
	return &d
}

// ParseCoin parses and returns sdk.Coin.
func ParseCoin(s string) sdk.Coin {
	coin, err := sdk.ParseCoinNormalized(strings.ReplaceAll(s, "_", ""))
	if err != nil {
		panic(err)
	}
	return coin
}

// ParseCoins parses and returns sdk.Coins.
func ParseCoins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(strings.ReplaceAll(s, "_", ""))
	if err != nil {
		panic(err)
	}
	return coins
}

// ParseDecCoins parses and returns sdk.DecCoins.
func ParseDecCoins(s string) sdk.DecCoins {
	coins, err := sdk.ParseDecCoins(strings.ReplaceAll(s, "_", ""))
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

// DecApproxSqrt returns an approximate estimation of x's square root.
func DecApproxSqrt(x sdk.Dec) (r sdk.Dec) {
	var err error
	r, err = x.ApproxSqrt()
	if err != nil {
		panic(err)
	}
	return
}

// RandomInt returns a random integer in the half-open interval [min, max).
func RandomInt(r *rand.Rand, min, max sdk.Int) sdk.Int {
	return min.Add(sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.Sub(min).BigInt())))
}

// RandomDec returns a random decimal in the half-open interval [min, max).
func RandomDec(r *rand.Rand, min, max sdk.Dec) sdk.Dec {
	return min.Add(sdk.NewDecFromBigIntWithPrec(new(big.Int).Rand(r, max.Sub(min).BigInt()), sdk.Precision))
}

// GenAndDeliverTx generates a transactions and delivers it.
func GenAndDeliverTx(txCtx simulation.OperationInput, fees sdk.Coins, gas uint64) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	tx, err := helpers.GenTx(
		txCtx.TxGen,
		[]sdk.Msg{txCtx.Msg},
		fees,
		gas,
		txCtx.Context.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		txCtx.SimAccount.PrivKey,
	)

	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate mock tx"), nil, err
	}

	_, _, err = txCtx.App.Deliver(txCtx.TxGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to deliver tx"), nil, err
	}

	return simtypes.NewOperationMsg(txCtx.Msg, true, "", txCtx.Cdc), nil, nil

}

// GenAndDeliverTxWithFees generates a transaction with given fee and delivers it.
func GenAndDeliverTxWithFees(txCtx simulation.OperationInput, gas uint64, fees sdk.Coins) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	spendable := txCtx.Bankkeeper.SpendableCoins(txCtx.Context, account.GetAddress())

	var err error

	_, hasNeg := spendable.SafeSub(txCtx.CoinsSpentInMsg)
	if hasNeg {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "message doesn't leave room for fees"), nil, err
	}

	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(txCtx, fees, gas)
}

// ShuffleSimAccounts returns randomly shuffled simulation accounts.
func ShuffleSimAccounts(r *rand.Rand, accs []simtypes.Account) []simtypes.Account {
	accs2 := make([]simtypes.Account, len(accs))
	copy(accs2, accs)
	r.Shuffle(len(accs2), func(i, j int) {
		accs2[i], accs2[j] = accs2[j], accs2[i]
	})
	return accs2
}

// TestAddress returns an address for testing purpose.
// TestAddress returns same address when addrNum is same.
func TestAddress(addrNum int) sdk.AccAddress {
	addr := make(sdk.AccAddress, 20)
	binary.PutVarint(addr, int64(addrNum))
	return addr
}

// SafeMath runs f in safe mode, which means that any panics occurred inside f
// gets caught by recover() and if the panic was an overflow, onOverflow is run.
// Otherwise, if the panic was not an overflow, then SafeMath will re-throw
// the panic.
func SafeMath(f, onOverflow func()) {
	defer func() {
		if r := recover(); r != nil {
			if IsOverflow(r) {
				onOverflow()
			} else {
				panic(r)
			}
		}
	}()
	f()
}

// IsOverflow returns true if the panic value can be interpreted as an overflow.
func IsOverflow(r interface{}) bool {
	switch r := r.(type) {
	case string:
		s := strings.ToLower(r)
		return strings.Contains(s, "overflow") || strings.HasSuffix(s, "out of bound")
	}
	return false
}

// LengthPrefixString returns length-prefixed bytes representation of a string.
func LengthPrefixString(s string) []byte {
	bz := []byte(s)
	bzLen := len(bz)
	return append([]byte{byte(bzLen)}, bz...)
}
